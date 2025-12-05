package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// AssetOptimizationMiddleware handles static asset optimization
type AssetOptimizationMiddleware struct {
	// Configuration
	EnableCompression bool
	EnableCaching     bool
	MaxAge            time.Duration

	// Statistics
	totalRequests      int64
	compressedRequests int64
	cachedResponses    int64
	bytesServed        int64
	bytesSaved         int64
}

// AssetConfig holds configuration for different asset types
type AssetConfig struct {
	ContentType string
	MaxAge      time.Duration
	Compress    bool
	Immutable   bool
}

// NewAssetOptimizationMiddleware creates a new asset optimization middleware
func NewAssetOptimizationMiddleware() *AssetOptimizationMiddleware {
	return &AssetOptimizationMiddleware{
		EnableCompression: true,
		EnableCaching:     true,
		MaxAge:            24 * time.Hour, // Default 24 hours
	}
}

// getAssetConfig returns configuration based on file extension
func (a *AssetOptimizationMiddleware) getAssetConfig(ext string) AssetConfig {
	switch strings.ToLower(ext) {
	case ".css":
		return AssetConfig{
			ContentType: "text/css; charset=utf-8",
			MaxAge:      7 * 24 * time.Hour, // 1 week
			Compress:    true,
			Immutable:   false,
		}
	case ".js":
		return AssetConfig{
			ContentType: "application/javascript; charset=utf-8",
			MaxAge:      7 * 24 * time.Hour, // 1 week
			Compress:    true,
			Immutable:   false,
		}
	case ".png", ".jpg", ".jpeg", ".gif":
		return AssetConfig{
			ContentType: "image/" + strings.TrimPrefix(ext, "."),
			MaxAge:      30 * 24 * time.Hour, // 1 month
			Compress:    false,               // Images are already compressed
			Immutable:   true,
		}
	case ".ico":
		return AssetConfig{
			ContentType: "image/x-icon",
			MaxAge:      365 * 24 * time.Hour, // 1 year
			Compress:    false,
			Immutable:   true,
		}
	case ".woff", ".woff2":
		return AssetConfig{
			ContentType: "font/" + strings.TrimPrefix(ext, "."),
			MaxAge:      365 * 24 * time.Hour, // 1 year
			Compress:    false,                // Fonts are already compressed
			Immutable:   true,
		}
	case ".svg":
		return AssetConfig{
			ContentType: "image/svg+xml; charset=utf-8",
			MaxAge:      30 * 24 * time.Hour, // 1 month
			Compress:    true,
			Immutable:   false,
		}
	default:
		return AssetConfig{
			ContentType: "application/octet-stream",
			MaxAge:      24 * time.Hour, // 1 day
			Compress:    true,
			Immutable:   false,
		}
	}
}

// Middleware returns the asset optimization middleware handler
func (a *AssetOptimizationMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.totalRequests++

		// Only handle GET requests for static assets
		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}

		// Get file extension
		ext := filepath.Ext(r.URL.Path)
		if ext == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Get asset configuration
		config := a.getAssetConfig(ext)

		// Set content type
		w.Header().Set("Content-Type", config.ContentType)

		// Set cache headers if caching is enabled
		if a.EnableCaching {
			a.setCacheHeaders(w, config)
		}

		// Set security headers for assets
		a.setSecurityHeaders(w)

		// Handle compression if enabled and supported
		if a.EnableCompression && config.Compress && a.supportsCompression(r) {
			a.serveCompressed(w, r, next)
			return
		}

		// Serve normally
		next.ServeHTTP(w, r)
	})
}

