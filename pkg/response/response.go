package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"splitobillo/pkg/apperr"
)

type successEnvelope struct {
	Success bool `json:"success"`
	Data    any  `json:"data,omitempty"`
}

type errorEnvelope struct {
	Success bool      `json:"success"`
	Error   errorBody `json:"error"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, successEnvelope{Success: true, Data: data})
}

func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, successEnvelope{Success: true, Data: data})
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func Error(c *gin.Context, err error) {
	var appErr *apperr.AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.HTTPStatus, errorEnvelope{
			Success: false,
			Error:   errorBody{Code: appErr.Code, Message: appErr.Message},
		})
		return
	}
	c.JSON(http.StatusInternalServerError, errorEnvelope{
		Success: false,
		Error:   errorBody{Code: "INTERNAL_SERVER_ERROR", Message: "Internal server error"},
	})
}
