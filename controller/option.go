package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/console_setting"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
	"github.com/QuantumNous/new-api/setting/system_setting"

	"github.com/gin-gonic/gin"
)

func GetOptions(c *gin.Context) {
	var options []*model.Option
	common.OptionMapRWMutex.Lock()
	for k, v := range common.OptionMap {
		if strings.HasSuffix(k, "Token") ||
			strings.HasSuffix(k, "Secret") ||
			strings.HasSuffix(k, "Key") ||
			strings.HasSuffix(k, "secret") ||
			strings.HasSuffix(k, "api_key") {
			continue
		}
		options = append(options, &model.Option{
			Key:   k,
			Value: common.Interface2String(v),
		})
	}
	common.OptionMapRWMutex.Unlock()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    options,
	})
	return
}

type OptionUpdateRequest struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

func UpdateOption(c *gin.Context) {
	var option OptionUpdateRequest
	err := json.NewDecoder(c.Request.Body).Decode(&option)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}
	switch option.Value.(type) {
	case bool:
		option.Value = common.Interface2String(option.Value.(bool))
	case float64:
		option.Value = common.Interface2String(option.Value.(float64))
	case int:
		option.Value = common.Interface2String(option.Value.(int))
	default:
		option.Value = fmt.Sprintf("%v", option.Value)
	}
	switch option.Key {
	case "GitHubOAuthEnabled":
		if option.Value == "true" && common.GitHubClientId == "" {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无法启用 GitHub OAuth，请先填入 GitHub Client Id 以及 GitHub Client Secret！",
			})
			return
		}
	case "discord.enabled":
		if option.Value == "true" && system_setting.GetDiscordSettings().ClientId == "" {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无法启用 Discord OAuth，请先填入 Discord Client Id 以及 Discord Client Secret！",
			})
			return
		}
	case "oidc.enabled":
		if option.Value == "true" && system_setting.GetOIDCSettings().ClientId == "" {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无法启用 OIDC 登录，请先填入 OIDC Client Id 以及 OIDC Client Secret！",
			})
			return
		}
	case "LinuxDOOAuthEnabled":
		if option.Value == "true" && common.LinuxDOClientId == "" {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无法启用 LinuxDO OAuth，请先填入 LinuxDO Client Id 以及 LinuxDO Client Secret！",
			})
			return
		}
	case "EmailDomainRestrictionEnabled":
		if option.Value == "true" && len(common.EmailDomainWhitelist) == 0 {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无法启用邮箱域名限制，请先填入限制的邮箱域名！",
			})
			return
		}
	case "WeChatAuthEnabled":
		if option.Value == "true" && common.WeChatServerAddress == "" {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无法启用微信登录，请先填入微信登录相关配置信息！",
			})
			return
		}
	case "TurnstileCheckEnabled":
		if option.Value == "true" && common.TurnstileSiteKey == "" {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无法启用 Turnstile 校验，请先填入 Turnstile 校验相关配置信息！",
			})

			return
		}
	case "TelegramOAuthEnabled":
		if option.Value == "true" && common.TelegramBotToken == "" {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无法启用 Telegram OAuth，请先填入 Telegram Bot Token！",
			})
			return
		}
	case "GroupRatio":
		err = ratio_setting.CheckGroupRatio(option.Value.(string))
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	case "ImageRatio":
		err = ratio_setting.UpdateImageRatioByJSONString(option.Value.(string))
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "图片倍率设置失败: " + err.Error(),
			})
			return
		}
	case "AudioRatio":
		err = ratio_setting.UpdateAudioRatioByJSONString(option.Value.(string))
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "音频倍率设置失败: " + err.Error(),
			})
			return
		}
	case "AudioCompletionRatio":
		err = ratio_setting.UpdateAudioCompletionRatioByJSONString(option.Value.(string))
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "音频补全倍率设置失败: " + err.Error(),
			})
			return
		}
	case "ModelRequestRateLimitGroup":
		err = setting.CheckModelRequestRateLimitGroup(option.Value.(string))
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	case "AutomaticDisableStatusCodes":
		_, err = operation_setting.ParseHTTPStatusCodeRanges(option.Value.(string))
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	case "console_setting.api_info":
		err = console_setting.ValidateConsoleSettings(option.Value.(string), "ApiInfo")
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	case "console_setting.announcements":
		err = console_setting.ValidateConsoleSettings(option.Value.(string), "Announcements")
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	case "console_setting.faq":
		err = console_setting.ValidateConsoleSettings(option.Value.(string), "FAQ")
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	case "console_setting.uptime_kuma_groups":
		err = console_setting.ValidateConsoleSettings(option.Value.(string), "UptimeKumaGroups")
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	}
	err = model.UpdateOption(option.Key, option.Value.(string))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

