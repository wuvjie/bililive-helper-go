// Package handler 提供 HTTP 请求处理器。
// 包含认证、配置管理、状态查询、合并/清理任务执行、历史记录等 API 端点的处理逻辑。
package handler

import (
	"bililive-helper/internal/config"
	"bililive-helper/internal/service"
	"sync"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// Handler 持有所有 HTTP 处理器的依赖。
// 包含配置、日志、各业务服务实例，以及 bcrypt 哈希后的密码（运行时比对用）。
type Handler struct {
	config         *config.Config
	logger         *zap.Logger
	merge          *service.MergeService
	clean          *service.CleanService
	history        *service.HistoryService
	scheduler      *service.SchedulerService
	passwordMu     sync.RWMutex
	hashedPassword string
}

// NewHandler 创建一个新的 Handler 实例。
// 启动时将明文密码 bcrypt 哈希，后续登录比较哈希值而非明文。
func NewHandler(config *config.Config, logger *zap.Logger, merge *service.MergeService, clean *service.CleanService, history *service.HistoryService, scheduler *service.SchedulerService) *Handler {
	hashed, _ := bcrypt.GenerateFromPassword([]byte(config.Password), bcrypt.DefaultCost)
	return &Handler{
		config:         config,
		logger:         logger,
		merge:          merge,
		clean:          clean,
		history:        history,
		scheduler:      scheduler,
		hashedPassword: string(hashed),
	}
}
