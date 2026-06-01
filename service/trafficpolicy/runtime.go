package trafficpolicy

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/up-zero/my-proxy/logger"
	"github.com/up-zero/my-proxy/models"
	"github.com/up-zero/my-proxy/service/alert"
	"go.uber.org/zap"
)

const (
	policyCacheTTL   = 5 * time.Second
	alertCooldown    = 5 * time.Minute
	fallbackSlowdown = 200 * time.Millisecond
	runtimeQueueSize = 1024
)

// runtimePolicy 是转发热路径中使用的策略对象。
//
// 设计原则：
// 1. policy 基础配置只读，随策略快照整体替换，不在转发路径修改；
// 2. quotaBytes/windowBytes/lastAlertUnix 使用 atomic，避免高并发下计数错乱；
// 3. 流量使用量仅维护在内存缓存中，RecordTraffic 不执行 DB 操作；
// 4. 告警写入通过非阻塞 channel 投递到异步 worker。
type runtimePolicy struct {
	policy        *models.TrafficPolicy
	rateBytes     int64
	quotaLimit    int64
	maxConn       int64
	actions       []string
	quotaBytes    atomic.Int64
	windowSec     atomic.Int64
	windowBytes   atomic.Int64
	lastAlertUnix atomic.Int64
}

// runtimeCache 是不可变策略快照。
// RecordTraffic 只通过 atomic.Value 读取快照，不加全局锁。
type runtimeCache struct {
	policies    []*runtimePolicy
	proxyTagMap map[string]map[string]struct{}
	loadedAt    int64
}

type alertEvent struct {
	policyUuid string
	level      string
	title      string
	content    string
}

var (
	cacheValue atomic.Value // stores *runtimeCache
	reloadMu   sync.Mutex
	workerOnce sync.Once
	alertCh    = make(chan alertEvent, runtimeQueueSize)
)

type TrafficEvent struct {
	ProxyUuid         string
	Direction         string
	Bytes             int64
	ActiveConnections int64
}

func parseBytesInt(value string) int64 {
	bytes, ok := parseQuotaBytes(value)
	if !ok || bytes <= 0 || math.IsNaN(bytes) || math.IsInf(bytes, 0) {
		return 0
	}
	return int64(bytes)
}

func parseInt64(value string) int64 {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	n, err := strconv.ParseInt(value, 10, 64)
	if err != nil || n < 0 {
		return 0
	}
	return n
}

func formatBytes(bytes int64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	value := float64(bytes)
	idx := 0
	for value >= 1024 && idx < len(units)-1 {
		value /= 1024
		idx++
	}
	if idx == 0 {
		return fmt.Sprintf("%d%s", bytes, units[idx])
	}
	return fmt.Sprintf("%.2f%s", value, units[idx])
}

func startWorkers() {
	workerOnce.Do(func() {
		go func() {
			for event := range alertCh {
				if err := alert.CreateRecord(alert.SourceTrafficPolicy, event.policyUuid, event.level, event.title, event.content); err != nil {
					logger.Error("[traffic-policy] create alert error", zap.Error(err), zap.String("uuid", event.policyUuid))
				}
			}
		}()
	})
}

// StartRuntime 在服务启动时预加载限速配额策略并启动异步 worker。
// 该函数不在代理转发热路径上，允许同步读取数据库。
func StartRuntime() {
	startWorkers()
	RefreshRuntime()
}

// RefreshRuntime 同步刷新运行时快照。
// 用于服务启动、策略 CRUD、状态切换等低频路径；不在每次流量转发中调用。
func RefreshRuntime() {
	startWorkers()
	reloadMu.Lock()
	defer reloadMu.Unlock()
	cacheValue.Store(loadRuntimeCache())
}

func maybeRefreshRuntimeAsync(cache *runtimeCache) {
	if cache != nil && time.Since(time.UnixMilli(cache.loadedAt)) < policyCacheTTL {
		return
	}
	if !reloadMu.TryLock() {
		return
	}
	go func() {
		defer reloadMu.Unlock()
		cacheValue.Store(loadRuntimeCache())
	}()
}