// setCacheHeaders sets appropriate cache headers
func (a *AssetOptimizationMiddleware) setCacheHeaders(w http.ResponseWriter, config AssetConfig) {
	// Set Cache-Control header
	cacheControl := "public, max-age=" + string(rune(int(config.MaxAge.Seconds())))
	if config.Immutable {
		cacheControl += ", immutable"
	}
	w.Header().Set("Cache-Control", cacheControl)

	// Set Expires header
	expires := time.Now().Add(config.MaxAge)
	w.Header().Set("Expires", expires.Format(http.TimeFormat))

	// Set Last-Modified header (current time for generated content)
	w.Header().Set("Last-Modified", time.Now().Format(http.TimeFormat))

	// Set ETag for better caching
	etag := `"` + strings.ToLower(filepath.Base(w.Header().Get("Content-Type"))) + `-` +
		strings.ToLower(string(rune(time.Now().Unix()))) + `"`
	w.Header().Set("ETag", etag)

	// Enable conditional requests
	w.Header().Set("Vary", "Accept-Encoding")

	a.cachedResponses++
}

// setSecurityHeaders sets security headers for static assets
func (a *AssetOptimizationMiddleware) setSecurityHeaders(w http.ResponseWriter) {
	// Prevent MIME type sniffing
	w.Header().Set("X-Content-Type-Options", "nosniff")

	// Set referrer policy for privacy
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

	// Add CORS headers for cross-origin requests (if needed)
	origin := os.Getenv("CORS_ALLOWED_ORIGIN")
	if origin == "" {
		origin = "http://localhost:8080" // Safe fallback for development
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours
}

// supportsCompression checks if client supports gzip compression
func (a *AssetOptimizationMiddleware) supportsCompression(r *http.Request) bool {
	acceptEncoding := r.Header.Get("Accept-Encoding")
	return strings.Contains(acceptEncoding, "gzip")
}

// serveCompressed serves content with gzip compression
func (a *AssetOptimizationMiddleware) serveCompressed(w http.ResponseWriter, r *http.Request, next http.Handler) {
	// Set compression headers
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Vary", "Accept-Encoding")

	// Create gzip writer
	gz := gzip.NewWriter(w)
	defer gz.Close()

	// Create a custom response writer that writes to gzip
	gzw := &gzipResponseWriter{
		ResponseWriter: w,
		Writer:         gz,
	}

	// Serve content through gzip writer
	next.ServeHTTP(gzw, r)

	a.compressedRequests++
}

// gzipResponseWriter wraps http.ResponseWriter to write compressed content
type gzipResponseWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	return g.Writer.Write(b)
}

// GetStats returns middleware statistics
func (a *AssetOptimizationMiddleware) GetStats() map[string]interface{} {
	compressionRatio := float64(0)
	if a.totalRequests > 0 {
		compressionRatio = float64(a.compressedRequests) / float64(a.totalRequests) * 100
	}

	cachingRatio := float64(0)
	if a.totalRequests > 0 {
		cachingRatio = float64(a.cachedResponses) / float64(a.totalRequests) * 100
	}

	return map[string]interface{}{
		"total_requests":      a.totalRequests,
		"compressed_requests": a.compressedRequests,
		"cached_responses":    a.cachedResponses,
		"compression_ratio":   compressionRatio,
		"caching_ratio":       cachingRatio,
		"bytes_served":        a.bytesServed,
		"bytes_saved":         a.bytesSaved,
		"enabled_features": map[string]bool{
			"compression": a.EnableCompression,
			"caching":     a.EnableCaching,
		},
	}
}

// OptimizedStaticFileServer returns an optimized static file server
func OptimizedStaticFileServer(root http.FileSystem) http.Handler {
	fileServer := http.FileServer(root)
	assetMiddleware := NewAssetOptimizationMiddleware()

	return assetMiddleware.Middleware(fileServer)
}

// AssetOptimizationConfig holds global configuration
type AssetOptimizationConfig struct {
	EnableCompression bool
	EnableCaching     bool
	DefaultMaxAge     time.Duration
	CDNEnabled        bool
	CDNDomain         string
}

// DefaultAssetOptimizationConfig returns production-ready defaults
func DefaultAssetOptimizationConfig() AssetOptimizationConfig {
	return AssetOptimizationConfig{
		EnableCompression: true,
		EnableCaching:     true,
		DefaultMaxAge:     24 * time.Hour,
		CDNEnabled:        false,
		CDNDomain:         "",
	}
}
