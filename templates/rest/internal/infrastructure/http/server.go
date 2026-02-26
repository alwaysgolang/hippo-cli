package http

import (
	"context"
	"net/http"
	"strconv"
	"time"

	pingController "gotemplate/internal/adapter/http/controllers/ping"
	"gotemplate/internal/config"
	"gotemplate/pkg/logs"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Server struct {
	Engine         *gin.Engine
	httpServer     *http.Server
	appCfg         *config.AppConfig
	httpCfg        *config.HTTPConfig
	PingController *pingController.Controller
}

const (
	RequestIDHeader = "X-Request-ID"
	requestIDKey    = "request_id"
)

func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set(requestIDKey, requestID)
		detachedCtx := context.WithoutCancel(c)

		ctx := logs.ContextWithRequestID(detachedCtx, requestID)

		c.Request = c.Request.WithContext(ctx)
		c.Header(RequestIDHeader, requestID)
		c.Next()
	}
}

func zapLogger(cfg *config.AppConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		if cfg.Mode == gin.DebugMode {
			logs.InfoCtx(c, "HTTP Request",
				"status", c.Writer.Status(),
				"method", c.Request.Method,
				"path", path,
				"query", query,
				"ip", c.ClientIP(),
				"user-agent", c.Request.UserAgent(),
				"latency", latency,
			)
		} else {
			logs.InfoCtx(c, "HTTP Request",
				"status", c.Writer.Status(),
				"method", c.Request.Method,
				"path", path,
				"latency", latency,
			)
		}
	}
}

func zapRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logs.ErrorCtx(c.Request.Context(), "Panic recovered",
					"error", err,
					"path", c.Request.URL.Path,
				)
				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	}
}

func NewServer(
	appCfg *config.AppConfig,
	httpCfg *config.HTTPConfig,
	pingCtrl *pingController.Controller,
) (*Server, func(), error) {
	gin.SetMode(appCfg.Mode)
	engine := gin.New()

	engine.Use(requestIDMiddleware())
	engine.Use(zapRecovery())
	engine.Use(zapLogger(appCfg))

	server := &Server{
		appCfg:         appCfg,
		httpCfg:        httpCfg,
		Engine:         engine,
		PingController: pingCtrl,
	}

	cleanupServer := func() {
		logs.Info("Shutting down HTTP server...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if server.httpServer != nil {
			_ = server.httpServer.Shutdown(ctx)
		}
	}

	return server, cleanupServer, nil
}

func (s *Server) Run() error {
	s.registerRoutes()
	s.httpServer = &http.Server{
		Addr:              ":" + strconv.Itoa(s.httpCfg.Port),
		Handler:           s.Engine,
		ReadHeaderTimeout: 1 * time.Second,
	}
	return s.httpServer.ListenAndServe()
}
