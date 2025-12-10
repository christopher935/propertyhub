package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ChartHandlers struct {
	DB *gorm.DB
}

func NewChartHandlers(db *gorm.DB) *ChartHandlers {
	return &ChartHandlers{DB: db}
}

type DailyBooking struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type SourceCount struct {
	Source string `json:"source"`
	Count  int    `json:"count"`
}

func (h *ChartHandlers) GetBookingTrendsChart(c *gin.Context) {
	var results []DailyBooking
	h.DB.Raw(`
		SELECT DATE(created_at) as date, COUNT(*) as count 
		FROM bookings 
		WHERE created_at >= ? 
		GROUP BY DATE(created_at) 
		ORDER BY date
	`, time.Now().AddDate(0, 0, -30)).Scan(&results)

	c.JSON(http.StatusOK, gin.H{
		"labels": extractDates(results),
		"data":   extractCounts(results),
	})
}

func (h *ChartHandlers) GetLeadSourcesChart(c *gin.Context) {
	var results []SourceCount
	h.DB.Raw(`
		SELECT COALESCE(source, 'Direct') as source, COUNT(*) as count 
		FROM leads 
		GROUP BY source
	`).Scan(&results)

	c.JSON(http.StatusOK, gin.H{
		"labels": extractSources(results),
		"data":   extractSourceCounts(results),
	})
}

func extractDates(results []DailyBooking) []string {
	dates := make([]string, len(results))
	for i, r := range results {
		dates[i] = r.Date
	}
	return dates
}

func extractCounts(results []DailyBooking) []int {
	counts := make([]int, len(results))
	for i, r := range results {
		counts[i] = r.Count
	}
	return counts
}

func extractSources(results []SourceCount) []string {
	sources := make([]string, len(results))
	for i, r := range results {
		sources[i] = r.Source
	}
	return sources
}

func extractSourceCounts(results []SourceCount) []int {
	counts := make([]int, len(results))
	for i, r := range results {
		counts[i] = r.Count
	}
	return counts
}
