package util

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	LanguageHeader     = "language"
	LanguageContextKey = "language"
	LanguageZhCN       = "zh-CN"
	LanguageEnUS       = "en-US"
	DefaultLanguage    = LanguageZhCN
)

type translationPattern struct {
	pattern  *regexp.Regexp
	template string
}

var localizedMessages = map[string]map[string]string{
	LanguageZhCN: {
		MsgOk:                        "成功",
		MsgErr:                       "系统异常",
		MsgErrAuth:                   "登录过期，请重新登录",
		MsgErrParam:                  "参数异常",
		MsgErrDB:                     "数据库异常",
		MsgErrNet:                    "网络异常",
		MsgErrUsernameOrPassword:     "用户名或密码错误",
		MsgErrOldPasswordWrong:       "旧密码错误",
		MsgErrNameExist:              "名称已存在",
		MsgErrDataNotExist:           "暂无数据",
		"name already exists":        "名称已存在",
		"listen port already exists": "监听端口已存在",
		"proxy not exists":           "代理不存在",
		"token is invalid":           "token 无效",
		"task uuid is required":      "task_uuid 不能为空",
		"task is not running":        "当前任务未运行",
		"API route not found":        "接口不存在",
		"cannot delete local node":   "不能删除本地节点",
		"node auth failed":           "子节点认证失败，请检查节点密钥配置",
	},
	LanguageEnUS: {
		MsgOk:                        "Success",
		MsgErr:                       "System error",
		MsgErrAuth:                   "Login has expired. Please sign in again",
		MsgErrParam:                  "Invalid parameters",
		MsgErrDB:                     "Database error",
		MsgErrNet:                    "Network error",
		MsgErrUsernameOrPassword:     "Incorrect username or password",
		MsgErrOldPasswordWrong:       "Incorrect current password",
		MsgErrNameExist:              "Name already exists",
		MsgErrDataNotExist:           "No data available",
		"name already exists":        "Name already exists",
		"listen port already exists": "Listen port already exists",
		"proxy not exists":           "Proxy does not exist",
		"token is invalid":           "Token is invalid",
		"task uuid is required":      "task_uuid is required",
		"task is not running":        "Task is not running",
		"API route not found":        "API route not found",
		"cannot delete local node":   "Cannot delete the local node",
		"node auth failed":           "Child node authentication failed, please verify the node secret key",
	},
}

