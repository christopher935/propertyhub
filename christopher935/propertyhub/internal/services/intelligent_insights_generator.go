package services

import (
	"fmt"
	"log"
	"time"
	"gorm.io/gorm"
)

type IntelligentInsightsGenerator struct {
	db *gorm.DB
}

type ContextualInsight struct {
	ID          string                      `json:"id"`
	Type        string                      `json:"type"`
	Icon        string                      `json:"icon"`
	Title       string                      `json:"title"`
	Message     string                      `json:"message"`
	Priority    int                         `json:"priority"`
	Actions     []ContextualInsightAction   `json:"actions"`
	Data        map[string]interface{}      `json:"data"`
	GeneratedAt time.Time                   `json:"generated_at"`
}

type ContextualInsightAction struct {
	Label  string                 `json:"label"`
	Action string                 `json:"action"`
	Params map[string]interface{} `json:"params"`
	Style  string                 `json:"style"`
}

func NewIntelligentInsightsGenerator(db *gorm.DB) *IntelligentInsightsGenerator {
	return &IntelligentInsightsGenerator{db: db}
}

func (g *IntelligentInsightsGenerator) GenerateAllInsights() ([]ContextualInsight, error) {
	log.Println("🧠 Generating insights...")
	var insights []ContextualInsight
	
	hotLeads, _ := g.generateHotLeadInsights()
	insights = append(insights, hotLeads...)
	
	return insights, nil
}

func (g *IntelligentInsightsGenerator) generateHotLeadInsights() ([]ContextualInsight, error) {
	var insights []ContextualInsight
	
	type HotLead struct {
		LeadID uint
		Name   string
		Score  float64
	}
	
	var leads []HotLead
	g.db.Raw(`
		SELECT l.id as lead_id, COALESCE(l.name, 'Lead') as name, COALESCE(bs.score, 0) as score
		FROM leads l 
		LEFT JOIN behavioral_scores bs ON bs.lead_id = l.id
		WHERE l.status IN ('new', 'active')
		ORDER BY score DESC LIMIT 3
	`).Scan(&leads)
	
	for _, lead := range leads {
		if lead.Score < 70 {
			continue
		}
		
		insights = append(insights, ContextualInsight{
			ID:       fmt.Sprintf("hot-%d", lead.LeadID),
			Type:     "urgent",
			Icon:     "🔥",
			Title:    fmt.Sprintf("%s showing extreme intent", lead.Name),
			Message:  fmt.Sprintf("Score %.0f. Contact within 1 hour.", lead.Score),
			Priority: 10,
			Actions: []ContextualInsightAction{{
				Label:  "Contact Lead",
				Action: "contact_lead",
				Params: map[string]interface{}{"lead_id": lead.LeadID},
				Style:  "primary",
			}},
			Data:        map[string]interface{}{"lead_id": lead.LeadID},
			GeneratedAt: time.Now(),
		})
	}
	
	return insights, nil
}
