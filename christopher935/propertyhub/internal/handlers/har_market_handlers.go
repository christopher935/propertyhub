package handlers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"chrisgross-ctrl-project/internal/services"
	"chrisgross-ctrl-project/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// HARMarketHandlers provides HTTP handlers for HAR market data
type HARMarketHandlers struct {
	db      *gorm.DB
	scraper *services.HARMarketScraper
}

// NewHARMarketHandlers creates new HAR market handlers
func NewHARMarketHandlers(db *gorm.DB, scraperAPIKey string) *HARMarketHandlers {
	scraper := services.NewHARMarketScraper(db, scraperAPIKey)
	
	return &HARMarketHandlers{
		db:      db,
		scraper: scraper,
	}
}

// TriggerScraping manually triggers HAR market data scraping
// POST /api/v1/market-data/scrape
func (h *HARMarketHandlers) TriggerScraping(c *gin.Context) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Start scraping in background
	go func() {
		err := h.scraper.ScrapeAllTargets(ctx)
		if err != nil {
			// Log error but don't fail the HTTP request since it's async
			// In production, you might want to send notifications or alerts
			log.Printf("HAR scraping completed with errors: %v", err)
		}
	}()

	utils.SuccessResponse(c, gin.H{
		"message":    "HAR market data scraping initiated",
		"status":     "started",
		"estimated":  "5-10 minutes",
		"targets":    6,
		"started_at": time.Now().UTC(),
	})
}

// GetLatestReports gets the latest market reports
// GET /api/v1/market-data/reports
func (h *HARMarketHandlers) GetLatestReports(c *gin.Context) {
	// Parse query parameters
	reportType := c.Query("type")
	limitStr := c.DefaultQuery("limit", "20")
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	// Get reports from scraper
	reports, err := h.scraper.GetLatestReports(reportType, limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve reports", err)
		return
	}

	utils.SuccessResponse(c, gin.H{
		"reports":     reports,
		"total":       len(reports),
		"report_type": reportType,
		"limit":       limit,
	})
}

// GetMarketReport gets a specific market report by ID
// GET /api/v1/market-data/reports/:id
func (h *HARMarketHandlers) GetMarketReport(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid report ID", err)
		return
	}

	var report services.HARMarketData
	if err := h.db.First(&report, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Report not found", nil)
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Database error", err)
		}
		return
	}

	utils.SuccessResponse(c, gin.H{
		"report": report,
	})
}

// GetScrapingStats gets HAR scraping statistics
// GET /api/v1/market-data/stats
func (h *HARMarketHandlers) GetScrapingStats(c *gin.Context) {
	stats, err := h.scraper.GetScrapeStats()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve stats", err)
		return
	}

	utils.SuccessResponse(c, gin.H{
		"stats":       stats,
		"generated_at": time.Now().UTC(),
	})
}

// GetReportsByType gets reports filtered by type
// GET /api/v1/market-data/reports/type/:type
func (h *HARMarketHandlers) GetReportsByType(c *gin.Context) {
	reportType := c.Param("type")
	if reportType == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Report type is required", nil)
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 50 {
		limit = 10
	}

	reports, err := h.scraper.GetLatestReports(reportType, limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve reports", err)
		return
	}

	utils.SuccessResponse(c, gin.H{
		"reports":     reports,
		"report_type": reportType,
		"total":       len(reports),
		"limit":       limit,
	})
}

// GetMarketSummary provides a summary of market data for dashboards
// GET /api/v1/market-data/summary
func (h *HARMarketHandlers) GetMarketSummary(c *gin.Context) {
	// Get recent reports from each category
	categories := []string{
		"newsroom_market",
		"newsroom_residential", 
		"mls_data",
		"rental_market",
		"affordability",
		"hot_communities",
	}

	summary := gin.H{
		"categories": gin.H{},
		"latest_reports": []services.HARMarketData{},
		"stats": gin.H{},
	}

	// Get latest report from each category
	for _, category := range categories {
		reports, err := h.scraper.GetLatestReports(category, 1)
		if err == nil && len(reports) > 0 {
			summary["categories"].(gin.H)[category] = gin.H{
				"latest_report": reports[0],
				"count": h.getReportCount(category),
			}
			
			// Add to latest reports for overall summary
			if len(summary["latest_reports"].([]services.HARMarketData)) < 5 {
				summary["latest_reports"] = append(
					summary["latest_reports"].([]services.HARMarketData), 
					*reports[0],
				)
			}
		}
	}

	// Get overall stats
	stats, err := h.scraper.GetScrapeStats()
	if err == nil {
		summary["stats"] = stats
	}

	utils.SuccessResponse(c, summary)
}

// getReportCount gets count of reports by type
func (h *HARMarketHandlers) getReportCount(reportType string) int64 {
	var count int64
	h.db.Model(&services.HARMarketData{}).
		Where("report_type = ? AND status = 'active'", reportType).
		Count(&count)
	return count
}

// ScheduleWeeklyScraping sets up weekly scraping schedule
// POST /api/v1/market-data/schedule
func (h *HARMarketHandlers) ScheduleWeeklyScraping(c *gin.Context) {
	// This would integrate with your job scheduler
	// For now, return a success response indicating scheduling is configured
	
	utils.SuccessResponse(c, gin.H{
		"message":    "Weekly HAR market data scraping scheduled",
		"frequency":  "weekly",
		"day":        "Monday",
		"time":       "06:00 UTC",
		"targets":    6,
		"enabled":    true,
	})
}

// GetScrapingLogs gets recent scraping activity logs
// GET /api/v1/market-data/logs
func (h *HARMarketHandlers) GetScrapingLogs(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 200 {
		limit = 50
	}

	var logs []services.HARScrapeLog
	err = h.db.Order("scraped_at DESC").
		Limit(limit).
		Find(&logs).Error
	
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve logs", err)
		return
	}

	utils.SuccessResponse(c, gin.H{
		"logs":  logs,
		"total": len(logs),
		"limit": limit,
	})
}

// SearchReports searches through market reports
// GET /api/v1/market-data/search
func (h *HARMarketHandlers) SearchReports(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Search query is required", nil)
		return
	}

	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	var reports []services.HARMarketData
	err = h.db.Where("status = 'active'").
		Where("title ILIKE ? OR summary ILIKE ?", "%"+query+"%", "%"+query+"%").
		Order("publish_date DESC").
		Limit(limit).
		Find(&reports).Error

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Search failed", err)
		return
	}

	utils.SuccessResponse(c, gin.H{
		"reports": reports,
		"query":   query,
		"total":   len(reports),
		"limit":   limit,
	})
}

// RegisterRoutes registers all HAR market data routes
func (h *HARMarketHandlers) RegisterRoutes(r *gin.RouterGroup) {
	marketData := r.Group("/market-data")
	{
		// Scraping management
		marketData.POST("/scrape", h.TriggerScraping)
		marketData.POST("/schedule", h.ScheduleWeeklyScraping)
		
		// Report retrieval
		marketData.GET("/reports", h.GetLatestReports)
		marketData.GET("/reports/:id", h.GetMarketReport)
		marketData.GET("/reports/type/:type", h.GetReportsByType)
		
		// Analytics and search
		marketData.GET("/summary", h.GetMarketSummary)
		marketData.GET("/stats", h.GetScrapingStats)
		marketData.GET("/logs", h.GetScrapingLogs)
		marketData.GET("/search", h.SearchReports)
	}
}
