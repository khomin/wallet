package dto

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}

func InvalidParameters(c *gin.Context) {
	c.JSON(http.StatusBadRequest, ErrorResponse{
		Code:    "BAD_REQUEST",
		Message: "invalid_parameters",
	})
}

func InvalidParametersMessage(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, ErrorResponse{
		Code:    "BAD_REQUEST",
		Message: message,
	})
}

func InternallError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Code:    "INTERNAL_ERROR",
		Message: "internal_error",
	})
}

func InternallErrorMessage(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Code:    "INTERNAL_ERROR",
		Message: message,
	})
}

func UnauthorizedError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Code:    "UNAUTHORIZED_ERROR",
		Message: "unauthorized",
	})
}

func UnauthorizedErrorMessage(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Code:    "UNAUTHORIZED_ERROR",
		Message: message,
	})
}

func NotFoundErrorMessage(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, ErrorResponse{
		Code:    "NOT_FOUND",
		Message: message,
	})
}
