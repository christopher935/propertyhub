package services

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

// HARMarketData represents scraped HAR market data
type HARMarketData struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ReportType  string    `gorm:"index;not null" json:"report_type"` // "newsroom", "mls", "rental", "affordability", "communities"
	Title       string    `gorm:"not null" json:"title"`
	URL         string    `gorm:"unique;not null" json:"url"`
	PublishDate time.Time `gorm:"index" json:"publish_date"`
	Content     string    `gorm:"type:text" json:"content"`
	Summary     string    `gorm:"type:text" json:"summary"`
	KeyMetrics  string    `gorm:"type:json" json:"key_metrics"` // JSON encoded metrics
	Status      string    `gorm:"default:'active'" json:"status"`
	ScrapedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"scraped_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// HARScrapeLog represents scraping activity logs
type HARScrapeLog struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	ScrapeType   string    `gorm:"not null" json:"scrape_type"`
	URL          string    `gorm:"not null" json:"url"`
	Status       string    `gorm:"not null" json:"status"` // "success", "failed", "partial"
	ItemsFound   int       `json:"items_found"`
	ItemsNew     int       `json:"items_new"`
	ItemsUpdated int       `json:"items_updated"`
	ErrorMessage string    `gorm:"type:text" json:"error_message,omitempty"`
	Duration     int       `json:"duration_ms"` // in milliseconds
	ScrapedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"scraped_at"`
}

// HARMarketScraper handles scraping HAR market data
type HARMarketScraper struct {
	db            *gorm.DB
	scraperAPIKey string
	httpClient    *http.Client
	mutex         sync.RWMutex
	
	// Configuration
	maxRetries    int
	retryDelay    time.Duration
	requestDelay  time.Duration
	userAgent     string
	
	// Scraping targets
	targets map[string]string
}

// ScraperAPIResponse represents response from ScraperAPI
type ScraperAPIResponse struct {
	Success bool   `json:"success"`
	Data    string `json:"data"`
	Error   string `json:"error,omitempty"`
}

// NewHARMarketScraper creates a new HAR market data scraper
func NewHARMarketScraper(db *gorm.DB, scraperAPIKey string) *HARMarketScraper {
	// Define HAR scraping targets
	targets := map[string]string{
		"newsroom_market":      "https://www.har.com/content/department/newsroom?pid=2005",
		"newsroom_residential": "https://www.har.com/content/department/newsroom?pid=2006", 
		"mls_data":            "https://www.har.com/content/department/mls",
		"rental_market":       "https://www.har.com/content/department/rental_market_update",
		"affordability":       "https://www.har.com/content/department/housing_affordability",
		"hot_communities":     "https://www.har.com/content/department/hottest_communities",
	}

	return &HARMarketScraper{
		db:            db,
		scraperAPIKey: scraperAPIKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		maxRetries:   3,
		retryDelay:   2 * time.Second,
		requestDelay: 1 * time.Second,
		userAgent:    "Mozilla/5.0 (compatible; PropertyHub/1.0; +https://chrisgross-ctrl-project.com)",
		targets:      targets,
	}
}

// ScrapeAllTargets scrapes all HAR market data targets
func (h *HARMarketScraper) ScrapeAllTargets(ctx context.Context) error {
	log.Println("🕷️  Starting comprehensive HAR market data scraping...")
	
	startTime := time.Now()
	var wg sync.WaitGroup
	results := make(chan ScrapeResult, len(h.targets))
	
	// Scrape all targets concurrently with rate limiting
	for reportType, targetURL := range h.targets {
		wg.Add(1)
		go func(rType, url string) {
			defer wg.Done()
			
			// Rate limiting between requests
			time.Sleep(h.requestDelay)
			
			result := h.scrapeTarget(ctx, rType, url)
			results <- result
		}(reportType, targetURL)
	}
	
	// Wait for all scraping to complete
	go func() {
		wg.Wait()
		close(results)
	}()
	
	// Process results
	totalFound := 0
	totalNew := 0
	totalUpdated := 0
	errors := []string{}
	
	for result := range results {
		totalFound += result.ItemsFound
		totalNew += result.ItemsNew
		totalUpdated += result.ItemsUpdated
		
		if result.Error != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", result.ReportType, result.Error))
		}
		
		// Log individual scrape result
		h.logScrapeResult(result)
	}
	
	duration := time.Since(startTime)
	log.Printf("✅ HAR scraping completed in %v: Found %d items, %d new, %d updated", 
		duration, totalFound, totalNew, totalUpdated)
	
	if len(errors) > 0 {
		log.Printf("⚠️  Scraping errors: %s", strings.Join(errors, "; "))
		return fmt.Errorf("partial scraping failure: %s", strings.Join(errors, "; "))
	}
	
	return nil
}

// ScrapeResult represents the result of a scraping operation
type ScrapeResult struct {
	ReportType   string
	URL          string
	ItemsFound   int
	ItemsNew     int
	ItemsUpdated int
	Duration     time.Duration
	Error        error
}

// scrapeTarget scrapes a specific HAR target URL
func (h *HARMarketScraper) scrapeTarget(ctx context.Context, reportType, targetURL string) ScrapeResult {
	startTime := time.Now()
	result := ScrapeResult{
		ReportType: reportType,
		URL:        targetURL,
	}
	
	log.Printf("📊 Scraping %s: %s", reportType, targetURL)
	
	// Get page content via ScraperAPI
	content, err := h.fetchWithScraperAPI(targetURL)
	if err != nil {
		result.Error = fmt.Errorf("failed to fetch %s: %v", targetURL, err)
		result.Duration = time.Since(startTime)
		return result
	}
	
	// Parse reports from the page
	reports, err := h.parseReportsFromHTML(content, reportType, targetURL)
	if err != nil {
		result.Error = fmt.Errorf("failed to parse %s: %v", targetURL, err)
		result.Duration = time.Since(startTime)
		return result
	}
	
	result.ItemsFound = len(reports)
	
	// Save reports to database
	for _, report := range reports {
		isNew, err := h.saveMarketData(report)
		if err != nil {
			log.Printf("⚠️  Failed to save report %s: %v", report.Title, err)
			continue
		}
		
		if isNew {
			result.ItemsNew++
		} else {
			result.ItemsUpdated++
		}
	}
	
	result.Duration = time.Since(startTime)
	return result
}

// fetchWithScraperAPI fetches content using ScraperAPI
func (h *HARMarketScraper) fetchWithScraperAPI(targetURL string) (string, error) {
	// ScraperAPI endpoint
	apiURL := "http://api.scraperapi.com"
	
	// Build request parameters
	params := url.Values{}
	params.Add("api_key", h.scraperAPIKey)
	params.Add("url", targetURL)
	params.Add("render", "false") // Don't need JavaScript rendering for HAR
	params.Add("country_code", "us")
	params.Add("premium", "true") // Use premium for better success rate
	
	fullURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())
	
	var lastErr error
	for attempt := 0; attempt < h.maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(h.retryDelay * time.Duration(attempt))
			log.Printf("🔄 Retry attempt %d for %s", attempt+1, targetURL)
		}
		
		req, err := http.NewRequest("GET", fullURL, nil)
		if err != nil {
			lastErr = err
			continue
		}
		
		req.Header.Set("User-Agent", h.userAgent)
		
		resp, err := h.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
			continue
		}
		
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			continue
		}
		
		return string(body), nil
	}
	
	return "", lastErr
}

// parseReportsFromHTML parses market reports from HTML content
func (h *HARMarketScraper) parseReportsFromHTML(html, reportType, sourceURL string) ([]*HARMarketData, error) {
	reports := []*HARMarketData{}
	
	// Different parsing strategies for different page types
	switch reportType {
	case "newsroom_market", "newsroom_residential":
		return h.parseNewsroomReports(html, reportType, sourceURL)
	case "mls_data":
		return h.parseMLSReports(html, reportType, sourceURL)
	case "rental_market":
		return h.parseRentalReports(html, reportType, sourceURL)
	case "affordability":
		return h.parseAffordabilityReports(html, reportType, sourceURL)
	case "hot_communities":
		return h.parseCommunitiesReports(html, reportType, sourceURL)
	default:
		return h.parseGenericReports(html, reportType, sourceURL)
	}
	
	return reports, nil
}

// parseNewsroomReports parses HAR newsroom press releases
func (h *HARMarketScraper) parseNewsroomReports(html, reportType, sourceURL string) ([]*HARMarketData, error) {
	reports := []*HARMarketData{}
	
	// Look for news article links and titles
	// HAR typically uses patterns like: <a href="/content/news/article/...">Title</a>
	linkPattern := regexp.MustCompile(`<a[^>]*href="([^"]*(?:news|article|report)[^"]*)"[^>]*>([^<]+)</a>`)
	matches := linkPattern.FindAllStringSubmatch(html, -1)
	
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		
		relativeURL := strings.TrimSpace(match[1])
		title := strings.TrimSpace(match[2])
		
		// Skip empty titles or URLs
		if title == "" || relativeURL == "" {
			continue
		}
		
		// Convert relative URL to absolute
		fullURL := h.resolveURL(sourceURL, relativeURL)
		
		// Try to extract date from title or URL
		publishDate := h.extractDateFromText(title)
		if publishDate.IsZero() {
			publishDate = time.Now().AddDate(0, 0, -1) // Default to yesterday
		}
		
		report := &HARMarketData{
			ReportType:  reportType,
			Title:       title,
			URL:         fullURL,
			PublishDate: publishDate,
			Content:     "", // Will be filled by detailed scraping
			Summary:     h.generateSummaryFromTitle(title),
			KeyMetrics:  "{}",
			Status:      "active",
		}
		
		reports = append(reports, report)
	}
	
	// Also look for direct report links
	reportPattern := regexp.MustCompile(`(?i)<a[^>]*href="([^"]*(?:report|market|housing)[^"]*\.(?:pdf|html)[^"]*)"[^>]*>([^<]+)</a>`)
	reportMatches := reportPattern.FindAllStringSubmatch(html, -1)
	
	for _, match := range reportMatches {
		if len(match) < 3 {
			continue
		}
		
		relativeURL := strings.TrimSpace(match[1])
		title := strings.TrimSpace(match[2])
		
		if title == "" || relativeURL == "" {
			continue
		}
		
		fullURL := h.resolveURL(sourceURL, relativeURL)
		publishDate := h.extractDateFromText(title)
		if publishDate.IsZero() {
			publishDate = time.Now().AddDate(0, 0, -1)
		}
		
		report := &HARMarketData{
			ReportType:  reportType + "_report",
			Title:       title,
			URL:         fullURL,
			PublishDate: publishDate,
			Content:     "",
			Summary:     h.generateSummaryFromTitle(title),
			KeyMetrics:  "{}",
			Status:      "active",
		}
		
		reports = append(reports, report)
	}
	
	return reports, nil
}

