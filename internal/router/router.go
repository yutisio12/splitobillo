package router

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"splitobillo/internal/config"
	"splitobillo/internal/handler"
	"splitobillo/internal/health"
	"splitobillo/internal/middleware"
	"splitobillo/internal/repository"
	"splitobillo/internal/service"
	"splitobillo/pkg/response"
)

func New(cfg *config.Config, db *gorm.DB, logger *slog.Logger) *gin.Engine {
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(
		middleware.RequestID(),
		middleware.Logger(logger),
		middleware.Recovery(logger),
		middleware.CORS(cfg),
		middleware.RateLimit(cfg),
	)

	r.GET("/health", health.Handler(db, cfg))

	receiptRepo := repository.NewReceiptRepository(db)
	ocrService := service.NewOCRService(cfg)
	receiptService := service.NewReceiptService(receiptRepo, cfg, ocrService)
	receiptHandler := handler.NewReceiptHandler(receiptService)

	api := r.Group("/api/v1")
	api.Use(middleware.Auth(cfg))
	{
		api.GET("/ping", func(c *gin.Context) {
			response.OK(c, gin.H{"session_id": c.GetString(middleware.CtxSessionID)})
		})

		api.POST("/receipts", receiptHandler.Upload)
		api.GET("/receipts", receiptHandler.List)
		api.GET("/receipts/:id", receiptHandler.Get)
		api.PUT("/receipts/:id", receiptHandler.Update)
		api.DELETE("/receipts/:id", receiptHandler.Delete)
		api.POST("/receipts/:id/ocr/retry", receiptHandler.RetryOCR)
	}

	return r
}
