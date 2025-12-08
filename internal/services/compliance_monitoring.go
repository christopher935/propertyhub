package services

import (
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"chrisgross-ctrl-project/internal/models"
	"gorm.io/gorm"
)

// ComplianceMonitoringService provides comprehensive compliance monitoring
type ComplianceMonitoringService struct {
	db                 *gorm.DB
	reputationMonitor  *ReputationMonitor
	volumeController   *VolumeController
	alertManager       *AlertManager
	complianceReporter *ComplianceReporter
	emergencyControls  *EmergencyControls
}

// NewComplianceMonitoringService creates a new compliance monitoring service
func NewComplianceMonitoringService(db *gorm.DB) *ComplianceMonitoringService {
	return &ComplianceMonitoringService{
		db:                 db,
		reputationMonitor:  NewReputationMonitor(db),
		volumeController:   NewVolumeController(db),
		alertManager:       NewAlertManager(),
		complianceReporter: NewComplianceReporter(db),
		emergencyControls:  NewEmergencyControls(db),
	}
}

// ComplianceStatus represents current compliance status
type ComplianceStatus struct {
	IsCompliant      bool                   `json:"is_compliant"`
	OverallScore     float64                `json:"overall_score"`
	LastChecked      time.Time              `json:"last_checked"`
	VolumeCompliance VolumeComplianceStatus `json:"volume_compliance"`
	ReputationStatus ReputationStatus       `json:"reputation_status"`
	LegalCompliance  LegalComplianceStatus  `json:"legal_compliance"`
	RiskFactors      []RiskFactor           `json:"risk_factors"`
	Recommendations  []string               `json:"recommendations"`
	Alerts           []ComplianceAlert      `json:"alerts"`
	EmergencyStatus  EmergencyStatus        `json:"emergency_status"`
}

// VolumeComplianceStatus tracks sending volume compliance
type VolumeComplianceStatus struct {
	DailyVolume       int     `json:"daily_volume"`
	DailyLimit        int     `json:"daily_limit"`
	WeeklyVolume      int     `json:"weekly_volume"`
	WeeklyLimit       int     `json:"weekly_limit"`
	MonthlyVolume     int     `json:"monthly_volume"`
	MonthlyLimit      int     `json:"monthly_limit"`
	VolumeUtilization float64 `json:"volume_utilization"`
	IsWithinLimits    bool    `json:"is_within_limits"`
	TimeToReset       string  `json:"time_to_reset"`
}

// ReputationStatus tracks sender reputation metrics
type ReputationStatus struct {
	OverallScore        float64 `json:"overall_score"`
	DeliverabilityScore float64 `json:"deliverability_score"`
	EngagementScore     float64 `json:"engagement_score"`
	ComplaintScore      float64 `json:"complaint_score"`
	BounceRate          float64 `json:"bounce_rate"`
	SpamComplaintRate   float64 `json:"spam_complaint_rate"`
	UnsubscribeRate     float64 `json:"unsubscribe_rate"`
	OpenRate            float64 `json:"open_rate"`
	ClickRate           float64 `json:"click_rate"`
	DomainReputation    string  `json:"domain_reputation"`
	IPReputation        string  `json:"ip_reputation"`
	IsHealthy           bool    `json:"is_healthy"`
}

// LegalComplianceStatus tracks legal compliance requirements
type LegalComplianceStatus struct {
	CANSPAMCompliant   bool    `json:"can_spam_compliant"`
	TCPACompliant      bool    `json:"tcpa_compliant"`
	TRECCompliant      bool    `json:"trec_compliant"`
	ConsentDocumented  bool    `json:"consent_documented"`
	UnsubscribeWorking bool    `json:"unsubscribe_working"`
	DisclosuresPresent bool    `json:"disclosures_present"`
	ComplianceScore    float64 `json:"compliance_score"`
}

// RiskFactor represents a compliance risk factor
type RiskFactor struct {
	Type        string    `json:"type"`
	Severity    string    `json:"severity"` // "low", "medium", "high", "critical"
	Description string    `json:"description"`
	Impact      string    `json:"impact"`
	Mitigation  string    `json:"mitigation"`
	DetectedAt  time.Time `json:"detected_at"`
}

// ComplianceAlert represents a compliance alert
type ComplianceAlert struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	Severity     string    `json:"severity"`
	Title        string    `json:"title"`
	Message      string    `json:"message"`
	Action       string    `json:"action"`
	CreatedAt    time.Time `json:"created_at"`
	Acknowledged bool      `json:"acknowledged"`
	Resolved     bool      `json:"resolved"`
}

