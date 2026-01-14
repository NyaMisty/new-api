package model

import (
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"

	"github.com/gin-gonic/gin"
)

// LogContent 用于存储详细的请求和响应内容
type LogContent struct {
	LogId        int    `json:"log_id" gorm:"primaryKey;index"`
	RequestBody  string `json:"request_body" gorm:"type:text"`
	ResponseBody string `json:"response_body" gorm:"type:text"`
}

func (LogContent) TableName() string {
	return "log_contents"
}

// CreateLogContent 创建日志内容记录
func CreateLogContent(c *gin.Context, logId int, requestBody, responseBody string) error {
	if !common.LogContentsEnabled {
		return nil
	}

	logContent := &LogContent{
		LogId:        logId,
		RequestBody:  requestBody,
		ResponseBody: responseBody,
	}

	err := LOG_DB.Create(logContent).Error
	if err != nil {
		// 如果表不存在，尝试创建表
		if strings.Contains(err.Error(), "no such table") || strings.Contains(err.Error(), "doesn't exist") {
			logger.LogInfo(c, "log_contents table not found, attempting to create it")
			if migErr := LOG_DB.AutoMigrate(&LogContent{}); migErr != nil {
				logger.LogError(c, "failed to auto-migrate log_contents table: "+migErr.Error())
				return migErr
			}
			// 创建表后重试
			err = LOG_DB.Create(logContent).Error
		}
		if err != nil {
			logger.LogError(c, "failed to create log content: "+err.Error())
		}
	}
	return err
}

// GetLogContentByLogId 根据日志ID获取详细内容
func GetLogContentByLogId(logId int) (*LogContent, error) {
	var logContent LogContent
	err := LOG_DB.Where("log_id = ?", logId).First(&logContent).Error
	if err != nil {
		return nil, err
	}
	return &logContent, nil
}

// DeleteLogContentByLogId 删除日志内容
func DeleteLogContentByLogId(logId int) error {
	return LOG_DB.Where("log_id = ?", logId).Delete(&LogContent{}).Error
}

// DeleteOldLogContent 删除旧的日志内容
func DeleteOldLogContent(targetTimestamp int64, limit int) (int64, error) {
	// 先查找需要删除的 log_id
	var logIds []int
	err := LOG_DB.Table("logs").
		Select("id").
		Where("created_at < ?", targetTimestamp).
		Limit(limit).
		Pluck("id", &logIds).Error

	if err != nil {
		return 0, err
	}

	if len(logIds) == 0 {
		return 0, nil
	}

	// 删除对应的 log_contents
	result := LOG_DB.Where("log_id IN ?", logIds).Delete(&LogContent{})
	return result.RowsAffected, result.Error
}
