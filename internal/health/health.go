package health

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"splitobillo/internal/config"
	"splitobillo/pkg/apperr"
	"splitobillo/pkg/response"
)

func Handler(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		components := gin.H{"db": "ok", "ocr": "ok"}
		healthy := true

		sqlDB, err := db.DB()
		if err != nil || sqlDB.PingContext(c.Request.Context()) != nil {
			components["db"] = "unreachable"
			healthy = false
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.OCRServiceURL+"/health", nil)
		if err != nil || httpGet(req) != nil {
			components["ocr"] = "unreachable"
			healthy = false
		}

		if !healthy {
			response.Error(c, apperr.ServiceUnavailable("one or more dependencies unavailable"))
			return
		}
		response.OK(c, gin.H{"status": "ok", "components": components})
	}
}

func httpGet(req *http.Request) error {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return apperr.ServiceUnavailable("ocr service unhealthy")
	}
	return nil
}