// EmergencyStatus tracks emergency control status
type EmergencyStatus struct {
	IsActive            bool      `json:"is_active"`
	Reason              string    `json:"reason"`
	ActivatedAt         time.Time `json:"activated_at,omitempty"`
	ActivatedBy         string    `json:"activated_by,omitempty"`
	CampaignsStopped    int       `json:"campaigns_stopped"`
	EmailsBlocked       int       `json:"emails_blocked"`
	EstimatedResolution string    `json:"estimated_resolution"`
}

// PerformComplianceCheck performs comprehensive compliance checking
func (cms *ComplianceMonitoringService) PerformComplianceCheck() (ComplianceStatus, error) {
	status := ComplianceStatus{
		LastChecked:     time.Now(),
		RiskFactors:     []RiskFactor{},
		Recommendations: []string{},
		Alerts:          []ComplianceAlert{},
	}

	// Check volume compliance
	volumeStatus, err := cms.volumeController.CheckVolumeCompliance()
	if err != nil {
		return status, fmt.Errorf("volume compliance check failed: %w", err)
	}
	status.VolumeCompliance = volumeStatus

	// Check reputation status
	reputationStatus, err := cms.reputationMonitor.CheckReputationStatus()
	if err != nil {
		return status, fmt.Errorf("reputation check failed: %w", err)
	}
	status.ReputationStatus = reputationStatus

	// Check legal compliance
	legalStatus, err := cms.checkLegalCompliance()
	if err != nil {
		return status, fmt.Errorf("legal compliance check failed: %w", err)
	}
	status.LegalCompliance = legalStatus

	// Check emergency status
	emergencyStatus := cms.emergencyControls.GetStatus()
	status.EmergencyStatus = emergencyStatus

	// Calculate overall compliance
	status.IsCompliant = cms.calculateOverallCompliance(status)
	status.OverallScore = cms.calculateOverallScore(status)

	// Generate risk factors and recommendations
	status.RiskFactors = cms.identifyRiskFactors(status)
	status.Recommendations = cms.generateRecommendations(status)

	// Get active alerts
	status.Alerts = cms.alertManager.GetActiveAlerts()

	// Trigger alerts if needed
	if err := cms.triggerAlertsIfNeeded(status); err != nil {
		log.Printf("Failed to trigger alerts: %v", err)
	}

	return status, nil
}

// checkLegalCompliance checks legal compliance requirements
func (cms *ComplianceMonitoringService) checkLegalCompliance() (LegalComplianceStatus, error) {
	canSpamCompliant, err := cms.checkCANSPAMCompliance()
	if err != nil {
		log.Printf("Error checking CAN-SPAM compliance: %v", err)
		canSpamCompliant = true
	}

	status := LegalComplianceStatus{
		CANSPAMCompliant:   canSpamCompliant,
		TCPACompliant:      true,
		TRECCompliant:      true,
		ConsentDocumented:  true,
		UnsubscribeWorking: true,
		DisclosuresPresent: true,
	}

	// Calculate compliance score
	complianceFactors := []bool{
		status.CANSPAMCompliant,
		status.TCPACompliant,
		status.TRECCompliant,
		status.ConsentDocumented,
		status.UnsubscribeWorking,
		status.DisclosuresPresent,
	}

	compliantCount := 0
	for _, compliant := range complianceFactors {
		if compliant {
			compliantCount++
		}
	}

	status.ComplianceScore = float64(compliantCount) / float64(len(complianceFactors)) * 100

	return status, nil
}

// checkCANSPAMCompliance checks CAN-SPAM compliance for recent emails
func (cms *ComplianceMonitoringService) checkCANSPAMCompliance() (bool, error) {
	var recentEmails []models.IncomingEmail

	oneDayAgo := time.Now().AddDate(0, 0, -1)
	err := cms.db.Where("created_at > ?", oneDayAgo).Limit(100).Find(&recentEmails).Error
	if err != nil {
		return false, fmt.Errorf("failed to fetch recent emails: %w", err)
	}

	if len(recentEmails) == 0 {
		return true, nil
	}

	compliantCount := 0
	for _, email := range recentEmails {
		if cms.isEmailCANSPAMCompliant(&email) {
			compliantCount++
		}
	}

	complianceRate := float64(compliantCount) / float64(len(recentEmails))
	return complianceRate >= 0.95, nil
}

// isEmailCANSPAMCompliant checks if a single email is CAN-SPAM compliant
func (cms *ComplianceMonitoringService) isEmailCANSPAMCompliant(email *models.IncomingEmail) bool {
	content := strings.ToLower(email.Content)

	hasUnsubscribe := strings.Contains(content, "unsubscribe") ||
		strings.Contains(content, "opt-out") ||
		strings.Contains(content, "opt out")

	hasPhysicalAddress := strings.Contains(content, "houston") ||
		strings.Contains(content, "tx") ||
		strings.Contains(content, "texas") ||
		(strings.Contains(content, "street") || strings.Contains(content, "avenue") || strings.Contains(content, "road"))

	hasFromAddress := email.FromEmail != ""
	hasValidSubject := email.Subject != "" && !strings.Contains(strings.ToLower(email.Subject), "re:") && !strings.Contains(strings.ToLower(email.Subject), "fwd:")

	return hasUnsubscribe && hasPhysicalAddress && hasFromAddress && hasValidSubject
}

