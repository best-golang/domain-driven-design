package rest

import (
	"bufio"
	"fmt"
	"github.com/a-trium/domain-driven-design/implementation-1ambda/service-gateway/internal/config"
	"net"
	"net/http"
	"time"
)

// wrap `http.ResponseWriter` to persist status and size which are not exposed
type statusWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (c *statusWriter) WriteHeader(status int) {
	c.status = status
	c.ResponseWriter.WriteHeader(status)
}

func (c *statusWriter) Write(b []byte) (int, error) {
	// When WriteHeader is not called, it's safe to assume the status will be 200.
	if c.status == 0 {
		c.status = 200
	}

	size, err := c.ResponseWriter.Write(b)
	c.size += size
	return size, err
}

func (c *statusWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := c.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, fmt.Errorf("ResponseWriter does not implement the Hijacker interface")
}

func InjectHttpLoggingMiddleware(next http.Handler) http.Handler {
	env := config.Env
	logger := config.GetLogger().With("service_name", env.ServiceName, "service_id", env.ServiceId)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// if cors
		if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
			next.ServeHTTP(w, r)
			return
		}

		if !env.DebugHTTPEnabled() {
			next.ServeHTTP(w, r)
			return
		}

		// TODO: security flag + session check middleware

		start := time.Now()
		sw := statusWriter{ResponseWriter: w}
		next.ServeHTTP(&sw, r)
		latency := time.Since(start)

		logger.Infow("HTTP",
			"status", sw.status,
			"size", sw.size,
			"duration", latency,
			"remote", r.RemoteAddr,
			"request", r.RequestURI,
			"method", r.Method,
		)
	})
}