func currentCache() *runtimeCache {
	value := cacheValue.Load()
	if value == nil {
		return nil
	}
	cache, _ := value.(*runtimeCache)
	return cache
}

func loadRuntimeCache() *runtimeCache {
	list := make([]*models.TrafficPolicy, 0)
	if err := models.DB.Model(new(models.TrafficPolicy)).Where("status = ?", models.TrafficPolicyStatusEnabled).Find(&list).Error; err != nil {
		logger.Error("[traffic-policy] load policies error", zap.Error(err))
		return currentCacheOrEmpty()
	}

	oldByUuid := map[string]*runtimePolicy{}
	if old := currentCache(); old != nil {
		for _, item := range old.policies {
			oldByUuid[item.policy.Uuid] = item
		}
	}

	proxyTagMap := loadProxyTagSnapshot(list)
	nextPolicies := make([]*runtimePolicy, 0, len(list))
	nowSec := time.Now().Unix()
	for _, item := range list {
		rp := &runtimePolicy{
			policy:     item,
			rateBytes:  parseBytesInt(item.OutboundLimit),
			quotaLimit: parseBytesInt(item.PeriodQuota),
			maxConn:    parseInt64(item.MaxConnections),
			actions:    splitActions(item.OverLimitAction),
		}
		if old := oldByUuid[item.Uuid]; old != nil {
			rp.quotaBytes.Store(old.quotaBytes.Load())
			rp.windowSec.Store(old.windowSec.Load())
			rp.windowBytes.Store(old.windowBytes.Load())
			rp.lastAlertUnix.Store(old.lastAlertUnix.Load())
		} else {
			rp.quotaBytes.Store(parseBytesInt(item.QuotaUsed))
			rp.windowSec.Store(nowSec)
		}
		nextPolicies = append(nextPolicies, rp)
	}
	return &runtimeCache{
		policies:    nextPolicies,
		proxyTagMap: proxyTagMap,
		loadedAt:    time.Now().UnixMilli(),
	}
}

func currentCacheOrEmpty() *runtimeCache {
	if cache := currentCache(); cache != nil {
		return cache
	}
	return &runtimeCache{
		policies:    []*runtimePolicy{},
		proxyTagMap: map[string]map[string]struct{}{},
		loadedAt:    time.Now().UnixMilli(),
	}
}

func loadProxyTagSnapshot(list []*models.TrafficPolicy) map[string]map[string]struct{} {
	hasTagScope := false
	for _, item := range list {
		if item.ScopeType == ScopeTag {
			hasTagScope = true
			break
		}
	}
	if !hasTagScope {
		return map[string]map[string]struct{}{}
	}
	fullTagMap, err := models.LoadProxyTagListMap(nil)
	if err != nil {
		logger.Error("[traffic-policy] load all proxy tags error", zap.Error(err))
		return map[string]map[string]struct{}{}
	}
	res := make(map[string]map[string]struct{}, len(fullTagMap))
	for proxyUuid, tags := range fullTagMap {
		res[proxyUuid] = map[string]struct{}{}
		for _, tag := range tags {
			res[proxyUuid][tag.Uuid] = struct{}{}
		}
	}
	return res
}

func policyMatches(cache *runtimeCache, proxyUuid string, policy *models.TrafficPolicy) bool {
	switch policy.ScopeType {
	case ScopeAll:
		return true
	case ScopeProxy:
		return policy.ScopeValue == proxyUuid
	case ScopeTag:
		_, ok := cache.proxyTagMap[proxyUuid][policy.ScopeValue]
		return ok
	default:
		return false
	}
}

func containsAction(actions []string, target string) bool {
	for _, action := range actions {
		if action == target {
			return true
		}
	}
	return false
}