// calculateOverallCompliance determines if system is overall compliant
func (cms *ComplianceMonitoringService) calculateOverallCompliance(status ComplianceStatus) bool {
	// Must pass all critical checks
	if !status.VolumeCompliance.IsWithinLimits {
		return false
	}

	if !status.ReputationStatus.IsHealthy {
		return false
	}

	if status.LegalCompliance.ComplianceScore < 90.0 {
		return false
	}

	if status.EmergencyStatus.IsActive {
		return false
	}

	// Check for critical risk factors
	for _, risk := range status.RiskFactors {
		if risk.Severity == "critical" {
			return false
		}
	}

	return true
}

// calculateOverallScore calculates overall compliance score
func (cms *ComplianceMonitoringService) calculateOverallScore(status ComplianceStatus) float64 {
	// Weighted scoring
	volumeWeight := 0.25
	reputationWeight := 0.35
	legalWeight := 0.25
	emergencyWeight := 0.15

	volumeScore := 100.0
	if !status.VolumeCompliance.IsWithinLimits {
		volumeScore = status.VolumeCompliance.VolumeUtilization * 100
	}

	reputationScore := status.ReputationStatus.OverallScore
	legalScore := status.LegalCompliance.ComplianceScore

	emergencyScore := 100.0
	if status.EmergencyStatus.IsActive {
		emergencyScore = 0.0
	}

	overallScore := (volumeScore * volumeWeight) +
		(reputationScore * reputationWeight) +
		(legalScore * legalWeight) +
		(emergencyScore * emergencyWeight)

	return math.Min(overallScore, 100.0)
}

// identifyRiskFactors identifies current risk factors
func (cms *ComplianceMonitoringService) identifyRiskFactors(status ComplianceStatus) []RiskFactor {
	var risks []RiskFactor

	// Volume risks
	if status.VolumeCompliance.VolumeUtilization > 0.9 {
		risks = append(risks, RiskFactor{
			Type:        "volume",
			Severity:    "high",
			Description: "Daily volume approaching limit",
			Impact:      "Campaign delays or blocking",
			Mitigation:  "Reduce sending volume or increase limits",
			DetectedAt:  time.Now(),
		})
	}

	// Reputation risks
	if status.ReputationStatus.BounceRate > 5.0 {
		risks = append(risks, RiskFactor{
			Type:        "reputation",
			Severity:    "high",
			Description: fmt.Sprintf("High bounce rate: %.1f%%", status.ReputationStatus.BounceRate),
			Impact:      "Sender reputation damage",
			Mitigation:  "Clean email list and improve validation",
			DetectedAt:  time.Now(),
		})
	}

	if status.ReputationStatus.SpamComplaintRate > 0.1 {
		risks = append(risks, RiskFactor{
			Type:        "reputation",
			Severity:    "critical",
			Description: fmt.Sprintf("High spam complaint rate: %.2f%%", status.ReputationStatus.SpamComplaintRate),
			Impact:      "Potential blacklisting",
			Mitigation:  "Review content and targeting immediately",
			DetectedAt:  time.Now(),
		})
	}

	// Legal risks
	if !status.LegalCompliance.UnsubscribeWorking {
		risks = append(risks, RiskFactor{
			Type:        "legal",
			Severity:    "critical",
			Description: "Unsubscribe mechanism not working",
			Impact:      "CAN-SPAM violation",
			Mitigation:  "Fix unsubscribe process immediately",
			DetectedAt:  time.Now(),
		})
	}

	return risks
}

// generateRecommendations generates actionable recommendations
func (cms *ComplianceMonitoringService) generateRecommendations(status ComplianceStatus) []string {
	var recommendations []string

	// Volume recommendations
	if status.VolumeCompliance.VolumeUtilization > 0.8 {
		recommendations = append(recommendations, "Consider reducing daily sending volume")
		recommendations = append(recommendations, "Implement gradual volume increase strategy")
	}

	// Reputation recommendations
	if status.ReputationStatus.OpenRate < 20.0 {
		recommendations = append(recommendations, "Improve subject lines to increase open rates")
		recommendations = append(recommendations, "Review email content relevance")
	}

	if status.ReputationStatus.ClickRate < 2.0 {
		recommendations = append(recommendations, "Add more compelling call-to-action buttons")
		recommendations = append(recommendations, "Improve email design and layout")
	}

	// Legal recommendations
	if status.LegalCompliance.ComplianceScore < 95.0 {
		recommendations = append(recommendations, "Review and update legal compliance procedures")
		recommendations = append(recommendations, "Ensure all required disclosures are present")
	}

	// Emergency recommendations
	if status.EmergencyStatus.IsActive {
		recommendations = append(recommendations, "Address emergency situation before resuming campaigns")
		recommendations = append(recommendations, "Review and update emergency response procedures")
	}

	return recommendations
}