// parseMLSReports parses MLS data reports
func (h *HARMarketScraper) parseMLSReports(html, reportType, sourceURL string) ([]*HARMarketData, error) {
	reports := []*HARMarketData{}
	
	// Look for MLS statistics and reports
	mlsPattern := regexp.MustCompile(`(?i)<a[^>]*href="([^"]*(?:mls|statistic|data)[^"]*)"[^>]*>([^<]+)</a>`)
	matches := mlsPattern.FindAllStringSubmatch(html, -1)
	
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		
		relativeURL := strings.TrimSpace(match[1])
		title := strings.TrimSpace(match[2])
		
		if title == "" || relativeURL == "" {
			continue
		}
		
		fullURL := h.resolveURL(sourceURL, relativeURL)
		publishDate := h.extractDateFromText(title)
		if publishDate.IsZero() {
			publishDate = time.Now().AddDate(0, 0, -2) // Default to 2 days ago
		}
		
		report := &HARMarketData{
			ReportType:  reportType,
			Title:       title,
			URL:         fullURL,
			PublishDate: publishDate,
			Content:     "",
			Summary:     h.generateSummaryFromTitle(title),
			KeyMetrics:  "{}",
			Status:      "active",
		}
		
		reports = append(reports, report)
	}
	
	return reports, nil
}