func maybeEnqueueAlert(policy *runtimePolicy, reason string, now time.Time) {
	if !containsAction(policy.actions, models.OverLimitActionAlert) {
		return
	}
	nowUnix := now.Unix()
	for {
		last := policy.lastAlertUnix.Load()
		if last > 0 && nowUnix-last < int64(alertCooldown/time.Second) {
			return
		}
		if policy.lastAlertUnix.CompareAndSwap(last, nowUnix) {
			break
		}
	}
	select {
	case alertCh <- alertEvent{
		policyUuid: policy.policy.Uuid,
		level:      models.AlertLevelWarning,
		title:      "限速配额规则触发告警",
		content:    fmt.Sprintf("策略【%s】已触发：%s", policy.policy.Name, reason),
	}:
	default:
		logger.Warn("[traffic-policy] alert queue full", zap.String("uuid", policy.policy.Uuid))
	}
}

func recordRateWindow(policy *runtimePolicy, bytes int64, now time.Time) int64 {
	nowSec := now.Unix()
	oldSec := policy.windowSec.Load()
	if oldSec != nowSec && policy.windowSec.CompareAndSwap(oldSec, nowSec) {
		policy.windowBytes.Store(0)
	}
	return policy.windowBytes.Add(bytes)
}

// RecordTraffic 记录真实代理出站流量，并返回当前流量包需要额外等待的时间。
//
// 高并发设计说明：
// 1. 热路径不使用全局锁，策略配置通过 atomic.Value 读取不可变快照；
// 2. 每条策略的配额、速率窗口、告警冷却时间均使用 atomic 计数；
// 3. 流量使用量现阶段只放在内存缓存中，不在代理热路径落库；
// 4. 告警写入通过非阻塞 channel 投递到异步 worker；
// 5. 没有匹配规则或未触发阈值时直接返回 0，不影响正常代理转发。
func RecordTraffic(event TrafficEvent) time.Duration {
	if event.ProxyUuid == "" || event.Bytes <= 0 || strings.ToUpper(event.Direction) != "OUT" {
		return 0
	}
	startWorkers()
	cache := currentCache()
	maybeRefreshRuntimeAsync(cache)
	if cache == nil || len(cache.policies) == 0 {
		return 0
	}

	now := time.Now()
	var maxDelay time.Duration
	for _, policy := range cache.policies {
		if !policyMatches(cache, event.ProxyUuid, policy.policy) {
			continue
		}
		usedBytes := policy.quotaBytes.Add(event.Bytes)
		triggered := false
		reason := ""

		if policy.maxConn > 0 && event.ActiveConnections > policy.maxConn {
			triggered = true
			reason = fmt.Sprintf("当前连接数 %d 超过并发连接上限 %d", event.ActiveConnections, policy.maxConn)
		}
		if policy.quotaLimit > 0 && usedBytes >= policy.quotaLimit {
			triggered = true
			reason = fmt.Sprintf("配额已用 %s，达到或超过周期配额 %s", formatBytes(usedBytes), policy.policy.PeriodQuota)
		}
		if policy.rateBytes > 0 {
			windowBytes := recordRateWindow(policy, event.Bytes, now)
			if windowBytes > policy.rateBytes {
				triggered = true
				reason = fmt.Sprintf("当前出站速率超过上限 %s", policy.policy.OutboundLimit)
				if containsAction(policy.actions, models.OverLimitActionSlowdown) {
					delay := time.Duration(float64(event.Bytes) / float64(policy.rateBytes) * float64(time.Second))
					if delay > maxDelay {
						maxDelay = delay
					}
				}
			}
		}
		if triggered {
			maybeEnqueueAlert(policy, reason, now)
			if containsAction(policy.actions, models.OverLimitActionSlowdown) && maxDelay == 0 {
				maxDelay = fallbackSlowdown
			}
		}
	}
	return maxDelay
}

// RuntimeQuotaUsedMap 返回当前运行时缓存中的配额使用量快照。
func RuntimeQuotaUsedMap() map[string]string {
	cache := currentCache()
	if cache == nil || len(cache.policies) == 0 {
		return map[string]string{}
	}
	res := make(map[string]string, len(cache.policies))
	for _, policy := range cache.policies {
		res[policy.policy.Uuid] = formatBytes(policy.quotaBytes.Load())
	}
	return res
}
