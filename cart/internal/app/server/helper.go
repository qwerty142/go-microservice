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

func writeError(ctx context.Context, writer http.ResponseWriter, status int, message string) {
	writer.WriteHeader(status)

	if _, err := writer.Write([]byte(message)); err != nil {
		logger.Errorw(ctx, "Writer error", "error", err)
	}
}

func setInternalServerError(ctx context.Context, writer http.ResponseWriter, message string) {
	writeError(ctx, writer, http.StatusInternalServerError, message)
}

func setBadRequestError(ctx context.Context, writer http.ResponseWriter, message string) {
	writeError(ctx, writer, http.StatusBadRequest, message)
}
