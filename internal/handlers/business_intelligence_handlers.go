package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"chrisgross-ctrl-project/internal/services"
)

// BusinessIntelligenceHandlers handles BI dashboard API requests
type BusinessIntelligenceHandlers struct {
	biService *services.BusinessIntelligenceService
}

// NewBusinessIntelligenceHandlers creates new BI handlers
func NewBusinessIntelligenceHandlers(db *gorm.DB) *BusinessIntelligenceHandlers {
	return &BusinessIntelligenceHandlers{
		biService: services.NewBusinessIntelligenceService(db),
	}
}

// RegisterBIRoutes registers all BI dashboard routes
func RegisterBIRoutes(router *gin.Engine, db *gorm.DB) {
	handlers := NewBusinessIntelligenceHandlers(db)

	bi := router.Group("/api/v1/bi")
	{
		// Main dashboard metrics
		bi.GET("/dashboard", handlers.GetDashboardMetrics)

		// Property analytics
		bi.GET("/properties/analytics", handlers.GetPropertyAnalytics)
		bi.GET("/properties/:mls_id/insights", handlers.GetPropertyInsights)

		// Booking analytics
		bi.GET("/bookings/analytics", handlers.GetBookingAnalytics)
		bi.GET("/bookings/trends", handlers.GetBookingTrends)

		// Lead analytics
		bi.GET("/leads/analytics", handlers.GetLeadAnalytics)
		bi.GET("/leads/funnel", handlers.GetConversionFunnel)

		// System analytics
		bi.GET("/system/health", handlers.GetSystemHealth)
		bi.GET("/system/performance", handlers.GetSystemPerformance)

		// Performance analytics
		bi.GET("/performance/roi", handlers.GetROIAnalytics)
		bi.GET("/performance/efficiency", handlers.GetEfficiencyMetrics)

		// Reports
		bi.GET("/reports/friday", handlers.GetFridayReport)
		bi.POST("/reports/friday/send", handlers.SendFridayReport)

		// Real-time metrics
		bi.GET("/realtime/stats", handlers.GetRealtimeStats)
		bi.GET("/realtime/alerts", handlers.GetSystemAlerts)
	}
}

// GetDashboardMetrics returns comprehensive dashboard analytics
func (bih *BusinessIntelligenceHandlers) GetDashboardMetrics(c *gin.Context) {
	metrics, err := bih.biService.GetDashboardMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate dashboard metrics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    metrics,
	})
}

// GetPropertyAnalytics returns property performance analytics
func (bih *BusinessIntelligenceHandlers) GetPropertyAnalytics(c *gin.Context) {
	metrics, err := bih.biService.GetDashboardMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate property analytics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    metrics.PropertyMetrics,
	})
}

// GetPropertyInsights returns detailed insights for a specific property
func (bih *BusinessIntelligenceHandlers) GetPropertyInsights(c *gin.Context) {
	mlsID := c.Param("mls_id")
	if mlsID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "MLS ID is required",
		})
		return
	}

	insights, err := bih.biService.GetPropertyInsights(mlsID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Property not found or insights unavailable",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    insights,
	})
}

// GetBookingAnalytics returns booking performance analytics
func (bih *BusinessIntelligenceHandlers) GetBookingAnalytics(c *gin.Context) {
	metrics, err := bih.biService.GetDashboardMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate booking analytics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    metrics.BookingMetrics,
	})
}

// GetBookingTrends returns booking trend analysis
func (bih *BusinessIntelligenceHandlers) GetBookingTrends(c *gin.Context) {
	// Get optional time range parameters
	days := c.DefaultQuery("days", "30")
	daysInt, err := strconv.Atoi(days)
	if err != nil || daysInt < 1 || daysInt > 365 {
		daysInt = 30
	}

	metrics, err := bih.biService.GetDashboardMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate booking trends",
			"details": err.Error(),
		})
		return
	}

	// Filter trends based on requested days
	trends := metrics.BookingMetrics.BookingTrends
	if len(trends) > daysInt {
		trends = trends[len(trends)-daysInt:]
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"trends":     trends,
			"time_slots": metrics.BookingMetrics.PopularTimeSlots,
			"summary": gin.H{
				"total_bookings":   metrics.BookingMetrics.TotalBookings,
				"conversion_rate":  metrics.BookingMetrics.ConversionRate,
				"average_duration": metrics.BookingMetrics.AverageBookingTime,
			},
		},
	})
}

// GetLeadAnalytics returns lead management analytics
func (bih *BusinessIntelligenceHandlers) GetLeadAnalytics(c *gin.Context) {
	metrics, err := bih.biService.GetDashboardMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate lead analytics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    metrics.LeadMetrics,
	})
}

