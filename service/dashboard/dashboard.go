package dashboard

import (
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/up-zero/my-proxy/logger"
	"github.com/up-zero/my-proxy/models"
	"github.com/up-zero/my-proxy/service/serve"
	"github.com/up-zero/my-proxy/util"
	"go.uber.org/zap"
)

const (
	sampleInterval = 2 * time.Second
	historyLimit   = 30
)

type taskRateSnapshot struct {
	TrafficIn  int64
	TrafficOut int64
}

type dashboardSample struct {
	Timestamp        int64
	InboundRate      float64
	OutboundRate     float64
	TotalTrafficIn   int64
	TotalTrafficOut  int64
	TotalConnections int64
	CPUPercent       float64
	MemoryPercent    float64
	MemoryUsed       uint64
	MemoryTotal      uint64
	GoMemoryAlloc    uint64
	Goroutines       int
	NodeRates        map[string]taskRateSnapshot
}

type collector struct {
	mu              sync.RWMutex
	startTime       time.Time
	lastSampleAt    time.Time
	lastTrafficIn   int64
	lastTrafficOut  int64
	lastTaskTraffic map[string]taskRateSnapshot
	history         []dashboardSample
}

var (
	defaultCollector = &collector{
		startTime:       time.Now(),
		lastTaskTraffic: make(map[string]taskRateSnapshot),
		history:         make([]dashboardSample, 0, historyLimit),
	}
	startOnce sync.Once
)

func Start() {
	startOnce.Do(func() {
		defaultCollector.sampleOnce()
		go defaultCollector.loop()
	})
}

func Overview(c *gin.Context) {
	Start()
	data, err := defaultCollector.buildOverview()
	if err != nil {
		logger.Error("[dashboard] build overview error", zap.Error(err))
		util.ResponseError(c, err)
		return
	}
	util.ResponseOkWithData(c, data)
}

func (c *collector) loop() {
	ticker := time.NewTicker(sampleInterval)
	defer ticker.Stop()

	for range ticker.C {
		c.sampleOnce()
	}
}

func (c *collector) sampleOnce() {
	tasks := serve.ProxyTaskSnapshots()
	now := time.Now()
	currentTrafficIn := int64(0)
	currentTrafficOut := int64(0)
	currentConnections := int64(0)
	nodeRates := make(map[string]taskRateSnapshot, len(tasks))

	for _, task := range tasks {
		currentTrafficIn += task.TrafficIn
		currentTrafficOut += task.TrafficOut
		currentConnections += task.ActiveConnections
	}

	cpuPercent := 0.0
	if values, err := cpu.Percent(0, false); err == nil && len(values) > 0 {
		cpuPercent = values[0]
	}

	memoryUsed := uint64(0)
	memoryTotal := uint64(0)
	memoryPercent := 0.0
	if vm, err := mem.VirtualMemory(); err == nil {
		memoryUsed = vm.Used
		memoryTotal = vm.Total
		memoryPercent = vm.UsedPercent
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	inboundRate := 0.0
	outboundRate := 0.0
	deltaSeconds := now.Sub(c.lastSampleAt).Seconds()
	if !c.lastSampleAt.IsZero() && deltaSeconds > 0 {
		inboundRate = clampRate(float64(currentTrafficIn-c.lastTrafficIn) / deltaSeconds)
		outboundRate = clampRate(float64(currentTrafficOut-c.lastTrafficOut) / deltaSeconds)
	}

	for _, task := range tasks {
		prev := c.lastTaskTraffic[task.Uuid]
		taskInRate := 0.0
		taskOutRate := 0.0
		if !c.lastSampleAt.IsZero() && deltaSeconds > 0 {
			taskInRate = clampRate(float64(task.TrafficIn-prev.TrafficIn) / deltaSeconds)
			taskOutRate = clampRate(float64(task.TrafficOut-prev.TrafficOut) / deltaSeconds)
		}
		nodeRates[task.Uuid] = taskRateSnapshot{
			TrafficIn:  int64(taskInRate),
			TrafficOut: int64(taskOutRate),
		}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.lastSampleAt = now
	c.lastTrafficIn = currentTrafficIn
	c.lastTrafficOut = currentTrafficOut
	c.lastTaskTraffic = make(map[string]taskRateSnapshot, len(tasks))
	for _, task := range tasks {
		c.lastTaskTraffic[task.Uuid] = taskRateSnapshot{
			TrafficIn:  task.TrafficIn,
			TrafficOut: task.TrafficOut,
		}
	}

	c.history = append(c.history, dashboardSample{
		Timestamp:        now.UnixMilli(),
		InboundRate:      inboundRate,
		OutboundRate:     outboundRate,
		TotalTrafficIn:   currentTrafficIn,
		TotalTrafficOut:  currentTrafficOut,
		TotalConnections: currentConnections,
		CPUPercent:       cpuPercent,
		MemoryPercent:    memoryPercent,
		MemoryUsed:       memoryUsed,
		MemoryTotal:      memoryTotal,
		GoMemoryAlloc:    memStats.Alloc,
		Goroutines:       runtime.NumGoroutine(),
		NodeRates:        nodeRates,
	})
	if len(c.history) > historyLimit {
		c.history = append([]dashboardSample(nil), c.history[len(c.history)-historyLimit:]...)
	}
}

func (c *collector) buildOverview() (*OverviewResponse, error) {
	proxyList, err := loadProxyBasics()
	if err != nil {
		return nil, err
	}
	groupMap, err := loadProxyGroupMap()
	if err != nil {
		return nil, err
	}

	taskMap := make(map[string]*serve.TaskSnapshot)
	for _, task := range serve.ProxyTaskSnapshots() {
		taskMap[task.Uuid] = task
	}

	latest, history := c.snapshotHistory()
	nodes := make([]NodeLoadMetric, 0, len(proxyList))
	running := 0
	totalConnections := int64(0)
	totalTrafficIn := int64(0)
	totalTrafficOut := int64(0)

	for _, proxyBasic := range proxyList {
		task := taskMap[proxyBasic.Uuid]
		state := models.ProxyStateStopped
		trafficIn := int64(0)
		trafficOut := int64(0)
		activeConnections := int64(0)
		inRate := 0.0
		outRate := 0.0
		if task != nil {
			state = task.State
			trafficIn = task.TrafficIn
			trafficOut = task.TrafficOut
			activeConnections = task.ActiveConnections
			if latest.NodeRates != nil {
				if rate, ok := latest.NodeRates[proxyBasic.Uuid]; ok {
					inRate = float64(rate.TrafficIn)
					outRate = float64(rate.TrafficOut)
				}
			}
		}
		if state == models.ProxyStateRunning {
			running += 1
		}
		totalConnections += activeConnections
		totalTrafficIn += trafficIn
		totalTrafficOut += trafficOut

		nodes = append(nodes, NodeLoadMetric{
			Uuid:              proxyBasic.Uuid,
			Name:              proxyBasic.Name,
			GroupName:         groupMap[proxyBasic.GroupUuid],
			Type:              proxyBasic.Type,
			State:             state,
			ListenAddress:     proxyBasic.ListenAddress,
			ListenPort:        proxyBasic.ListenPort,
			TargetAddress:     proxyBasic.TargetAddress,
			TargetPort:        proxyBasic.TargetPort,
			ActiveConnections: activeConnections,
			TrafficIn:         trafficIn,
			TrafficOut:        trafficOut,
			InboundRate:       inRate,
			OutboundRate:      outRate,
			LoadScore:         calcLoadScore(activeConnections, inRate, outRate),
		})
	}

	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].LoadScore == nodes[j].LoadScore {
			return nodes[i].Name < nodes[j].Name
		}
		return nodes[i].LoadScore > nodes[j].LoadScore
	})
	if len(nodes) > 6 {
		nodes = nodes[:6]
	}

	resp := &OverviewResponse{
		Summary: SummaryMetrics{
			ProxyTotal:       len(proxyList),
			ProxyRunning:     running,
			ProxyStopped:     len(proxyList) - running,
			TotalConnections: totalConnections,
			TotalTrafficIn:   totalTrafficIn,
			TotalTrafficOut:  totalTrafficOut,
			InboundRate:      latest.InboundRate,
			OutboundRate:     latest.OutboundRate,
			UptimeSeconds:    int64(time.Since(c.startTime).Seconds()),
			UpdatedAt:        latest.Timestamp,
		},
		System: SystemMetrics{
			CPUPercent:       latest.CPUPercent,
			MemoryPercent:    latest.MemoryPercent,
			MemoryUsed:       latest.MemoryUsed,
			MemoryTotal:      latest.MemoryTotal,
			GoMemoryAlloc:    latest.GoMemoryAlloc,
			Goroutines:       latest.Goroutines,
			SampleIntervalMs: sampleInterval.Milliseconds(),
		},
		Charts: DashboardCharts{
			Traffic:     buildTrafficPoints(history),
			Connections: buildConnectionPoints(history),
			System:      buildSystemPoints(history),
		},
		Nodes: nodes,
	}
	return resp, nil
}

