package utils

import (
	"crypto/rand"
	"encoding/hex"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func GetPagination(c *gin.Context) (int, int, int) {
	page := 1
	limit := 20

	if p := c.Query("page"); p != "" {
		if pInt, err := strconv.Atoi(p); err == nil && pInt > 0 {
			page = pInt
		}
	}
	if l := c.Query("limit"); l != "" {
		if lInt, err := strconv.Atoi(l); err == nil && lInt > 0 && lInt <= 100 {
			limit = lInt
		}
	}

	offset := (page - 1) * limit
	return page, limit, offset
}

func StringPtr(s string) *string {
	return &s
}