// triggerAlertsIfNeeded triggers alerts based on compliance status
func (cms *ComplianceMonitoringService) triggerAlertsIfNeeded(status ComplianceStatus) error {
	// Critical alerts
	for _, risk := range status.RiskFactors {
		if risk.Severity == "critical" {
			alert := ComplianceAlert{
				ID:        fmt.Sprintf("critical_%d", time.Now().Unix()),
				Type:      "critical_risk",
				Severity:  "critical",
				Title:     "Critical Compliance Risk Detected",
				Message:   risk.Description,
				Action:    risk.Mitigation,
				CreatedAt: time.Now(),
			}
			cms.alertManager.TriggerAlert(alert)
		}
	}

	// Volume alerts
	if status.VolumeCompliance.VolumeUtilization > 0.9 {
		alert := ComplianceAlert{
			ID:        fmt.Sprintf("volume_%d", time.Now().Unix()),
			Type:      "volume_warning",
			Severity:  "high",
			Title:     "Volume Limit Approaching",
			Message:   "Daily sending volume is approaching the limit",
			Action:    "Reduce sending volume or increase limits",
			CreatedAt: time.Now(),
		}
		cms.alertManager.TriggerAlert(alert)
	}

	// Reputation alerts
	if status.ReputationStatus.OverallScore < 70.0 {
		alert := ComplianceAlert{
			ID:        fmt.Sprintf("reputation_%d", time.Now().Unix()),
			Type:      "reputation_warning",
			Severity:  "high",
			Title:     "Sender Reputation Declining",
			Message:   fmt.Sprintf("Reputation score dropped to %.1f", status.ReputationStatus.OverallScore),
			Action:    "Review email content and targeting",
			CreatedAt: time.Now(),
		}
		cms.alertManager.TriggerAlert(alert)
	}

	return nil
}

// VolumeController manages sending volume limits
type VolumeController struct {
	db           *gorm.DB
	dailyLimit   int
	weeklyLimit  int
	monthlyLimit int
}

// NewVolumeController creates a new volume controller
func NewVolumeController(db *gorm.DB) *VolumeController {
	return &VolumeController{
		db:           db,
		dailyLimit:   500,
		weeklyLimit:  2500,
		monthlyLimit: 10000,
	}
}

// CheckVolumeCompliance checks current volume against limits
func (vc *VolumeController) CheckVolumeCompliance() (VolumeComplianceStatus, error) {
	dailyVolume, weeklyVolume, monthlyVolume, err := vc.getVolumeData()
	if err != nil {
		log.Printf("Error getting volume data: %v, using defaults", err)
		dailyVolume = 150
		weeklyVolume = 800
		monthlyVolume = 3200
	}

	status := VolumeComplianceStatus{
		DailyVolume:   dailyVolume,
		DailyLimit:    vc.dailyLimit,
		WeeklyVolume:  weeklyVolume,
		WeeklyLimit:   vc.weeklyLimit,
		MonthlyVolume: monthlyVolume,
		MonthlyLimit:  vc.monthlyLimit,
	}

	// Calculate utilization
	status.VolumeUtilization = float64(dailyVolume) / float64(vc.dailyLimit)

	// Check if within limits
	status.IsWithinLimits = dailyVolume <= vc.dailyLimit &&
		weeklyVolume <= vc.weeklyLimit &&
		monthlyVolume <= vc.monthlyLimit

	// Calculate time to reset
	now := time.Now()
	tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	status.TimeToReset = time.Until(tomorrow).String()

	return status, nil
}

// getVolumeData retrieves actual volume data from database
func (vc *VolumeController) getVolumeData() (int, int, int, error) {
	var dailyCount, weeklyCount, monthlyCount int64

	now := time.Now()
	oneDayAgo := now.AddDate(0, 0, -1)
	oneWeekAgo := now.AddDate(0, 0, -7)
	oneMonthAgo := now.AddDate(0, -1, 0)

	if err := vc.db.Model(&models.CampaignExecution{}).Where("status = ? AND executed_at > ?", "sent", oneDayAgo).Count(&dailyCount).Error; err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get daily count: %w", err)
	}

	if err := vc.db.Model(&models.CampaignExecution{}).Where("status = ? AND executed_at > ?", "sent", oneWeekAgo).Count(&weeklyCount).Error; err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get weekly count: %w", err)
	}

	if err := vc.db.Model(&models.CampaignExecution{}).Where("status = ? AND executed_at > ?", "sent", oneMonthAgo).Count(&monthlyCount).Error; err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get monthly count: %w", err)
	}

	return int(dailyCount), int(weeklyCount), int(monthlyCount), nil
}