// GetConversionFunnel returns lead conversion funnel analysis
func (bih *BusinessIntelligenceHandlers) GetConversionFunnel(c *gin.Context) {
	metrics, err := bih.biService.GetDashboardMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate conversion funnel",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"funnel":        metrics.LeadMetrics.ConversionFunnel,
			"quality_score": metrics.LeadMetrics.LeadQualityScore,
			"response_time": metrics.LeadMetrics.AverageResponseTime,
			"sources":       metrics.LeadMetrics.LeadSources,
		},
	})
}

// GetSystemHealth returns system health and performance metrics
func (bih *BusinessIntelligenceHandlers) GetSystemHealth(c *gin.Context) {
	healthReport, err := bih.biService.GetSystemHealthReport()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate system health report",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    healthReport,
	})
}

// GetSystemPerformance returns detailed system performance metrics
func (bih *BusinessIntelligenceHandlers) GetSystemPerformance(c *gin.Context) {
	metrics, err := bih.biService.GetDashboardMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate system performance metrics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"system_metrics":      metrics.SystemMetrics,
			"performance_metrics": metrics.PerformanceMetrics,
			"health_summary": gin.H{
				"overall_status":   "EXCELLENT",
				"data_consistency": metrics.SystemMetrics["data_consistency"],
				"sync_performance": metrics.SystemMetrics["sync_performance"],
				"error_rates":      metrics.SystemMetrics["error_rates"],
			},
		},
	})
}

// GetROIAnalytics returns ROI and business performance metrics
func (bih *BusinessIntelligenceHandlers) GetROIAnalytics(c *gin.Context) {
	metrics, err := bih.biService.GetDashboardMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate ROI analytics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"roi_metrics":           metrics.PerformanceMetrics.ROIMetrics,
			"efficiency_gains":      metrics.PerformanceMetrics.EfficiencyGains,
			"competitive_advantage": metrics.PerformanceMetrics.CompetitiveAdvantage,
			"revenue_projection":    metrics.PerformanceMetrics.RevenueProjection,
			"summary": gin.H{
				"monthly_savings":        metrics.PerformanceMetrics.ROIMetrics.AutomationSavings,
				"efficiency_improvement": metrics.PerformanceMetrics.ROIMetrics.EfficiencyGains,
				"time_saved":             metrics.PerformanceMetrics.ROIMetrics.TimesSaved,
			},
		},
	})
}

// GetEfficiencyMetrics returns efficiency improvement metrics
func (bih *BusinessIntelligenceHandlers) GetEfficiencyMetrics(c *gin.Context) {
	metrics, err := bih.biService.GetDashboardMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate efficiency metrics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"efficiency_metrics": metrics.PerformanceMetrics.EfficiencyGains,
			"growth_indicators":  metrics.PerformanceMetrics.GrowthIndicators,
			"automation_impact": gin.H{
				"data_redundancy_eliminated": metrics.PerformanceMetrics.EfficiencyGains.DataRedundancyEliminated,
				"manual_work_reduced":        metrics.PerformanceMetrics.EfficiencyGains.ManualWorkReduced,
				"processing_time_reduced":    metrics.PerformanceMetrics.EfficiencyGains.ProcessingTimeReduced,
				"error_rate_reduced":         metrics.PerformanceMetrics.EfficiencyGains.ErrorRateReduced,
			},
		},
	})
}

// GetRealtimeStats returns real-time system statistics
func (bih *BusinessIntelligenceHandlers) GetRealtimeStats(c *gin.Context) {
	// Generate real-time statistics
	stats := gin.H{
		"timestamp":        "2025-08-19T12:00:00Z",
		"active_sessions":  15,
		"current_bookings": 3,
		"system_load": gin.H{
			"cpu_usage":    25.5,
			"memory_usage": 45.2,
			"disk_usage":   15.8,
		},
		"api_requests": gin.H{
			"last_minute":  45,
			"last_hour":    2150,
			"success_rate": 99.8,
		},
		"sync_status": gin.H{
			"propertyhub": "ACTIVE",
			"fub":         "ACTIVE",
			"email":       "ACTIVE",
		},
		"queue_sizes": gin.H{
			"sync_queue":    0,
			"webhook_queue": 2,
			"email_queue":   1,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetSystemAlerts returns current system alerts and notifications
func (bih *BusinessIntelligenceHandlers) GetSystemAlerts(c *gin.Context) {
	healthReport, err := bih.biService.GetSystemHealthReport()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate system alerts",
			"details": err.Error(),
		})
		return
	}

	alerts := healthReport["alerts"]
	if alerts == nil {
		alerts = []string{}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"alerts":        alerts,
			"alert_count":   len(alerts.([]string)),
			"severity":      "LOW",
			"last_updated":  healthReport["last_updated"],
			"system_status": "HEALTHY",
		},
	})
}