func (c *collector) snapshotHistory() (dashboardSample, []dashboardSample) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.history) == 0 {
		return dashboardSample{}, nil
	}

	latest := c.history[len(c.history)-1]
	history := append([]dashboardSample(nil), c.history...)
	return latest, history
}

func buildTrafficPoints(history []dashboardSample) []TrafficPoint {
	points := make([]TrafficPoint, 0, len(history))
	for _, item := range history {
		points = append(points, TrafficPoint{
			Timestamp:    item.Timestamp,
			InboundRate:  item.InboundRate,
			OutboundRate: item.OutboundRate,
		})
	}
	return points
}

func buildConnectionPoints(history []dashboardSample) []ConnectionPoint {
	points := make([]ConnectionPoint, 0, len(history))
	for _, item := range history {
		points = append(points, ConnectionPoint{
			Timestamp:   item.Timestamp,
			Connections: item.TotalConnections,
		})
	}
	return points
}

func buildSystemPoints(history []dashboardSample) []SystemPoint {
	points := make([]SystemPoint, 0, len(history))
	for _, item := range history {
		points = append(points, SystemPoint{
			Timestamp:     item.Timestamp,
			CPUPercent:    item.CPUPercent,
			MemoryPercent: item.MemoryPercent,
		})
	}
	return points
}

func calcLoadScore(connections int64, inboundRate float64, outboundRate float64) float64 {
	return float64(connections)*10 + (inboundRate+outboundRate)/1024
}

func clampRate(rate float64) float64 {
	if rate < 0 {
		return 0
	}
	return rate
}

func loadProxyBasics() ([]*models.ProxyBasic, error) {
	list := make([]*models.ProxyBasic, 0)
	if err := models.DB.Model(new(models.ProxyBasic)).Order("name ASC").Find(&list).Error; err != nil {
		logger.Error("[db] get proxy list error", zap.Error(err))
		return nil, err
	}
	return list, nil
}

func loadProxyGroupMap() (map[string]string, error) {
	list := make([]*models.GroupBasic, 0)
	if err := models.DB.Model(new(models.GroupBasic)).Find(&list).Error; err != nil {
		logger.Error("[db] get group list error", zap.Error(err))
		return nil, err
	}
	groupMap := make(map[string]string, len(list))
	for _, item := range list {
		groupMap[item.Uuid] = item.Name
	}
	return groupMap, nil
}
