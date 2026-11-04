package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"bililive-helper-go/internal/config"

	"github.com/gin-gonic/gin"
)

// GetConfig 返回当前配置（不含敏感字段）。
func (h *Handler) GetConfig(c *gin.Context) {
	ok(c, h.config.ToDTO())
}

// SaveConfig 保存配置更新（部分更新，只修改请求中包含的字段）。
// 记录变更日志并异步写入历史记录。
func (h *Handler) SaveConfig(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 64*1024) // 64KB 限制
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		failBadRequest(c, "参数错误")
		return
	}
	// 序列化回 JSON 以使用类型安全的 ApplyFromJSON
	data, _ := json.Marshal(req)

	var changeDetail string
	if err := h.config.Apply(func() error {
		old := h.config.ToDTOSnapshot()
		h.config.ApplyFromJSON(data)
		if err := h.config.Validate(); err != nil {
			return err
		}
		changeDetail = config.DiffDTO(old, h.config.ToDTOSnapshot())
		return nil
	}); err != nil {
		failBadRequest(c, err.Error())
		return
	}

	if changeDetail != "" {
		h.logger.Info(changeDetail)
		go h.history.Add("config", "", "success", changeDetail, "")
	} else {
		go h.history.Add("config", "", "success", "配置未变更", "")
	}
	okStatus(c)
}

// DefaultConfig 返回默认配置值。
func (h *Handler) DefaultConfig(c *gin.Context) {
	d := config.DefaultConfig()
	ok(c, gin.H{
		"TRIGGER_THRESHOLD":     d.TriggerThreshold,
		"TARGET_THRESHOLD":      d.TargetThreshold,
		"MIN_KEEP_PER_STREAMER": d.MinKeepPerStreamer,
		"SAFE_AGE_MINUTES":      d.SafeAgeMinutes,
		"SAFE_MODE":             d.SafeMode,
		"SAFE_DAYS":             d.SafeDays,
		"GAP_MINUTES":           d.GapMinutes,
		"MERGE_AGE_MINUTES":     d.MergeAgeMinutes,
		"MAX_DELETE_PER_RUN":    d.MaxDeletePerRun,
		"BACKUP_START_HOUR":    d.BackupStartHour,
		"BACKUP_START_MINUTE":  d.BackupStartMinute,
		"BACKUP_END_HOUR":      d.BackupEndHour,
		"BACKUP_END_MINUTE":    d.BackupEndMinute,
	})
}

// ExportConfig 导出纯净配置数据（含全局策略、调度参数和主播规则）。
func (h *Handler) ExportConfig(c *gin.Context) {
	cfg := h.config.ToDTO()
	schedule := h.scheduler.GetStatus()

	ok(c, gin.H{
		"version":     "2.0.0",
		"exported_at": time.Now().Format(time.RFC3339),
		"config":      cfg,
		"schedule":    schedule,
	})
}

// ImportConfig 从导入数据中恢复配置。
func (h *Handler) ImportConfig(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1024*1024) // 1MB 限制（含历史记录）
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		failBadRequest(c, "参数错误")
		return
	}

	cfgData, ok := data["config"].(map[string]interface{})
	if !ok {
		failBadRequest(c, "缺少配置数据")
		return
	}

	// 过滤掉不应导入的内部字段
	delete(cfgData, "SESSION_VERSION")

	// 序列化为 JSON 以使用类型安全的 ApplyFromJSON
	cfgJSON, _ := json.Marshal(cfgData)
	if err := h.config.Apply(func() error {
		h.config.ApplyFromJSON(cfgJSON)
		return h.config.Validate()
	}); err != nil {
		failBadRequest(c, err.Error())
		return
	}

	go h.history.Add("config", "", "success", fmt.Sprintf("配置已导入（%d 项）", len(cfgData)), "")
	okMsg(c, "配置导入成功")
}