var localizedPatterns = map[string][]translationPattern{
	LanguageZhCN: {
		{pattern: regexp.MustCompile(`^name is required$`), template: "名称不能为空"},
		{pattern: regexp.MustCompile(`^type is required$`), template: "代理类型不能为空"},
		{pattern: regexp.MustCompile(`^listen_port is required$`), template: "监听端口不能为空"},
		{pattern: regexp.MustCompile(`^target_address is required$`), template: "目标地址不能为空"},
		{pattern: regexp.MustCompile(`^target_port is required$`), template: "目标端口不能为空"},
		{pattern: regexp.MustCompile(`^proxy type\((.+)\) not support$`), template: "代理类型(%s)不支持"},
		{pattern: regexp.MustCompile(`^tag\((.+)\) does not exist$`), template: "标签(%s)不存在"},
		{pattern: regexp.MustCompile(`^name\((.+)\) already exists$`), template: "名称(%s)已存在"},
		{pattern: regexp.MustCompile(`^listen_port\((.+)\) already exists$`), template: "监听端口(%s)已存在"},
		{pattern: regexp.MustCompile(`^unsupported socks version\((.+)\)$`), template: "不支持的 socks 版本(%s)"},
		{pattern: regexp.MustCompile(`^socks5 auth method not supported$`), template: "不支持当前 socks5 认证方式"},
		{pattern: regexp.MustCompile(`^socks5 command\((.+)\) not supported$`), template: "不支持的 socks5 命令(%s)"},
		{pattern: regexp.MustCompile(`^socks5 atyp\((.+)\) not supported$`), template: "不支持的 socks5 地址类型(%s)"},
		{pattern: regexp.MustCompile(`^task\((.*)\) not found$`), template: "任务(%s)不存在"},
		{pattern: regexp.MustCompile(`^task\((.*)\) is not running$`), template: "任务(%s)未运行"},
		{pattern: regexp.MustCompile(`^ResponseWriter does not implement Hijacker: (.+)$`), template: "ResponseWriter 未实现 Hijacker: %s"},
	},
	LanguageEnUS: {
		{pattern: regexp.MustCompile(`^name is required$`), template: "Name is required"},
		{pattern: regexp.MustCompile(`^type is required$`), template: "Proxy type is required"},
		{pattern: regexp.MustCompile(`^listen_port is required$`), template: "Listen port is required"},
		{pattern: regexp.MustCompile(`^target_address is required$`), template: "Target address is required"},
		{pattern: regexp.MustCompile(`^target_port is required$`), template: "Target port is required"},
		{pattern: regexp.MustCompile(`^proxy type\((.+)\) not support$`), template: "Proxy type (%s) is not supported"},
		{pattern: regexp.MustCompile(`^tag\((.+)\) does not exist$`), template: "Tag (%s) does not exist"},
		{pattern: regexp.MustCompile(`^name\((.+)\) already exists$`), template: "Name (%s) already exists"},
		{pattern: regexp.MustCompile(`^listen_port\((.+)\) already exists$`), template: "Listen port (%s) already exists"},
		{pattern: regexp.MustCompile(`^unsupported socks version\((.+)\)$`), template: "Unsupported socks version (%s)"},
		{pattern: regexp.MustCompile(`^socks5 auth method not supported$`), template: "SOCKS5 auth method is not supported"},
		{pattern: regexp.MustCompile(`^socks5 command\((.+)\) not supported$`), template: "SOCKS5 command (%s) is not supported"},
		{pattern: regexp.MustCompile(`^socks5 atyp\((.+)\) not supported$`), template: "SOCKS5 atyp (%s) is not supported"},
		{pattern: regexp.MustCompile(`^task\((.*)\) not found$`), template: "Task (%s) was not found"},
		{pattern: regexp.MustCompile(`^task\((.*)\) is not running$`), template: "Task (%s) is not running"},
		{pattern: regexp.MustCompile(`^ResponseWriter does not implement Hijacker: (.+)$`), template: "ResponseWriter does not implement Hijacker: %s"},
	},
}

func NormalizeLanguage(language string) string {
	value := strings.TrimSpace(language)
	if value == "" {
		return DefaultLanguage
	}

	lowerValue := strings.ToLower(value)
	switch {
	case strings.HasPrefix(lowerValue, "zh"):
		return LanguageZhCN
	case strings.HasPrefix(lowerValue, "en"):
		return LanguageEnUS
	default:
		return value
	}
}

func DetectRequestLanguage(c *gin.Context) string {
	if c == nil {
		return DefaultLanguage
	}
	if language := c.GetHeader(LanguageHeader); strings.TrimSpace(language) != "" {
		return NormalizeLanguage(language)
	}
	if language := c.Query(LanguageHeader); strings.TrimSpace(language) != "" {
		return NormalizeLanguage(language)
	}
	return DefaultLanguage
}

func SetRequestLanguage(c *gin.Context, language string) string {
	resolved := NormalizeLanguage(language)
	if c != nil {
		c.Set(LanguageContextKey, resolved)
		c.Header("Content-Language", resolved)
	}
	return resolved
}

func GetRequestLanguage(c *gin.Context) string {
	if c == nil {
		return DefaultLanguage
	}
	if value, ok := c.Get(LanguageContextKey); ok {
		if language, ok := value.(string); ok && strings.TrimSpace(language) != "" {
			return NormalizeLanguage(language)
		}
	}
	return DetectRequestLanguage(c)
}

func TranslateMessage(language string, message string) string {
	if strings.TrimSpace(message) == "" {
		return message
	}

	resolvedLanguage := NormalizeLanguage(language)
	if messageMap, ok := localizedMessages[resolvedLanguage]; ok {
		if translated, exists := messageMap[message]; exists {
			return translated
		}
	}

	for _, item := range localizedPatterns[resolvedLanguage] {
		matched := item.pattern.FindStringSubmatch(message)
		if len(matched) == 0 {
			continue
		}
		args := make([]any, 0, len(matched)-1)
		for _, value := range matched[1:] {
			args = append(args, value)
		}
		return fmt.Sprintf(item.template, args...)
	}

	return message
}

func LocalizeMessage(c *gin.Context, message string) string {
	return TranslateMessage(GetRequestLanguage(c), message)
}
