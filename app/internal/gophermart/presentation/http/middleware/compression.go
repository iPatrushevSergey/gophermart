package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"

	"gophermart/internal/gophermart/application/port"

	"github.com/gin-gonic/gin"
)

// Compressor defines a strategy for compression algorithms (gzip, brotli, etc.).
type Compressor interface {
	// ContentEncoding returns the header value (e.g. "gzip").
	ContentEncoding() string
	// NewReader returns a reader that decompresses data from r.
	NewReader(r io.Reader) (io.ReadCloser, error)
	// NewWriter returns a writer that compresses data to w.
	// The returned writer must be closed to flush data and release resources.
	NewWriter(w io.Writer) io.WriteCloser
}

// Compress middleware supporting multiple compression strategies.
func Compress(log port.Logger, compressors ...Compressor) gin.HandlerFunc {
	// Map for fast lookup by Content-Encoding
	compressorsByEncoding := make(map[string]Compressor)
	for _, c := range compressors {
		compressorsByEncoding[c.ContentEncoding()] = c
	}

	return func(c *gin.Context) {
		// Always set Vary header
		c.Header("Vary", "Accept-Encoding")

		// Decompress Request
		reqEncoding := c.GetHeader("Content-Encoding")
		if reqEncoding != "" {
			if comp, ok := compressorsByEncoding[reqEncoding]; ok {
				cr, err := comp.NewReader(c.Request.Body)
				if err != nil {
					log.Error("failed to create decompress reader", "error", err, "encoding", reqEncoding)
					c.AbortWithStatus(http.StatusBadRequest)
					return
				}
				c.Request.Body = cr
				defer cr.Close()
			}
		}

		// Negotiate Response Compression
		accept := c.GetHeader("Accept-Encoding")
		var target Compressor
		// Simple negotiation: pick first matching compressor
		for _, comp := range compressors {
			if strings.Contains(accept, comp.ContentEncoding()) {
				target = comp
				break
			}
		}

		if target == nil {
			c.Next()
			return
		}

		// Wrap Response Writer
		gz := target.NewWriter(c.Writer)

		originalWriter := c.Writer
		cw := &compressWriter{
			ResponseWriter: c.Writer,
			writer:         gz,
			encoding:       target.ContentEncoding(),
		}
		c.Writer = cw

		defer func() {
			c.Writer = originalWriter
			if cw.shouldCompress {
				// Close closes the compressor and returns it to pool (if applicable)
				_ = cw.writer.Close()
			}
		}()

		c.Next()
	}
}

// GzipCompressor implements Compressor using sync.Pool for performance.
type GzipCompressor struct {
	pool sync.Pool
}

func NewGzipCompressor() *GzipCompressor {
	return &GzipCompressor{
		pool: sync.Pool{
			New: func() any {
				w, _ := gzip.NewWriterLevel(io.Discard, gzip.BestSpeed)
				return w
			},
		},
	}
}

func (g *GzipCompressor) ContentEncoding() string {
	return "gzip"
}

func (g *GzipCompressor) NewReader(r io.Reader) (io.ReadCloser, error) {
	return gzip.NewReader(r)
}

func (g *GzipCompressor) NewWriter(w io.Writer) io.WriteCloser {
	gz := g.pool.Get().(*gzip.Writer)
	gz.Reset(w)
	return &pooledGzipWriter{
		Writer: gz,
		pool:   &g.pool,
	}
}

// pooledGzipWriter wraps gzip.Writer to return it to pool on Close.
type pooledGzipWriter struct {
	*gzip.Writer
	pool *sync.Pool
}

func (w *pooledGzipWriter) Close() error {
	err := w.Writer.Close()
	w.pool.Put(w.Writer)
	return err
}

// compressWriter wraps gin.ResponseWriter to intercept headers and write compressed data.
type compressWriter struct {
	gin.ResponseWriter
	writer   io.WriteCloser
	encoding string

	wroteHeader    bool
	shouldCompress bool
}

func (cw *compressWriter) Write(data []byte) (int, error) {
	if !cw.wroteHeader {
		cw.WriteHeader(http.StatusOK)
	}
	if cw.shouldCompress {
		return cw.writer.Write(data)
	}
	return cw.ResponseWriter.Write(data)
}

func (cw *compressWriter) WriteHeader(code int) {
	if cw.wroteHeader {
		return
	}
	cw.wroteHeader = true

	contentType := cw.Header().Get("Content-Type")
	if shouldCompress(contentType) {
		cw.shouldCompress = true
		cw.Header().Set("Content-Encoding", cw.encoding)
		cw.Header().Del("Content-Length")
	}

	cw.ResponseWriter.WriteHeader(code)
}

func (cw *compressWriter) WriteString(s string) (int, error) {
	return cw.Write([]byte(s))
}

func shouldCompress(contentType string) bool {
	return strings.Contains(contentType, "application/json") ||
		strings.Contains(contentType, "text/html") ||
		strings.Contains(contentType, "text/plain") ||
		strings.Contains(contentType, "application/xml") ||
		contentType == ""
}