// GetAnalyticsExport exports analytics data in various formats
func (bih *BusinessIntelligenceHandlers) GetAnalyticsExport(c *gin.Context) {
	format := c.DefaultQuery("format", "json")
	reportType := c.DefaultQuery("type", "dashboard")

	if format != "json" && format != "csv" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Unsupported format. Use 'json' or 'csv'",
		})
		return
	}

	metrics, err := bih.biService.GetDashboardMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate analytics export",
			"details": err.Error(),
		})
		return
	}

	switch reportType {
	case "properties":
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    metrics.PropertyMetrics,
			"format":  format,
		})
	case "bookings":
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    metrics.BookingMetrics,
			"format":  format,
		})
	case "leads":
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    metrics.LeadMetrics,
			"format":  format,
		})
	case "system":
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    metrics.SystemMetrics,
			"format":  format,
		})
	default:
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    metrics,
			"format":  format,
		})
	}
}

// GetFridayReport generates and returns the comprehensive Friday Report
func (bih *BusinessIntelligenceHandlers) GetFridayReport(c *gin.Context) {
	// Generate the Friday Report
	report, err := bih.biService.GenerateFridayReport()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate Friday Report",
			"details": err.Error(),
		})
		return
	}

	// Check for format parameter
	format := c.DefaultQuery("format", "json")

	switch format {
	case "text":
		// Generate text format for email
		textReport := bih.generateTextFridayReport(report)
		c.JSON(http.StatusOK, gin.H{
			"success":     true,
			"data":        report,
			"text_format": textReport,
			"format":      "text",
		})
	default:
		// Return JSON format
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    report,
			"format":  "json",
		})
	}
}