// parseRentalReports parses rental market update reports
func (h *HARMarketScraper) parseRentalReports(html, reportType, sourceURL string) ([]*HARMarketData, error) {
	// Similar parsing logic for rental reports
	return h.parseGenericReports(html, reportType, sourceURL)
}

// parseAffordabilityReports parses housing affordability reports
func (h *HARMarketScraper) parseAffordabilityReports(html, reportType, sourceURL string) ([]*HARMarketData, error) {
	// Similar parsing logic for affordability reports
	return h.parseGenericReports(html, reportType, sourceURL)
}

// parseCommunitiesReports parses hottest communities reports
func (h *HARMarketScraper) parseCommunitiesReports(html, reportType, sourceURL string) ([]*HARMarketData, error) {
	// Similar parsing logic for community reports
	return h.parseGenericReports(html, reportType, sourceURL)
}

// parseGenericReports provides generic parsing for any HAR page
func (h *HARMarketScraper) parseGenericReports(html, reportType, sourceURL string) ([]*HARMarketData, error) {
	reports := []*HARMarketData{}
	
	// Generic pattern to find any report-like links
	genericPattern := regexp.MustCompile(`<a[^>]*href="([^"]+)"[^>]*>([^<]+(?:report|update|analysis|market|housing)[^<]*)</a>`)
	matches := genericPattern.FindAllStringSubmatch(html, -1)
	
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		
		relativeURL := strings.TrimSpace(match[1])
		title := strings.TrimSpace(match[2])
		
		if title == "" || relativeURL == "" {
			continue
		}
		
		// Skip navigation and non-content links
		if strings.Contains(strings.ToLower(title), "home") || 
		   strings.Contains(strings.ToLower(title), "contact") ||
		   len(title) < 10 {
			continue
		}
		
		fullURL := h.resolveURL(sourceURL, relativeURL)
		publishDate := h.extractDateFromText(title)
		if publishDate.IsZero() {
			publishDate = time.Now().AddDate(0, 0, -3) // Default to 3 days ago
		}
		
		report := &HARMarketData{
			ReportType:  reportType,
			Title:       title,
			URL:         fullURL,
			PublishDate: publishDate,
			Content:     "",
			Summary:     h.generateSummaryFromTitle(title),
			KeyMetrics:  "{}",
			Status:      "active",
		}
		
		reports = append(reports, report)
	}
	
	return reports, nil
}