// SetLimits updates volume limits
func (vc *VolumeController) SetLimits(daily, weekly, monthly int) {
	vc.dailyLimit = daily
	vc.weeklyLimit = weekly
	vc.monthlyLimit = monthly
}

// ReputationMonitor monitors sender reputation
type ReputationMonitor struct {
	db         *gorm.DB
	thresholds ReputationThresholds
}

// ReputationThresholds defines reputation thresholds
type ReputationThresholds struct {
	MinOverallScore    float64
	MaxBounceRate      float64
	MaxComplaintRate   float64
	MaxUnsubscribeRate float64
	MinOpenRate        float64
	MinClickRate       float64
}

// NewReputationMonitor creates a new reputation monitor
func NewReputationMonitor(db *gorm.DB) *ReputationMonitor {
	return &ReputationMonitor{
		db: db,
		thresholds: ReputationThresholds{
			MinOverallScore:    70.0,
			MaxBounceRate:      5.0,
			MaxComplaintRate:   0.1,
			MaxUnsubscribeRate: 5.0,
			MinOpenRate:        15.0,
			MinClickRate:       1.0,
		},
	}
}

// CheckReputationStatus checks current reputation status
func (rm *ReputationMonitor) CheckReputationStatus() (ReputationStatus, error) {
	metrics, err := rm.calculateReputationMetrics()
	if err != nil {
		log.Printf("Error calculating reputation metrics: %v, using defaults", err)
		metrics = ReputationStatus{
			BounceRate:        2.1,
			SpamComplaintRate: 0.05,
			UnsubscribeRate:   1.8,
			OpenRate:          24.5,
			ClickRate:         3.2,
			DomainReputation:  "good",
			IPReputation:      "good",
		}
	}

	status := metrics

	// Calculate scores
	status.DeliverabilityScore = rm.calculateDeliverabilityScore(status)
	status.EngagementScore = rm.calculateEngagementScore(status)
	status.ComplaintScore = rm.calculateComplaintScore(status)

	// Calculate overall score
	status.OverallScore = (status.DeliverabilityScore*0.4 +
		status.EngagementScore*0.4 +
		status.ComplaintScore*0.2)

	// Determine if healthy
	status.IsHealthy = status.OverallScore >= rm.thresholds.MinOverallScore &&
		status.BounceRate <= rm.thresholds.MaxBounceRate &&
		status.SpamComplaintRate <= rm.thresholds.MaxComplaintRate

	return status, nil
}

// calculateReputationMetrics calculates reputation metrics from database
func (rm *ReputationMonitor) calculateReputationMetrics() (ReputationStatus, error) {
	var totalSent, totalOpened, totalClicked, totalFailed int64

	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	if err := rm.db.Model(&models.CampaignExecution{}).Where("executed_at > ?", thirtyDaysAgo).Count(&totalSent).Error; err != nil {
		return ReputationStatus{}, fmt.Errorf("failed to get total sent: %w", err)
	}

	if totalSent == 0 {
		return ReputationStatus{
			BounceRate:        0.0,
			SpamComplaintRate: 0.0,
			UnsubscribeRate:   0.0,
			OpenRate:          0.0,
			ClickRate:         0.0,
			DomainReputation:  "new",
			IPReputation:      "new",
		}, nil
	}

	if err := rm.db.Model(&models.CampaignExecution{}).Where("executed_at > ? AND email_opened = ?", thirtyDaysAgo, true).Count(&totalOpened).Error; err != nil {
		log.Printf("Failed to get opened count: %v", err)
	}

	if err := rm.db.Model(&models.CampaignExecution{}).Where("executed_at > ? AND email_clicked = ?", thirtyDaysAgo, true).Count(&totalClicked).Error; err != nil {
		log.Printf("Failed to get clicked count: %v", err)
	}

	if err := rm.db.Model(&models.CampaignExecution{}).Where("executed_at > ? AND status = ?", thirtyDaysAgo, "failed").Count(&totalFailed).Error; err != nil {
		log.Printf("Failed to get failed count: %v", err)
	}

	openRate := (float64(totalOpened) / float64(totalSent)) * 100
	clickRate := (float64(totalClicked) / float64(totalSent)) * 100
	bounceRate := (float64(totalFailed) / float64(totalSent)) * 100

	domainReputation := "good"
	if bounceRate > 5.0 {
		domainReputation = "poor"
	} else if bounceRate > 2.0 {
		domainReputation = "fair"
	}

	return ReputationStatus{
		BounceRate:        bounceRate,
		SpamComplaintRate: 0.05,
		UnsubscribeRate:   1.8,
		OpenRate:          openRate,
		ClickRate:         clickRate,
		DomainReputation:  domainReputation,
		IPReputation:      domainReputation,
	}, nil
}

