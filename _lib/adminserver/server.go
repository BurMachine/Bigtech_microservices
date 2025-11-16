package adminserver

import (
	"context"
	"net/http"

	loggerlib "github.com/Burmachine/MSA/lib/logger"
	platform_middleware "github.com/Burmachine/MSA/lib/middleware"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	rkgin "github.com/rookie-ninja/rk-gin/v2/boot"
)

type Server struct {
	config   platform_middleware.AdminConfig
	logger   *loggerlib.Logger
	ginEntry *rkgin.GinEntry
}

// New регистрирует admin endpoints на существующий GinEntry
func New(cfg platform_middleware.AdminConfig, logger *loggerlib.Logger, ginEntry *rkgin.GinEntry) *Server {
	if ginEntry == nil {
		logger.Warn(context.Background(), "gin entry is nil, admin endpoints not registered")
		return &Server{
			config: cfg,
			logger: logger,
		}
	}

	router := ginEntry.Router

	// Metrics endpoint
	if cfg.Metrics.Enabled {
		metricsPath := cfg.Metrics.Path
		if metricsPath == "" {
			metricsPath = "/metrics"
		}

		router.GET(metricsPath, gin.WrapH(promhttp.Handler()))

		logger.Info(context.Background(), "metrics endpoint registered",
			"path", metricsPath,
			"port", ginEntry.Port,
		)
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	// Readiness check
	router.GET("/ready", func(c *gin.Context) {
		c.String(http.StatusOK, "READY")
	})

	logger.Info(context.Background(), "admin endpoints registered",
		"port", ginEntry.Port,
	)

	return &Server{
		config:   cfg,
		logger:   logger,
		ginEntry: ginEntry,
	}
}

// Port возвращает порт на котором работает admin server
func (s *Server) Port() uint64 {
	if s.ginEntry == nil {
		return 0
	}
	return s.ginEntry.Port
}