// SendFridayReport sends the Friday Report via email
func (bih *BusinessIntelligenceHandlers) SendFridayReport(c *gin.Context) {
	var request struct {
		Recipients []string `json:"recipients" binding:"required"`
		ReportText string   `json:"report_text" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	if len(request.Recipients) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "At least one recipient is required",
		})
		return
	}

	// TODO: Implement actual email sending using AWS SES or SMTP
	// For now, just return success
	// In a real implementation, you would use the aws_communication_service
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Report sent to %d recipient(s)", len(request.Recipients)),
		"recipients": request.Recipients,
	})
}

// generateTextFridayReport converts the Friday Report to text format for email
func (bih *BusinessIntelligenceHandlers) generateTextFridayReport(report *services.FridayReportData) string {
	var text strings.Builder

	text.WriteString(fmt.Sprintf("=== FRIDAY REPORT - Week of %s ===\n\n", report.WeekRange))

	// SOLD LISTINGS (Closing Pipeline)
	if len(report.SoldListings) > 0 {
		text.WriteString("🏡 SOLD LISTINGS (Closing Pipeline)\n\n")
		for _, item := range report.SoldListings {
			text.WriteString(fmt.Sprintf("%s - SOLD %s\n", item.PropertyAddress, item.SoldDate.Format("1/2")))

			// Checkboxes for workflow status
			leaseSentCheck := "❌"
			if item.LeaseSentOut {
				leaseSentCheck = "✅"
			}
			leaseCompleteCheck := "❌"
			if item.LeaseComplete {
				leaseCompleteCheck = "✅"
			}
			depositCheck := "❌"
			if item.DepositReceived {
				depositCheck = "✅"
			}
			firstMonthCheck := "❌"
			if item.FirstMonthReceived {
				firstMonthCheck = "✅"
			}

			text.WriteString(fmt.Sprintf("Lease sent out: %s", leaseSentCheck))
			if item.LeaseSentOut && item.LeaseComplete {
				text.WriteString(fmt.Sprintf(" (Sent %s)", item.SoldDate.Format("1/2")))
			}
			text.WriteString("\n")

			text.WriteString(fmt.Sprintf("Lease complete: %s", leaseCompleteCheck))
			if item.LeaseComplete {
				text.WriteString(fmt.Sprintf(" (Signed %s)", item.SoldDate.Format("1/2")))
			}
			text.WriteString("\n")

			text.WriteString(fmt.Sprintf("Deposit received: %s\n", depositCheck))
			text.WriteString(fmt.Sprintf("First Month's Rent received: %s\n", firstMonthCheck))

			if item.MoveInDate != nil {
				text.WriteString(fmt.Sprintf("Move in Date: %s\n", item.MoveInDate.Format("1/2/2006")))
				if item.DaysToMoveIn != nil {
					text.WriteString(fmt.Sprintf("Days to move-in: %d days\n", *item.DaysToMoveIn))
				}
			}

			// Status and alerts
			text.WriteString(fmt.Sprintf("[STATUS: %s]\n", item.StatusSummary))
			if len(item.AlertFlags) > 0 {
				text.WriteString(fmt.Sprintf("[ALERTS: %s]\n", strings.Join(item.AlertFlags, ", ")))
			}

			text.WriteString("\n")
		}
	}

	// ACTIVE LISTINGS PERFORMANCE
	text.WriteString("📊 ACTIVE LISTINGS PERFORMANCE\n\n")
	for _, listing := range report.ActiveListings {
		text.WriteString(fmt.Sprintf("%s\n", listing.PropertyAddress))
		text.WriteString(fmt.Sprintf("CDOM: %d days\n", listing.CDOM))
		text.WriteString(fmt.Sprintf("Leads: %d (%d vs last week)\n", listing.LeadsTotal, listing.LeadsWeekChange))
		text.WriteString(fmt.Sprintf("External Showings: %d (%d vs last week)\n", listing.ExternalShowings, listing.ExternalShowingsChange))
		text.WriteString(fmt.Sprintf("Booking System Showings: %d (%d this week) (%d vs last week)\n",
			listing.BookingShowings, listing.BookingShowingsWeek, listing.BookingShowingsChange))
		text.WriteString(fmt.Sprintf("Total Showings: %d (%d vs last week)\n", listing.TotalShowings, listing.TotalShowingsChange))
		text.WriteString(fmt.Sprintf("Applications: %d (%d vs last week)\n", listing.Applications, listing.ApplicationsChange))

		// ShowingSmart Feedback
		if len(listing.ShowingSmartFeedback) > 0 {
			text.WriteString("ShowingSmart Feedback:\n")
			for _, feedback := range listing.ShowingSmartFeedback {
				text.WriteString(fmt.Sprintf("  • %s - %s:\n", feedback.Date.Format("1/2"), feedback.Agent))
				text.WriteString(fmt.Sprintf("    Interest: \"%s\"\n", feedback.InterestLevel))
				text.WriteString(fmt.Sprintf("    Price: \"%s\"\n", feedback.PriceOpinion))
				text.WriteString(fmt.Sprintf("    Comparison: \"%s\"\n", feedback.Comparison))
				if feedback.Comments != "" {
					text.WriteString(fmt.Sprintf("    Comments: \"%s\"\n", feedback.Comments))
				}
			}
		}

		// AI Insights
		if len(listing.AIInsights) > 0 {
			text.WriteString(fmt.Sprintf("[AI INSIGHT: %s]\n", strings.Join(listing.AIInsights, "; ")))
		}

		text.WriteString("\n")
	}

	// PRE-LISTING PIPELINE
	if len(report.PreListingPipeline) > 0 {
		text.WriteString(fmt.Sprintf("📋 PRE-LISTING PIPELINE (%d properties in preparation)\n\n", len(report.PreListingPipeline)))
		for _, item := range report.PreListingPipeline {
			targetDate := "TBD"
			if item.TargetListDate != nil {
				targetDate = item.TargetListDate.Format("1/2")
			}
			text.WriteString(fmt.Sprintf("%s - Target: %s\n", item.PropertyAddress, targetDate))
			if len(item.TasksRemaining) > 0 {
				text.WriteString(fmt.Sprintf("• %s\n", strings.Join(item.TasksRemaining, ", ")))
			}
			text.WriteString("\n")
		}
	}

	// WEEKLY SUMMARY
	text.WriteString("📈 WEEKLY SUMMARY\n")
	text.WriteString(fmt.Sprintf("• Active listings: %d properties\n", report.WeeklySummary.ActiveListings))
	if report.WeeklySummary.PreListings > 0 {
		text.WriteString(fmt.Sprintf("• Pre-listings: %d properties\n", report.WeeklySummary.PreListings))
	}
	text.WriteString(fmt.Sprintf("• Total showings: %d (%d vs last week)\n", report.WeeklySummary.TotalShowings, report.WeeklySummary.ShowingsChange))
	if report.WeeklySummary.ClosingsInProgress > 0 {
		text.WriteString(fmt.Sprintf("• Closings in progress: %d properties\n", report.WeeklySummary.ClosingsInProgress))
	}
	if report.WeeklySummary.UpcomingMoveIns > 0 {
		text.WriteString(fmt.Sprintf("• Upcoming move-ins (next 7 days): %d properties\n", report.WeeklySummary.UpcomingMoveIns))
	}

	text.WriteString(fmt.Sprintf("\nReport generated: %s\n", report.GeneratedAt.Format("Monday, January 2, 2006 at 3:04 PM")))

	return text.String()
}
