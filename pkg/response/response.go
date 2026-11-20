package response

import (
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalPages int         `json:"total_pages"`
}

// ensureSlice converts a nil slice into an empty slice so JSON serializes it as
// `[]` instead of `null`. Non-slice values are returned unchanged.
func ensureSlice(data interface{}) interface{} {
	if data == nil {
		return data
	}
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return data
		}
		v = v.Elem()
	}
	if v.Kind() == reflect.Slice && v.IsNil() {
		return reflect.MakeSlice(v.Type(), 0, 0).Interface()
	}
	return data
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Success: true, Data: ensureSlice(data)})
}

func OKMessage(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Response{Success: true, Message: message})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{Success: true, Data: ensureSlice(data)})
}

func Paginated(c *gin.Context, data interface{}, total int64, page, limit int) {
	totalPages := 0
	if limit > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}
	c.JSON(http.StatusOK, PaginatedResponse{
		Success:    true,
		Data:       ensureSlice(data),
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}

func BadRequest(c *gin.Context, err string) {
	c.JSON(http.StatusBadRequest, Response{Success: false, Error: err})
}

func Unauthorized(c *gin.Context, err string) {
	c.JSON(http.StatusUnauthorized, Response{Success: false, Error: err})
}

func Forbidden(c *gin.Context, err string) {
	c.JSON(http.StatusForbidden, Response{Success: false, Error: err})
}

func NotFound(c *gin.Context, err string) {
	c.JSON(http.StatusNotFound, Response{Success: false, Error: err})
}

func Conflict(c *gin.Context, err string) {
	c.JSON(http.StatusConflict, Response{Success: false, Error: err})
}

func InternalError(c *gin.Context, err string) {
	// Strip any client-facing detail we might have added and log the raw event.
	log.Printf("[error] internal server error: %s", strings.TrimSpace(err))
	c.JSON(http.StatusInternalServerError, Response{Success: false, Error: err})
}