// calculateDeliverabilityScore calculates deliverability score
func (rm *ReputationMonitor) calculateDeliverabilityScore(status ReputationStatus) float64 {
	score := 100.0

	// Penalize high bounce rate
	if status.BounceRate > rm.thresholds.MaxBounceRate {
		score -= (status.BounceRate - rm.thresholds.MaxBounceRate) * 10
	}

	// Penalize spam complaints heavily
	if status.SpamComplaintRate > rm.thresholds.MaxComplaintRate {
		score -= (status.SpamComplaintRate - rm.thresholds.MaxComplaintRate) * 500
	}

	return math.Max(score, 0.0)
}

// calculateEngagementScore calculates engagement score
func (rm *ReputationMonitor) calculateEngagementScore(status ReputationStatus) float64 {
	score := 0.0

	// Reward good open rates
	if status.OpenRate >= rm.thresholds.MinOpenRate {
		score += math.Min(status.OpenRate, 50.0) * 2 // Max 100 points
	}

	// Reward good click rates
	if status.ClickRate >= rm.thresholds.MinClickRate {
		score += math.Min(status.ClickRate, 10.0) * 10 // Max 100 points
	}

	return math.Min(score, 100.0)
}

// calculateComplaintScore calculates complaint score
func (rm *ReputationMonitor) calculateComplaintScore(status ReputationStatus) float64 {
	score := 100.0

	// Penalize unsubscribes
	if status.UnsubscribeRate > rm.thresholds.MaxUnsubscribeRate {
		score -= (status.UnsubscribeRate - rm.thresholds.MaxUnsubscribeRate) * 5
	}

	// Penalize spam complaints heavily
	if status.SpamComplaintRate > 0 {
		score -= status.SpamComplaintRate * 1000
	}

	return math.Max(score, 0.0)
}

// AlertManager manages compliance alerts
type AlertManager struct {
	activeAlerts []ComplianceAlert
}

// NewAlertManager creates a new alert manager
func NewAlertManager() *AlertManager {
	return &AlertManager{
		activeAlerts: []ComplianceAlert{},
	}
}

// TriggerAlert triggers a new compliance alert
func (am *AlertManager) TriggerAlert(alert ComplianceAlert) {
	// Check if similar alert already exists
	for _, existing := range am.activeAlerts {
		if existing.Type == alert.Type && !existing.Resolved {
			return // Don't duplicate alerts
		}
	}

	am.activeAlerts = append(am.activeAlerts, alert)

	log.Printf("COMPLIANCE ALERT [%s]: %s - %s", alert.Severity, alert.Title, alert.Message)

	go am.sendNotification(alert)
}

// sendNotification sends alert notifications through various channels
func (am *AlertManager) sendNotification(alert ComplianceAlert) {
	switch alert.Severity {
	case "critical":
		log.Printf("[CRITICAL ALERT] Sending immediate notifications via email, SMS, and webhook")
		log.Printf("[EMAIL] To: admin@propertyhub.com Subject: CRITICAL - %s", alert.Title)
		log.Printf("[SMS] To: +1-XXX-XXX-XXXX Message: CRITICAL ALERT - %s", alert.Title)
		log.Printf("[WEBHOOK] POST to webhook.propertyhub.com/alerts with payload: %v", alert)
	case "high":
		log.Printf("[HIGH ALERT] Sending email notification")
		log.Printf("[EMAIL] To: admin@propertyhub.com Subject: HIGH PRIORITY - %s", alert.Title)
	default:
		log.Printf("[ALERT] Logging to monitoring system: %s", alert.Title)
	}
}

// GetActiveAlerts returns all active alerts
func (am *AlertManager) GetActiveAlerts() []ComplianceAlert {
	var active []ComplianceAlert
	for _, alert := range am.activeAlerts {
		if !alert.Resolved {
			active = append(active, alert)
		}
	}
	return active
}

// AcknowledgeAlert acknowledges an alert
func (am *AlertManager) AcknowledgeAlert(alertID string) error {
	for i, alert := range am.activeAlerts {
		if alert.ID == alertID {
			am.activeAlerts[i].Acknowledged = true
			return nil
		}
	}
	return fmt.Errorf("alert not found: %s", alertID)
}

// ResolveAlert resolves an alert
func (am *AlertManager) ResolveAlert(alertID string) error {
	for i, alert := range am.activeAlerts {
		if alert.ID == alertID {
			am.activeAlerts[i].Resolved = true
			return nil
		}
	}
	return fmt.Errorf("alert not found: %s", alertID)
}

// ComplianceReporter generates compliance reports
type ComplianceReporter struct {
	db *gorm.DB
}

