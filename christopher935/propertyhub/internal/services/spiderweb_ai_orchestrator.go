package services

import (
	"log"
	"time"

	"gorm.io/gorm"
)

// SpiderwebAIOrchestrator coordinates all AI intelligence modules
type SpiderwebAIOrchestrator struct {
	db                     *gorm.DB
	relationshipEngine     *RelationshipIntelligenceEngine
	funnelAnalytics        *FunnelAnalyticsService
	propertyMatcher        *PropertyMatchingService
	campaignTriggers       *CampaignTriggerAutomation
	scoringEngine          *BehavioralScoringEngine
	insightGenerator       *InsightGeneratorService
	emailBatch             *EmailBatchService
	abandonmentRecovery    *AbandonmentRecoveryService
	cache                  *IntelligenceCacheService
}

// NewSpiderwebAIOrchestrator creates and initializes the complete AI system
func NewSpiderwebAIOrchestrator(
	db *gorm.DB,
	scoringEngine *BehavioralScoringEngine,
	insightGenerator *InsightGeneratorService,
	emailBatch *EmailBatchService,
	abandonmentRecovery *AbandonmentRecoveryService,
	cache *IntelligenceCacheService,
) *SpiderwebAIOrchestrator {
	
	log.Println("🕸️ Initializing Spiderweb AI System...")
	
	// Initialize all modules
	funnelAnalytics := NewFunnelAnalyticsService(db)
	propertyMatcher := NewPropertyMatchingService(db)
	relationshipEngine := NewRelationshipIntelligenceEngine(db, scoringEngine)
	
	// Set cross-dependencies
	relationshipEngine.SetFunnelAnalytics(funnelAnalytics)
	relationshipEngine.SetPropertyMatcher(propertyMatcher)
	relationshipEngine.SetInsightGenerator(insightGenerator)
	
	// Initialize campaign triggers
	campaignTriggers := NewCampaignTriggerAutomation(
		db,
		emailBatch,
		relationshipEngine,
		propertyMatcher,
		abandonmentRecovery,
	)
	
	orchestrator := &SpiderwebAIOrchestrator{
		db:                  db,
		relationshipEngine:  relationshipEngine,
		funnelAnalytics:     funnelAnalytics,
		propertyMatcher:     propertyMatcher,
		campaignTriggers:    campaignTriggers,
		scoringEngine:       scoringEngine,
		insightGenerator:    insightGenerator,
		emailBatch:          emailBatch,
		abandonmentRecovery: abandonmentRecovery,
		cache:               cache,
	}
	
	log.Println("✅ Spiderweb AI System initialized successfully")
	if cache != nil && cache.IsAvailable() {
		log.Println("✅ Redis intelligence cache enabled")
	} else {
		log.Println("⚠️ Redis cache not available - running without cache")
	}
	
	return orchestrator
}

// RunIntelligenceCycle runs a complete intelligence analysis and action cycle
func (sao *SpiderwebAIOrchestrator) RunIntelligenceCycle() error {
	log.Println("🕸️ ═══════════════════════════════════════════════════════")
	log.Println("🕸️ Starting Spiderweb AI Intelligence Cycle")
	log.Println("🕸️ ═══════════════════════════════════════════════════════")
	
	startTime := time.Now()
	
	// Step 1: Analyze opportunities (relationship intelligence)
	log.Println("\n📊 Step 1: Analyzing cross-entity relationships...")
	opportunities, err := sao.relationshipEngine.AnalyzeOpportunities()
	if err != nil {
		log.Printf("❌ Error analyzing opportunities: %v", err)
	} else {
		log.Printf("✅ Found %d opportunities", len(opportunities))
		
		// Log top 5 opportunities
		for i, opp := range opportunities {
			if i >= 5 {
				break
			}
			log.Printf("   %d. [Priority %d] %s - %s", i+1, opp.Priority, opp.Type, opp.LeadName)
		}
	}
	
	// Step 2: Analyze funnel performance
	log.Println("\n📈 Step 2: Analyzing conversion funnel...")
	funnelAnalysis, err := sao.funnelAnalytics.AnalyzeFunnel(30)
	if err != nil {
		log.Printf("❌ Error analyzing funnel: %v", err)
	} else {
		log.Printf("✅ Funnel analysis complete:")
		log.Printf("   Overall conversion: %.1f%%", funnelAnalysis.OverallConversion)
		log.Printf("   Total leads: %d", funnelAnalysis.TotalLeads)
		log.Printf("   Converted: %d", funnelAnalysis.Converted)
		log.Printf("   Bottlenecks: %v", funnelAnalysis.Bottlenecks)
	}
	
	// Step 3: Find new property matches
	log.Println("\n🏠 Step 3: Finding new property matches...")
	matches, err := sao.propertyMatcher.FindNewMatchesSince(time.Now().AddDate(0, 0, -1))
	if err != nil {
		log.Printf("❌ Error finding matches: %v", err)
	} else {
		log.Printf("✅ Found %d new property matches", len(matches))
	}
	
	// Step 4: Execute automated campaigns
	log.Println("\n📧 Step 4: Executing automated campaigns...")
	err = sao.campaignTriggers.RunAllTriggers()
	if err != nil {
		log.Printf("❌ Error running campaign triggers: %v", err)
	} else {
		log.Println("✅ Campaign triggers processed")
	}
	
	duration := time.Since(startTime)
	
	log.Println("\n🕸️ ═══════════════════════════════════════════════════════")
	log.Printf("🕸️ Intelligence Cycle Complete (%.2f seconds)", duration.Seconds())
	log.Println("🕸️ ═══════════════════════════════════════════════════════\n")
	
	return nil
}