// GetEnvVariables returns the list of environment variables that are currently set
func GetEnvVariables(c *gin.Context) {
		// Only show environment variables in self-hosted mode
		if !common.SelfHostedMode {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "环境变量信息仅在自用模式下可见",
			})
			return
		}
	
	// Define all environment variables that the application supports
	envVars := []map[string]interface{}{
		// Basic Settings
		{"key": "PORT", "value": common.GetEnvOrDefaultString("PORT", ""), "description": "端口号", "category": "基础配置"},
		{"key": "FRONTEND_BASE_URL", "value": common.GetEnvOrDefaultString("FRONTEND_BASE_URL", ""), "description": "前端基础URL", "category": "基础配置"},
		{"key": "NODE_TYPE", "value": common.GetEnvOrDefaultString("NODE_TYPE", ""), "description": "节点类型", "category": "基础配置"},
		
		// Debug Settings
		{"key": "ENABLE_PPROF", "value": common.GetEnvOrDefaultString("ENABLE_PPROF", ""), "description": "启用pprof", "category": "调试配置"},
		{"key": "DEBUG", "value": common.GetEnvOrDefaultString("DEBUG", ""), "description": "启用调试模式", "category": "调试配置"},
		{"key": "LOG_CONTENTS", "value": common.GetEnvOrDefaultString("LOG_CONTENTS", ""), "description": "记录请求/响应内容到日志", "category": "调试配置"},
		{"key": "PYROSCOPE_URL", "value": common.GetEnvOrDefaultString("PYROSCOPE_URL", ""), "description": "Pyroscope URL", "category": "调试配置"},
		{"key": "PYROSCOPE_APP_NAME", "value": common.GetEnvOrDefaultString("PYROSCOPE_APP_NAME", ""), "description": "Pyroscope 应用名称", "category": "调试配置"},
		{"key": "HOSTNAME", "value": common.GetEnvOrDefaultString("HOSTNAME", ""), "description": "主机名", "category": "调试配置"},
		
		// Database Settings
		{"key": "SQL_DSN", "value": maskSensitiveValue(common.GetEnvOrDefaultString("SQL_DSN", "")), "description": "数据库连接字符串", "category": "数据库配置"},
		{"key": "LOG_SQL_DSN", "value": maskSensitiveValue(common.GetEnvOrDefaultString("LOG_SQL_DSN", "")), "description": "日志数据库连接字符串", "category": "数据库配置"},
		{"key": "SQLITE_PATH", "value": common.GetEnvOrDefaultString("SQLITE_PATH", ""), "description": "SQLite数据库路径", "category": "数据库配置"},
		{"key": "SQL_MAX_IDLE_CONNS", "value": common.GetEnvOrDefaultString("SQL_MAX_IDLE_CONNS", ""), "description": "数据库最大空闲连接数", "category": "数据库配置"},
		{"key": "SQL_MAX_OPEN_CONNS", "value": common.GetEnvOrDefaultString("SQL_MAX_OPEN_CONNS", ""), "description": "数据库最大打开连接数", "category": "数据库配置"},
		{"key": "SQL_MAX_LIFETIME", "value": common.GetEnvOrDefaultString("SQL_MAX_LIFETIME", ""), "description": "数据库连接最大生命周期", "category": "数据库配置"},
		
		// Cache Settings
		{"key": "REDIS_CONN_STRING", "value": maskSensitiveValue(common.GetEnvOrDefaultString("REDIS_CONN_STRING", "")), "description": "Redis连接字符串", "category": "缓存配置"},
		{"key": "SYNC_FREQUENCY", "value": common.GetEnvOrDefaultString("SYNC_FREQUENCY", ""), "description": "同步频率(秒)", "category": "缓存配置"},
		{"key": "MEMORY_CACHE_ENABLED", "value": common.GetEnvOrDefaultString("MEMORY_CACHE_ENABLED", ""), "description": "内存缓存启用", "category": "缓存配置"},
		{"key": "CHANNEL_UPDATE_FREQUENCY", "value": common.GetEnvOrDefaultString("CHANNEL_UPDATE_FREQUENCY", ""), "description": "渠道更新频率(秒)", "category": "缓存配置"},
		{"key": "BATCH_UPDATE_ENABLED", "value": common.GetEnvOrDefaultString("BATCH_UPDATE_ENABLED", ""), "description": "批量更新启用", "category": "缓存配置"},
		{"key": "BATCH_UPDATE_INTERVAL", "value": common.GetEnvOrDefaultString("BATCH_UPDATE_INTERVAL", ""), "description": "批量更新间隔(秒)", "category": "缓存配置"},
		
		// Task Settings
		{"key": "UPDATE_TASK", "value": common.GetEnvOrDefaultString("UPDATE_TASK", ""), "description": "更新任务启用", "category": "任务配置"},
		
		// Timeout Settings
		{"key": "RELAY_TIMEOUT", "value": common.GetEnvOrDefaultString("RELAY_TIMEOUT", ""), "description": "所有请求超时时间(秒)", "category": "超时配置"},
		{"key": "STREAMING_TIMEOUT", "value": common.GetEnvOrDefaultString("STREAMING_TIMEOUT", ""), "description": "流模式无响应超时时间(秒)", "category": "超时配置"},
		
		// Other Settings
		{"key": "SESSION_SECRET", "value": maskSensitiveValue(common.GetEnvOrDefaultString("SESSION_SECRET", "")), "description": "会话密钥", "category": "其他配置"},
		{"key": "GENERATE_DEFAULT_TOKEN", "value": common.GetEnvOrDefaultString("GENERATE_DEFAULT_TOKEN", ""), "description": "生成默认token", "category": "其他配置"},
		{"key": "COHERE_SAFETY_SETTING", "value": common.GetEnvOrDefaultString("COHERE_SAFETY_SETTING", ""), "description": "Cohere 安全设置", "category": "其他配置"},
		{"key": "GET_MEDIA_TOKEN", "value": common.GetEnvOrDefaultString("GET_MEDIA_TOKEN", ""), "description": "是否统计图片token", "category": "其他配置"},
		{"key": "GET_MEDIA_TOKEN_NOT_STREAM", "value": common.GetEnvOrDefaultString("GET_MEDIA_TOKEN_NOT_STREAM", ""), "description": "非流情况下是否统计图片token", "category": "其他配置"},
		{"key": "DIFY_DEBUG", "value": common.GetEnvOrDefaultString("DIFY_DEBUG", ""), "description": "Dify 渠道是否输出工作流信息", "category": "其他配置"},
		{"key": "GEMINI_VISION_MAX_IMAGE_NUM", "value": common.GetEnvOrDefaultString("GEMINI_VISION_MAX_IMAGE_NUM", ""), "description": "Gemini 识别图片最大数量", "category": "其他配置"},
		{"key": "LINUX_DO_TOKEN_ENDPOINT", "value": common.GetEnvOrDefaultString("LINUX_DO_TOKEN_ENDPOINT", ""), "description": "LinuxDo Token端点", "category": "LinuxDo配置"},
		{"key": "LINUX_DO_USER_ENDPOINT", "value": common.GetEnvOrDefaultString("LINUX_DO_USER_ENDPOINT", ""), "description": "LinuxDo 用户端点", "category": "LinuxDo配置"},
		
		// Sync Settings
		{"key": "SYNC_UPSTREAM_BASE", "value": common.GetEnvOrDefaultString("SYNC_UPSTREAM_BASE", ""), "description": "同步上游基础URL", "category": "同步配置"},
		{"key": "SYNC_HTTP_TIMEOUT_SECONDS", "value": common.GetEnvOrDefaultString("SYNC_HTTP_TIMEOUT_SECONDS", ""), "description": "同步HTTP超时(秒)", "category": "同步配置"},
		{"key": "SYNC_HTTP_RETRY", "value": common.GetEnvOrDefaultString("SYNC_HTTP_RETRY", ""), "description": "同步HTTP重试次数", "category": "同步配置"},
		{"key": "SYNC_HTTP_MAX_MB", "value": common.GetEnvOrDefaultString("SYNC_HTTP_MAX_MB", ""), "description": "同步HTTP最大MB", "category": "同步配置"},
	}
	
	// Filter out empty values and keep only those that are set
	var enabledVars []map[string]interface{}
	for _, env := range envVars {
		if value, ok := env["value"].(string); ok && value != "" {
			enabledVars = append(enabledVars, env)
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    enabledVars,
	})
}