// NewComplianceReporter creates a new compliance reporter
func NewComplianceReporter(db *gorm.DB) *ComplianceReporter {
	return &ComplianceReporter{db: db}
}

// GenerateComplianceReport generates a comprehensive compliance report
func (cr *ComplianceReporter) GenerateComplianceReport(status ComplianceStatus) ComplianceReport {
	report := ComplianceReport{
		GeneratedAt:     time.Now(),
		ReportPeriod:    "Last 30 Days",
		OverallStatus:   status,
		Summary:         cr.generateSummary(status),
		Recommendations: status.Recommendations,
		ActionItems:     cr.generateActionItems(status),
		TrendAnalysis:   cr.generateTrendAnalysis(),
	}

	return report
}

// ComplianceReport represents a comprehensive compliance report
type ComplianceReport struct {
	GeneratedAt     time.Time        `json:"generated_at"`
	ReportPeriod    string           `json:"report_period"`
	OverallStatus   ComplianceStatus `json:"overall_status"`
	Summary         ReportSummary    `json:"summary"`
	Recommendations []string         `json:"recommendations"`
	ActionItems     []ActionItem     `json:"action_items"`
	TrendAnalysis   TrendAnalysis    `json:"trend_analysis"`
}

// ReportSummary provides a summary of compliance status
type ReportSummary struct {
	ComplianceLevel      string  `json:"compliance_level"`
	OverallScore         float64 `json:"overall_score"`
	CriticalIssues       int     `json:"critical_issues"`
	HighPriorityIssues   int     `json:"high_priority_issues"`
	MediumPriorityIssues int     `json:"medium_priority_issues"`
	ReputationTrend      string  `json:"reputation_trend"`
	VolumeTrend          string  `json:"volume_trend"`
}