// GetDashboardIntelligence returns intelligence data for the admin dashboard
func (sao *SpiderwebAIOrchestrator) GetDashboardIntelligence() (map[string]interface{}, error) {
	log.Println("📊 Gathering dashboard intelligence...")
	
	// Get top opportunities
	opportunities, _ := sao.relationshipEngine.AnalyzeOpportunities()
	
	// Get funnel analysis
	funnelAnalysis, _ := sao.funnelAnalytics.AnalyzeFunnel(30)
	
	// Get recent matches
	recentMatches, _ := sao.propertyMatcher.FindNewMatchesSince(time.Now().AddDate(0, 0, -7))
	
	// Compile intelligence
	intelligence := map[string]interface{}{
		"top_opportunities": opportunities[:min(10, len(opportunities))],
		"funnel_analysis":   funnelAnalysis,
		"recent_matches":    recentMatches[:min(10, len(recentMatches))],
		"generated_at":      time.Now(),
	}
	
	return intelligence, nil
}

// GetOpportunityInsights returns AI-generated insights for opportunities
func (sao *SpiderwebAIOrchestrator) GetOpportunityInsights() ([]Opportunity, error) {
	return sao.relationshipEngine.AnalyzeOpportunities()
}

// GetFunnelInsights returns funnel performance insights
func (sao *SpiderwebAIOrchestrator) GetFunnelInsights(days int) (*FunnelAnalysis, error) {
	return sao.funnelAnalytics.AnalyzeFunnel(days)
}

// GetPropertyMatches finds property matches for a specific lead
func (sao *SpiderwebAIOrchestrator) GetPropertyMatches(leadID int64) ([]PropertyMatch, error) {
	return sao.propertyMatcher.FindMatchesForLead(leadID)
}

// AnalyzeLeadOpportunity analyzes a specific lead for opportunities
func (sao *SpiderwebAIOrchestrator) AnalyzeLeadOpportunity(leadID int64) (map[string]interface{}, error) {
	// Get property matches
	matches, _ := sao.propertyMatcher.FindMatchesForLead(leadID)
	
	// Get behavioral score
	var score BehavioralScoreResult
	sao.db.Table("behavioral_scores").
		Where("lead_id = ?", leadID).
		Order("created_at DESC").
		First(&score)
	
	// Compile analysis
	analysis := map[string]interface{}{
		"lead_id":         leadID,
		"property_matches": matches,
		"behavioral_score": score,
		"analyzed_at":     time.Now(),
	}
	
	return analysis, nil
}

// StartAutomatedIntelligence starts the automated intelligence cycle (runs periodically)
func (sao *SpiderwebAIOrchestrator) StartAutomatedIntelligence(intervalMinutes int) {
	log.Printf("🤖 Starting automated intelligence cycle (every %d minutes)", intervalMinutes)
	
	ticker := time.NewTicker(time.Duration(intervalMinutes) * time.Minute)
	defer ticker.Stop()
	
	// Run immediately on start
	sao.RunIntelligenceCycle()
	
	// Then run on interval
	for range ticker.C {
		sao.RunIntelligenceCycle()
	}
}

// Helper functions

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// BehavioralScoreResult is a simplified score structure for queries
type BehavioralScoreResult struct {
	CompositeScore  int `json:"composite_score"`
	UrgencyScore    int `json:"urgency_score"`
	EngagementScore int `json:"engagement_score"`
	FinancialScore  int `json:"financial_score"`
}
