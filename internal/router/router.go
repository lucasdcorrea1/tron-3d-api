package router

import (
	"net/http"

	"github.com/tron-legacy/tron-3d-api/internal/handlers"
	"github.com/tron-legacy/tron-3d-api/internal/middleware"
)

func New() http.Handler {
	mux := http.NewServeMux()

	// ==========================================
	// PUBLIC ROUTES
	// ==========================================

	// Health check
	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Prometheus metrics
	mux.Handle("GET /metrics", middleware.PrometheusHandler())

	// Products (public)
	mux.HandleFunc("GET /api/v1/products", handlers.ListProducts)
	mux.HandleFunc("GET /api/v1/products/{slug}", handlers.GetProductBySlug)

	// Categories (public)
	mux.HandleFunc("GET /api/v1/categories", handlers.ListCategories)

	// ==========================================
	// AUTH REQUIRED ROUTES
	// ==========================================

	// Orders (auth required)
	mux.Handle("POST /api/v1/orders", middleware.Auth(http.HandlerFunc(handlers.CreateOrder)))
	mux.Handle("GET /api/v1/orders", middleware.Auth(http.HandlerFunc(handlers.ListMyOrders)))
	mux.Handle("GET /api/v1/orders/{id}", middleware.Auth(http.HandlerFunc(handlers.GetOrder)))

	// ==========================================
	// ADMIN ROUTES (auth + admin check)
	// ==========================================

	adminRoute := func(h http.HandlerFunc) http.Handler {
		return middleware.Auth(middleware.Admin(http.HandlerFunc(h)))
	}

	// Admin Products
	mux.Handle("GET /api/v1/admin/products", adminRoute(handlers.AdminListProducts))
	mux.Handle("POST /api/v1/admin/products", adminRoute(handlers.AdminCreateProduct))
	mux.Handle("PUT /api/v1/admin/products/{id}", adminRoute(handlers.AdminUpdateProduct))
	mux.Handle("DELETE /api/v1/admin/products/{id}", adminRoute(handlers.AdminDeleteProduct))

	// Admin Categories
	mux.Handle("POST /api/v1/admin/categories", adminRoute(handlers.AdminCreateCategory))
	mux.Handle("PUT /api/v1/admin/categories/{id}", adminRoute(handlers.AdminUpdateCategory))
	mux.Handle("DELETE /api/v1/admin/categories/{id}", adminRoute(handlers.AdminDeleteCategory))

	// Admin Orders
	mux.Handle("GET /api/v1/admin/orders", adminRoute(handlers.AdminListOrders))
	mux.Handle("PUT /api/v1/admin/orders/{id}/status", adminRoute(handlers.AdminUpdateOrderStatus))

	// ==========================================
	// GLOBAL MIDDLEWARES
	// ==========================================

	var handler http.Handler = mux
	handler = middleware.JSON(handler)
	handler = middleware.CORS(handler)
	handler = middleware.MetricsMiddleware(handler)
	handler = middleware.Logger(handler)

	return handler
}
