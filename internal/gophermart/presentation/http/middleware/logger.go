package middleware

import (
	"bytes"
	"io"
	"time"

	"gophermart/internal/gophermart/application/port"

	"github.com/gin-gonic/gin"
)

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func (r responseBodyWriter) WriteString(s string) (int, error) {
	r.body.WriteString(s)
	return r.ResponseWriter.WriteString(s)
}

// Logger logs HTTP requests: URI, method, duration, status, size. Debug: request/response bodies.
func Logger(log port.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		var requestBody []byte
		if c.Request.Body != nil && c.Request.ContentLength > 0 {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		w := &responseBodyWriter{body: bytes.NewBuffer(nil), ResponseWriter: c.Writer}
		c.Writer = w

		c.Next()

		duration := time.Since(start)
		log.Info("HTTP request",
			"uri", c.Request.RequestURI,
			"method", c.Request.Method,
			"duration", duration,
			"status", c.Writer.Status(),
			"size", c.Writer.Size(),
		)
		log.Debug("HTTP request/response body",
			"request_body", string(requestBody),
			"response_body", w.body.String(),
		)
	}
}
