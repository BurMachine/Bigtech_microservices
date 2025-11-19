package adminserver

import (
	"context"
	"net/http"
	"net/http/pprof"

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

	// Pprof endpoints
	if cfg.Pprof.Enabled {
		pprofPath := cfg.Pprof.Path
		if pprofPath == "" {
			pprofPath = "/debug/pprof"
		}

		registerPprofHandlers(router, pprofPath, logger)

		logger.Info(context.Background(), "pprof endpoints registered",
			"path", pprofPath,
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

// registerPprofHandlers регистрирует все pprof handlers
func registerPprofHandlers(router *gin.Engine, basePath string, logger *loggerlib.Logger) {
	// Index - список всех доступных профилей
	router.GET(basePath+"/", gin.WrapF(pprof.Index))

	// Cmdline - командная строка запуска программы
	router.GET(basePath+"/cmdline", gin.WrapF(pprof.Cmdline))

	// Profile - CPU профиль (по умолчанию 30 секунд)
	router.GET(basePath+"/profile", gin.WrapF(pprof.Profile))

	// Symbol - информация о символах
	router.GET(basePath+"/symbol", gin.WrapF(pprof.Symbol))

	// Trace - trace выполнения (по умолчанию 1 секунда)
	router.GET(basePath+"/trace", gin.WrapF(pprof.Trace))

	// Специфичные профили через handler
	router.GET(basePath+"/heap", gin.WrapH(pprof.Handler("heap")))                 // Heap (память)
	router.GET(basePath+"/goroutine", gin.WrapH(pprof.Handler("goroutine")))       // Goroutines
	router.GET(basePath+"/block", gin.WrapH(pprof.Handler("block")))               // Блокировки
	router.GET(basePath+"/threadcreate", gin.WrapH(pprof.Handler("threadcreate"))) // Создание потоков
	router.GET(basePath+"/mutex", gin.WrapH(pprof.Handler("mutex")))               // Mutex профиль
	router.GET(basePath+"/allocs", gin.WrapH(pprof.Handler("allocs")))             // Аллокации

	logger.Info(context.Background(), "registered pprof handlers",
		"endpoints", []string{
			basePath + "/",
			basePath + "/cmdline",
			basePath + "/profile",
			basePath + "/heap",
			basePath + "/goroutine",
			basePath + "/block",
			basePath + "/mutex",
			basePath + "/trace",
		},
	)
}

// Port возвращает порт на котором работает admin server
func (s *Server) Port() uint64 {
	if s.ginEntry == nil {
		return 0
	}
	return s.ginEntry.Port
}