// maskSensitiveValue masks sensitive values like passwords and tokens
func maskSensitiveValue(value string) string {
	if value == "" {
		return ""
	}
	if len(value) <= 8 {
		return "******"
	}
	return value[:4] + "******" + value[len(value)-4:]
}

// GetDatabaseInfo returns the database configuration information
func GetDatabaseInfo(c *gin.Context) {
		// Only show database info in self-hosted mode
		if !common.SelfHostedMode {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "数据库配置信息仅在自用模式下可见",
			})
			return
		}
	
	dbInfo := map[string]interface{}{
		"type": "sqlite", // default
		"host": "",
		"port": 0,
		"database": "",
		"username": "",
		"path": "",
		"fullPath": "",
	}

	// Determine database type
	sqlDsn := os.Getenv("SQL_DSN")
	if sqlDsn != "" {
		if strings.HasPrefix(sqlDsn, "postgres://") || strings.HasPrefix(sqlDsn, "postgresql://") {
			dbInfo["type"] = "postgresql"
			// Parse PostgreSQL DSN
			// Format: postgresql://user:password@host:port/dbname
			parsePostgresDSN(sqlDsn, dbInfo)
		} else if strings.HasPrefix(sqlDsn, "mysql://") || strings.Contains(sqlDsn, "@tcp(") {
			dbInfo["type"] = "mysql"
			// Parse MySQL DSN
			// Format: user:password@tcp(host:port)/dbname
			parseMySQLDSN(sqlDsn, dbInfo)
		}
	} else {
		// Using SQLite
		dbInfo["type"] = "sqlite"
		sqlitePath := common.GetEnvOrDefaultString("SQLITE_PATH", "one-api.db?_busy_timeout=30000")
		// Extract the actual path without the query parameters
		actualPath := strings.Split(sqlitePath, "?")[0]
		dbInfo["path"] = actualPath
		
		// Convert to absolute path
		absPath, err := filepath.Abs(actualPath)
		if err == nil {
			dbInfo["fullPath"] = absPath
		} else {
			dbInfo["fullPath"] = actualPath
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": dbInfo,
	})
}