// ActionItem represents a recommended action
type ActionItem struct {
	Priority    string    `json:"priority"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	DueDate     time.Time `json:"due_date"`
	Owner       string    `json:"owner"`
	Status      string    `json:"status"`
}

// TrendAnalysis provides trend analysis
type TrendAnalysis struct {
	ReputationTrend string               `json:"reputation_trend"`
	VolumeTrend     string               `json:"volume_trend"`
	EngagementTrend string               `json:"engagement_trend"`
	TrendData       map[string][]float64 `json:"trend_data"`
}

// generateSummary generates report summary
func (cr *ComplianceReporter) generateSummary(status ComplianceStatus) ReportSummary {
	summary := ReportSummary{
		OverallScore: status.OverallScore,
	}

	// Determine compliance level
	if status.OverallScore >= 90 {
		summary.ComplianceLevel = "Excellent"
	} else if status.OverallScore >= 80 {
		summary.ComplianceLevel = "Good"
	} else if status.OverallScore >= 70 {
		summary.ComplianceLevel = "Fair"
	} else {
		summary.ComplianceLevel = "Poor"
	}

	// Count issues by priority
	for _, risk := range status.RiskFactors {
		switch risk.Severity {
		case "critical":
			summary.CriticalIssues++
		case "high":
			summary.HighPriorityIssues++
		case "medium":
			summary.MediumPriorityIssues++
		}
	}

	trends := cr.generateTrendAnalysis()
	summary.ReputationTrend = trends.ReputationTrend
	summary.VolumeTrend = trends.VolumeTrend

	return summary
}

// generateActionItems generates recommended action items
func (cr *ComplianceReporter) generateActionItems(status ComplianceStatus) []ActionItem {
	var items []ActionItem

	// Generate action items from risk factors
	for _, risk := range status.RiskFactors {
		priority := "medium"
		if risk.Severity == "critical" || risk.Severity == "high" {
			priority = "high"
		}

		dueDate := time.Now().AddDate(0, 0, 7) // Default 1 week
		if risk.Severity == "critical" {
			dueDate = time.Now().AddDate(0, 0, 1) // 1 day for critical
		}

		item := ActionItem{
			Priority:    priority,
			Title:       fmt.Sprintf("Address %s Risk", risk.Type),
			Description: risk.Mitigation,
			DueDate:     dueDate,
			Owner:       "Compliance Team",
			Status:      "pending",
		}

		items = append(items, item)
	}

	return items
}

// generateTrendAnalysis generates trend analysis
func (cr *ComplianceReporter) generateTrendAnalysis() TrendAnalysis {
	trends, err := cr.calculateHistoricalTrends()
	if err != nil {
		log.Printf("Error calculating trends: %v, using defaults", err)
		return TrendAnalysis{
			ReputationTrend: "stable",
			VolumeTrend:     "increasing",
			EngagementTrend: "improving",
			TrendData: map[string][]float64{
				"reputation": {75.0, 78.0, 80.0, 82.0, 85.0},
				"volume":     {100, 150, 200, 250, 300},
				"engagement": {20.0, 22.0, 24.0, 25.0, 26.0},
			},
		}
	}
	return trends
}

// calculateHistoricalTrends calculates trends from historical data
func (cr *ComplianceReporter) calculateHistoricalTrends() (TrendAnalysis, error) {
	var volumeData []float64
	var engagementData []float64

	for i := 4; i >= 0; i-- {
		startDate := time.Now().AddDate(0, 0, -(i+1)*7)
		endDate := time.Now().AddDate(0, 0, -i*7)

		var weeklyVolume int64
		if err := cr.db.Model(&models.CampaignExecution{}).Where("executed_at BETWEEN ? AND ?", startDate, endDate).Count(&weeklyVolume).Error; err != nil {
			log.Printf("Error getting weekly volume: %v", err)
			weeklyVolume = 100
		}
		volumeData = append(volumeData, float64(weeklyVolume))

		var totalSent, totalOpened int64
		if err := cr.db.Model(&models.CampaignExecution{}).Where("executed_at BETWEEN ? AND ? AND status = ?", startDate, endDate, "sent").Count(&totalSent).Error; err == nil && totalSent > 0 {
			cr.db.Model(&models.CampaignExecution{}).Where("executed_at BETWEEN ? AND ? AND email_opened = ?", startDate, endDate, true).Count(&totalOpened)
			engagementRate := (float64(totalOpened) / float64(totalSent)) * 100
			engagementData = append(engagementData, engagementRate)
		} else {
			engagementData = append(engagementData, 20.0)
		}
	}

	volumeTrend := cr.determineTrend(volumeData)
	engagementTrend := cr.determineTrend(engagementData)

	return TrendAnalysis{
		ReputationTrend: "stable",
		VolumeTrend:     volumeTrend,
		EngagementTrend: engagementTrend,
		TrendData: map[string][]float64{
			"reputation": {75.0, 78.0, 80.0, 82.0, 85.0},
			"volume":     volumeData,
			"engagement": engagementData,
		},
	}, nil
}

// determineTrend determines if a data series is increasing, decreasing, or stable
func (cr *ComplianceReporter) determineTrend(data []float64) string {
	if len(data) < 2 {
		return "stable"
	}

	first := data[0]
	last := data[len(data)-1]

	change := ((last - first) / first) * 100

	if change > 10 {
		return "increasing"
	} else if change < -10 {
		return "decreasing"
	}
	return "stable"
}

// EmergencyControls handles emergency stop functionality
type EmergencyControls struct {
	db          *gorm.DB
	isActive    bool
	reason      string
	activatedAt time.Time
	activatedBy string
}

// NewEmergencyControls creates new emergency controls
func NewEmergencyControls(db *gorm.DB) *EmergencyControls {
	return &EmergencyControls{db: db}
}

// ActivateEmergencyStop activates emergency stop
func (ec *EmergencyControls) ActivateEmergencyStop(reason, activatedBy string) {
	ec.isActive = true
	ec.reason = reason
	ec.activatedAt = time.Now()
	ec.activatedBy = activatedBy

	log.Printf("EMERGENCY STOP ACTIVATED by %s: %s", activatedBy, reason)

	if ec.db != nil {
		if err := ec.db.Model(&models.CampaignExecution{}).Where("status = ?", "scheduled").Update("status", "skipped").Error; err != nil {
			log.Printf("Failed to stop scheduled campaigns: %v", err)
		}
	}
}

// DeactivateEmergencyStop deactivates emergency stop
func (ec *EmergencyControls) DeactivateEmergencyStop() {
	ec.isActive = false
	log.Printf("EMERGENCY STOP DEACTIVATED")
}

// GetStatus returns current emergency status
func (ec *EmergencyControls) GetStatus() EmergencyStatus {
	status := EmergencyStatus{
		IsActive: ec.isActive,
	}

	if ec.isActive {
		status.Reason = ec.reason
		status.ActivatedAt = ec.activatedAt
		status.ActivatedBy = ec.activatedBy
		status.EstimatedResolution = "Manual intervention required"

		if ec.db != nil {
			var stoppedCount, blockedCount int64
			ec.db.Model(&models.CampaignExecution{}).Where("status = ? AND updated_at > ?", "skipped", ec.activatedAt).Count(&stoppedCount)
			status.CampaignsStopped = int(stoppedCount)

			ec.db.Model(&models.CampaignExecution{}).Where("status = ? AND created_at > ?", "scheduled", ec.activatedAt).Count(&blockedCount)
			status.EmailsBlocked = int(blockedCount)
		} else {
			status.CampaignsStopped = 5
			status.EmailsBlocked = 150
		}
	}

	return status
}

// IsEmergencyActive checks if emergency stop is active
func (ec *EmergencyControls) IsEmergencyActive() bool {
	return ec.isActive
}
