package middleware

import (
	"bytes"
	"io"
	"time"

	"gophermart/internal/gophermart/application/port"

	"github.com/gin-gonic/gin"
)

// LogFormatter abstraction for logging HTTP request details.
type LogFormatter interface {
	Log(log port.Logger, params LogParams)
}

// LogParams contains data available for logging.
type LogParams struct {
	Ctx          *gin.Context
	Duration     time.Duration
	RequestBody  []byte
	ResponseBody *bytes.Buffer
}

// DefaultLogFormatter implements standard logging logic.
type DefaultLogFormatter struct{}

func (f *DefaultLogFormatter) Log(log port.Logger, p LogParams) {
	log.Info("HTTP request",
		"uri", p.Ctx.Request.RequestURI,
		"method", p.Ctx.Request.Method,
		"duration", p.Duration,
		"status", p.Ctx.Writer.Status(),
		"size", p.Ctx.Writer.Size(),
	)
	log.Debug("HTTP request/response body",
		"request_body", string(p.RequestBody),
		"response_body", p.ResponseBody.String(),
	)
}

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

// Logger middleware with injected formatter.
func Logger(log port.Logger, formatter LogFormatter) gin.HandlerFunc {
	if formatter == nil {
		formatter = &DefaultLogFormatter{}
	}

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

		formatter.Log(log, LogParams{
			Ctx:          c,
			Duration:     time.Since(start),
			RequestBody:  requestBody,
			ResponseBody: w.body,
		})
	}
}