// Helper functions

// resolveURL converts relative URLs to absolute URLs
func (h *HARMarketScraper) resolveURL(baseURL, relativeURL string) string {
	if strings.HasPrefix(relativeURL, "http") {
		return relativeURL
	}
	
	base, err := url.Parse(baseURL)
	if err != nil {
		return relativeURL
	}
	
	rel, err := url.Parse(relativeURL)
	if err != nil {
		return relativeURL
	}
	
	return base.ResolveReference(rel).String()
}

// extractDateFromText attempts to extract a date from text
func (h *HARMarketScraper) extractDateFromText(text string) time.Time {
	// Common date patterns in HAR reports
	datePatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(\w+)\s+(\d{4})`), // "January 2024"
		regexp.MustCompile(`(\d{1,2})/(\d{1,2})/(\d{4})`), // "1/15/2024"
		regexp.MustCompile(`(\d{4})-(\d{1,2})-(\d{1,2})`), // "2024-01-15"
		regexp.MustCompile(`(?i)(\d{4})\s+(\w+)`), // "2024 January"
	}
	
	for _, pattern := range datePatterns {
		match := pattern.FindStringSubmatch(text)
		if len(match) >= 3 {
			// Try to parse the date
			if date := h.parseMatchedDate(match); !date.IsZero() {
				return date
			}
		}
	}
	
	return time.Time{} // Return zero time if no date found
}

// parseMatchedDate parses a matched date pattern
func (h *HARMarketScraper) parseMatchedDate(match []string) time.Time {
	if len(match) < 3 {
		return time.Time{}
	}
	
	// Try different parsing strategies based on match length
	switch len(match) {
	case 3:
		// Two-part date
		if year, err := strconv.Atoi(match[2]); err == nil && year > 2020 && year <= time.Now().Year()+1 {
			// Assume it's "Month Year" format
			monthName := strings.ToLower(match[1])
			month := h.parseMonthName(monthName)
			if month > 0 {
				return time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
			}
		}
	case 4:
		// Three-part date (likely MM/DD/YYYY)
		if month, err := strconv.Atoi(match[1]); err == nil {
			if day, err := strconv.Atoi(match[2]); err == nil {
				if year, err := strconv.Atoi(match[3]); err == nil {
					if month >= 1 && month <= 12 && day >= 1 && day <= 31 && year > 2020 {
						return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
					}
				}
			}
		}
	}
	
	return time.Time{}
}

// parseMonthName converts month name to month number
func (h *HARMarketScraper) parseMonthName(monthName string) time.Month {
	months := map[string]time.Month{
		"january": time.January, "jan": time.January,
		"february": time.February, "feb": time.February,
		"march": time.March, "mar": time.March,
		"april": time.April, "apr": time.April,
		"may": time.May,
		"june": time.June, "jun": time.June,
		"july": time.July, "jul": time.July,
		"august": time.August, "aug": time.August,
		"september": time.September, "sep": time.September, "sept": time.September,
		"october": time.October, "oct": time.October,
		"november": time.November, "nov": time.November,
		"december": time.December, "dec": time.December,
	}
	
	return months[strings.ToLower(monthName)]
}

// generateSummaryFromTitle creates a basic summary from title
func (h *HARMarketScraper) generateSummaryFromTitle(title string) string {
	// Extract key information from title for summary
	title = strings.TrimSpace(title)
	if len(title) > 200 {
		return title[:197] + "..."
	}
	return title
}

// saveMarketData saves market data to database
func (h *HARMarketScraper) saveMarketData(data *HARMarketData) (bool, error) {
	// Check if report already exists
	var existing HARMarketData
	result := h.db.Where("url = ?", data.URL).First(&existing)
	
	if result.Error == gorm.ErrRecordNotFound {
		// New report - create it
		if err := h.db.Create(data).Error; err != nil {
			return false, err
		}
		log.Printf("📄 New report saved: %s", data.Title)
		return true, nil
	} else if result.Error != nil {
		return false, result.Error
	}
	
	// Report exists - update if needed
	existing.Title = data.Title
	existing.Summary = data.Summary
	existing.Status = data.Status
	existing.UpdatedAt = time.Now()
	
	if err := h.db.Save(&existing).Error; err != nil {
		return false, err
	}
	
	log.Printf("🔄 Report updated: %s", data.Title)
	return false, nil
}

// logScrapeResult logs the result of a scraping operation
func (h *HARMarketScraper) logScrapeResult(result ScrapeResult) {
	status := "success"
	errorMsg := ""
	
	if result.Error != nil {
		status = "failed"
		errorMsg = result.Error.Error()
	} else if result.ItemsFound == 0 {
		status = "partial"
	}
	
	scrapeLog := HARScrapeLog{
		ScrapeType:   result.ReportType,
		URL:          result.URL,
		Status:       status,
		ItemsFound:   result.ItemsFound,
		ItemsNew:     result.ItemsNew,
		ItemsUpdated: result.ItemsUpdated,
		ErrorMessage: errorMsg,
		Duration:     int(result.Duration.Milliseconds()),
	}
	
	if err := h.db.Create(&scrapeLog).Error; err != nil {
		log.Printf("⚠️  Failed to log scrape result: %v", err)
	}
}

// GetLatestReports retrieves the latest market reports
func (h *HARMarketScraper) GetLatestReports(reportType string, limit int) ([]*HARMarketData, error) {
	var reports []*HARMarketData
	query := h.db.Where("status = 'active'")
	
	if reportType != "" {
		query = query.Where("report_type = ?", reportType)
	}
	
	result := query.Order("publish_date DESC").
		Limit(limit).
		Find(&reports)
	
	// Handle empty results gracefully
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, result.Error
	}
	
	// Return empty array if no data
	if len(reports) == 0 {
		return []*HARMarketData{}, nil
	}
	
	return reports, nil
}

// GetScrapeStats returns scraping statistics
func (h *HARMarketScraper) GetScrapeStats() (map[string]interface{}, error) {
	var totalReports int64
	var totalLogs int64
	var recentSuccessful int64
	
	// Count total reports
	h.db.Model(&HARMarketData{}).Where("status = 'active'").Count(&totalReports)
	
	// Count total scrape logs
	h.db.Model(&HARScrapeLog{}).Count(&totalLogs)
	
	// Count recent successful scrapes (last 7 days)
	weekAgo := time.Now().AddDate(0, 0, -7)
	h.db.Model(&HARScrapeLog{}).
		Where("status = 'success' AND scraped_at > ?", weekAgo).
		Count(&recentSuccessful)
	
	return map[string]interface{}{
		"total_reports":        totalReports,
		"total_scrape_logs":    totalLogs,
		"recent_successful":    recentSuccessful,
		"scraping_targets":     len(h.targets),
		"last_scrape":          h.getLastScrapeTime(),
		"active_report_types":  h.getActiveReportTypes(),
	}, nil
}

// getLastScrapeTime gets the timestamp of the last scrape
func (h *HARMarketScraper) getLastScrapeTime() time.Time {
	var lastLog HARScrapeLog
	result := h.db.Order("scraped_at DESC").First(&lastLog)
	if result.Error != nil {
		return time.Time{}
	}
	return lastLog.ScrapedAt
}

// getActiveReportTypes gets list of active report types
func (h *HARMarketScraper) getActiveReportTypes() []string {
	var types []string
	h.db.Model(&HARMarketData{}).
		Where("status = 'active'").
		Distinct("report_type").
		Pluck("report_type", &types)
	return types
}
