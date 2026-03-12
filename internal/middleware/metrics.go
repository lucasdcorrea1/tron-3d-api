package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Metrics struct {
	mu              sync.RWMutex
	requestsTotal   map[string]int64
	requestDuration map[string][]float64
	activeRequests  int64
	startTime       time.Time

	// 3D Store metrics
	ordersCreated   int64
	productsCreated int64
	authErrors      int64
}

var metrics = &Metrics{
	requestsTotal:   make(map[string]int64),
	requestDuration: make(map[string][]float64),
	startTime:       time.Now(),
}

func GetMetrics() *Metrics {
	return metrics
}

func IncOrderCreated() {
	metrics.mu.Lock()
	metrics.ordersCreated++
	metrics.mu.Unlock()
}

func IncProductCreated() {
	metrics.mu.Lock()
	metrics.productsCreated++
	metrics.mu.Unlock()
}

func IncAuthError() {
	metrics.mu.Lock()
	metrics.authErrors++
	metrics.mu.Unlock()
}

// MetricsMiddleware collects HTTP metrics
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()

		metrics.mu.Lock()
		metrics.activeRequests++
		metrics.mu.Unlock()

		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)

		duration := time.Since(start).Seconds()

		metrics.mu.Lock()
		metrics.activeRequests--
		key := r.Method + "_" + normalizePathForMetrics(r.URL.Path) + "_" + strconv.Itoa(rw.status)
		metrics.requestsTotal[key]++
		metrics.requestDuration[key] = append(metrics.requestDuration[key], duration)
		metrics.mu.Unlock()
	})
}

func normalizePathForMetrics(path string) string {
	segments := []string{}
	for _, seg := range splitPath(path) {
		if isID(seg) {
			segments = append(segments, ":id")
		} else {
			segments = append(segments, seg)
		}
	}
	if len(segments) == 0 {
		return "/"
	}
	result := ""
	for _, s := range segments {
		result += "/" + s
	}
	return result
}

func splitPath(path string) []string {
	var result []string
	for _, seg := range strings.Split(path, "/") {
		if seg != "" {
			result = append(result, seg)
		}
	}
	return result
}

func isID(s string) bool {
	if len(s) == 24 {
		for _, c := range s {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return false
			}
		}
		return true
	}
	return false
}

// PrometheusHandler returns metrics in Prometheus format
func PrometheusHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metrics.mu.RLock()
		defer metrics.mu.RUnlock()

		w.Header().Set("Content-Type", "text/plain; version=0.0.4")

		w.Write([]byte("# HELP http_requests_total Total number of HTTP requests\n"))
		w.Write([]byte("# TYPE http_requests_total counter\n"))

		for key, count := range metrics.requestsTotal {
			method, path, status := parseKey(key)
			line := "http_requests_total{method=\"" + method + "\",path=\"" + path + "\",status=\"" + status + "\"} " + strconv.FormatInt(count, 10) + "\n"
			w.Write([]byte(line))
		}

		w.Write([]byte("\n# HELP http_active_requests Current number of active requests\n"))
		w.Write([]byte("# TYPE http_active_requests gauge\n"))
		w.Write([]byte("http_active_requests " + strconv.FormatInt(metrics.activeRequests, 10) + "\n"))

		w.Write([]byte("\n# HELP app_uptime_seconds Application uptime in seconds\n"))
		w.Write([]byte("# TYPE app_uptime_seconds counter\n"))
		uptime := time.Since(metrics.startTime).Seconds()
		w.Write([]byte("app_uptime_seconds " + strconv.FormatFloat(uptime, 'f', 0, 64) + "\n"))

		w.Write([]byte("\n# HELP store3d_orders_created_total Total number of orders created\n"))
		w.Write([]byte("# TYPE store3d_orders_created_total counter\n"))
		w.Write([]byte("store3d_orders_created_total " + strconv.FormatInt(metrics.ordersCreated, 10) + "\n"))

		w.Write([]byte("\n# HELP store3d_products_created_total Total number of products created\n"))
		w.Write([]byte("# TYPE store3d_products_created_total counter\n"))
		w.Write([]byte("store3d_products_created_total " + strconv.FormatInt(metrics.productsCreated, 10) + "\n"))

		w.Write([]byte("\n# HELP auth_errors_total Total number of authentication errors\n"))
		w.Write([]byte("# TYPE auth_errors_total counter\n"))
		w.Write([]byte("auth_errors_total " + strconv.FormatInt(metrics.authErrors, 10) + "\n"))
	})
}

func parseKey(key string) (method, path, status string) {
	first := -1
	last := -1
	for i, c := range key {
		if c == '_' {
			if first == -1 {
				first = i
			} else {
				last = i
			}
		}
	}
	if first > 0 && last > first {
		method = key[:first]
		path = key[first+1 : last]
		status = key[last+1:]
	}
	return
}
