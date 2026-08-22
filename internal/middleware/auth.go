package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"splitobillo/internal/config"
	"splitobillo/pkg/apperr"
	"splitobillo/pkg/response"
)

const (
	CtxSessionID       = "session_id"
	headerSessionToken = "X-Session-Token"
)

type sessionClaims struct {
	SID string `json:"sid"`
	jwt.RegisteredClaims
}

func Auth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		sid := parseSID(c.GetHeader(headerSessionToken), cfg)

		if sid == "" {
			sid = uuid.NewString()
			claims := sessionClaims{
				SID: sid,
				RegisteredClaims: jwt.RegisteredClaims{
					IssuedAt:  jwt.NewNumericDate(time.Now()),
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.SessionTTL)),
				},
			}
			token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(cfg.JWTSecret))
			if err != nil {
				response.Error(c, apperr.Internal("failed to create session token"))
				c.Abort()
				return
			}
			c.Header(headerSessionToken, token)
		}

		c.Set(CtxSessionID, sid)
		c.Next()
	}
}

func parseSID(tokenStr string, cfg *config.Config) string {
	if tokenStr == "" {
		return ""
	}
	parsed, err := jwt.ParseWithClaims(tokenStr, &sessionClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(cfg.JWTSecret), nil
	})
	if err != nil {
		return ""
	}
	if claims, ok := parsed.Claims.(*sessionClaims); ok && parsed.Valid && claims.SID != "" {
		return claims.SID
	}
	return ""
}
