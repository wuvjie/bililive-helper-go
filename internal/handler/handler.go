package handler

import (
	"bililive-helper/internal/config"
	"bililive-helper/internal/service"

	"go.uber.org/zap"
)

type Handler struct {
	config    *config.Config
	logger    *zap.Logger
	merge     *service.MergeService
	clean     *service.CleanService
	history   *service.HistoryService
	scheduler *service.SchedulerService
}

func NewHandler(config *config.Config, logger *zap.Logger, merge *service.MergeService, clean *service.CleanService, history *service.HistoryService, scheduler *service.SchedulerService) *Handler {
	return &Handler{
		config:    config,
		logger:    logger,
		merge:     merge,
		clean:     clean,
		history:   history,
		scheduler: scheduler,
	}
}