// parsePostgresDSN parses PostgreSQL connection string
func parsePostgresDSN(dsn string, dbInfo map[string]interface{}) {
	// Format: postgresql://user:password@host:port/dbname
	// Remove the scheme
	dsn = strings.TrimPrefix(dsn, "postgresql://")
	dsn = strings.TrimPrefix(dsn, "postgres://")
	
	// Split user info and host info
	var userInfo, hostInfo string
	if idx := strings.LastIndex(dsn, "@"); idx != -1 {
		userInfo = dsn[:idx]
		hostInfo = dsn[idx+1:]
	} else {
		hostInfo = dsn
	}
	
	// Parse user info
	if userInfo != "" {
		if idx := strings.Index(userInfo, ":"); idx != -1 {
			dbInfo["username"] = userInfo[:idx]
		} else {
			dbInfo["username"] = userInfo
		}
	}
	
	// Parse host info
	if idx := strings.Index(hostInfo, "/"); idx != -1 {
		hostPart := hostInfo[:idx]
		dbName := hostInfo[idx+1:]
		dbInfo["database"] = dbName
		
		// Parse host and port
		if idx := strings.Index(hostPart, ":"); idx != -1 {
			dbInfo["host"] = hostPart[:idx]
			port, _ := strconv.Atoi(hostPart[idx+1:])
			dbInfo["port"] = port
		} else {
			dbInfo["host"] = hostPart
			dbInfo["port"] = 5432 // Default PostgreSQL port
		}
	}
}

// parseMySQLDSN parses MySQL connection string
func parseMySQLDSN(dsn string, dbInfo map[string]interface{}) {
	// Format: user:password@tcp(host:port)/dbname?parseTime=true
	var userInfo, hostInfo string
	
	if idx := strings.Index(dsn, "@"); idx != -1 {
		userInfo = dsn[:idx]
		hostInfo = dsn[idx+1:]
	} else {
		hostInfo = dsn
	}
	
	// Parse user info
	if userInfo != "" {
		if idx := strings.Index(userInfo, ":"); idx != -1 {
			dbInfo["username"] = userInfo[:idx]
		} else {
			dbInfo["username"] = userInfo
		}
	}
	
	// Parse host info - extract from tcp(...) format
	if strings.HasPrefix(hostInfo, "tcp(") {
		// Format: tcp(host:port)/dbname
		if idx := strings.Index(hostInfo, ")"); idx != -1 {
			tcpPart := hostInfo[4:idx] // Remove "tcp(" and ")"
			dbNamePart := hostInfo[idx+2:] // Skip ")/"
			
			// Split database name from query parameters
			if idx := strings.Index(dbNamePart, "?"); idx != -1 {
				dbInfo["database"] = dbNamePart[:idx]
			} else {
				dbInfo["database"] = dbNamePart
			}
			
			// Parse host and port
			if idx := strings.Index(tcpPart, ":"); idx != -1 {
				dbInfo["host"] = tcpPart[:idx]
				port, _ := strconv.Atoi(tcpPart[idx+1:])
				dbInfo["port"] = port
			} else {
				dbInfo["host"] = tcpPart
				dbInfo["port"] = 3306 // Default MySQL port
			}
		}
	}
}
