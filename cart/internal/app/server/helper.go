package server

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"route256/cart/internal/pkg/util/logger"
)

const (
	TracerName = "cart"
	base10     = 10
	bitsInt64  = 64
)

func parseInt64(str string) (int64, error) {
	value, err := strconv.ParseInt(str, base10, bitsInt64)
	if err != nil {
		return value, fmt.Errorf("parse error: %w", err)
	}

	return value, nil
}

func setInternalServerError(writer http.ResponseWriter, message string) {
	writer.WriteHeader(http.StatusInternalServerError)

	_, err := writer.Write([]byte(message))
	if err != nil {
		logger.Errorw(context.TODO(), "Writer error", "error", err)
		return
	}
}
