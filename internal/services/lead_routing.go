package services

import (
	"fmt"
	"log"
	"sort"
	"time"
)

// LeadRoutingService provides intelligent lead routing (foundation for future scaling)
type LeadRoutingService struct {
	agents          map[string]*Agent
	routingRules    []RoutingRule
	workloadTracker *WorkloadTracker
	enabled         bool // Currently disabled for single agent setup
}

// Agent represents an agent in the system
type Agent struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	Email            string                 `json:"email"`
	Phone            string                 `json:"phone"`
	Active           bool                   `json:"active"`
	Specializations  []string               `json:"specializations"`  // luxury, investment, first_time_buyer, commercial
	GeographicAreas  []string               `json:"geographic_areas"` // zip codes or neighborhoods
	MaxDailyLeads    int                    `json:"max_daily_leads"`
	CurrentWorkload  int                    `json:"current_workload"`
	PerformanceScore float32                `json:"performance_score"` // 0.0 to 1.0
	Languages        []string               `json:"languages"`
	Availability     AgentAvailability      `json:"availability"`
	Preferences      map[string]interface{} `json:"preferences"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// AgentAvailability represents agent availability schedule
type AgentAvailability struct {
	Monday        []TimeSlot `json:"monday"`
	Tuesday       []TimeSlot `json:"tuesday"`
	Wednesday     []TimeSlot `json:"wednesday"`
	Thursday      []TimeSlot `json:"thursday"`
	Friday        []TimeSlot `json:"friday"`
	Saturday      []TimeSlot `json:"saturday"`
	Sunday        []TimeSlot `json:"sunday"`
	Timezone      string     `json:"timezone"`
	VacationDates []string   `json:"vacation_dates"` // YYYY-MM-DD format
}

// TimeSlot represents a time availability slot
type TimeSlot struct {
	StartTime string `json:"start_time"` // HH:MM format
	EndTime   string `json:"end_time"`   // HH:MM format
}

// RoutingRule defines how leads should be routed
type RoutingRule struct {
	ID         string             `json:"id"`
	Name       string             `json:"name"`
	Priority   int                `json:"priority"` // Higher number = higher priority
	Conditions []RoutingCondition `json:"conditions"`
	Action     RoutingAction      `json:"action"`
	Enabled    bool               `json:"enabled"`
	CreatedAt  time.Time          `json:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at"`
}

// RoutingCondition defines when a rule should trigger
type RoutingCondition struct {
	Field    string      `json:"field"`    // lead_score, property_price, location, etc.
	Operator string      `json:"operator"` // equals, greater_than, less_than, contains
	Value    interface{} `json:"value"`
}

// RoutingAction defines what to do when a rule matches
type RoutingAction struct {
	Type         string                 `json:"type"` // assign_agent, round_robin, skill_match, geographic_match
	Parameters   map[string]interface{} `json:"parameters"`
	FallbackType string                 `json:"fallback_type"` // What to do if primary action fails
}

// LeadRoutingRequest represents a lead routing request
type LeadRoutingRequest struct {
	LeadID            string                 `json:"lead_id"`
	LeadScore         float32                `json:"lead_score"`
	PropertyPrice     int                    `json:"property_price,omitempty"`
	PropertyType      string                 `json:"property_type,omitempty"`
	Location          string                 `json:"location,omitempty"`
	ZipCode           string                 `json:"zip_code,omitempty"`
	LeadSource        string                 `json:"lead_source"`
	PreferredLanguage string                 `json:"preferred_language,omitempty"`
	IsHighValue       bool                   `json:"is_high_value"`
	Urgency           string                 `json:"urgency"` // low, medium, high, urgent
	RequiredSkills    []string               `json:"required_skills,omitempty"`
	RequestedTime     time.Time              `json:"requested_time"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// LeadRoutingResult represents the routing decision
type LeadRoutingResult struct {
	AssignedAgentID   string    `json:"assigned_agent_id"`
	Agent             *Agent    `json:"agent,omitempty"`
	RoutingReason     string    `json:"routing_reason"`
	Confidence        float32   `json:"confidence"` // 0.0 to 1.0
	AlternativeAgents []string  `json:"alternative_agents,omitempty"`
	RoutingTime       time.Time `json:"routing_time"`
	RuleApplied       string    `json:"rule_applied,omitempty"`
}

// WorkloadTracker tracks agent workloads and performance
type WorkloadTracker struct {
	AgentWorkloads map[string]*AgentWorkload `json:"agent_workloads"`
	LastUpdated    time.Time                 `json:"last_updated"`
}

// AgentWorkload tracks an individual agent's workload
type AgentWorkload struct {
	AgentID           string    `json:"agent_id"`
	ActiveLeads       int       `json:"active_leads"`
	LeadsToday        int       `json:"leads_today"`
	ConversionsToday  int       `json:"conversions_today"`
	ResponseTime      float32   `json:"avg_response_time_minutes"`
	SatisfactionScore float32   `json:"satisfaction_score"`
	LastActivity      time.Time `json:"last_activity"`
}

// NewLeadRoutingService creates a new lead routing service
func NewLeadRoutingService() *LeadRoutingService {
	service := &LeadRoutingService{
		agents:       make(map[string]*Agent),
		routingRules: []RoutingRule{},
		workloadTracker: &WorkloadTracker{
			AgentWorkloads: make(map[string]*AgentWorkload),
			LastUpdated:    time.Now(),
		},
		enabled: false, // Disabled until multiple agents
	}

	// Initialize with default single agent
	service.initializeDefaultAgent()

	// Set up default routing rules (dormant until enabled)
	service.initializeDefaultRoutingRules()

	log.Printf("🎯 Lead routing service initialized (dormant - single agent mode)")
	return service
}

// RouteLeadToAgent routes a lead to the most appropriate agent
func (lrs *LeadRoutingService) RouteLeadToAgent(request LeadRoutingRequest) (*LeadRoutingResult, error) {
	// If routing is disabled (single agent), return the default agent
	if !lrs.enabled {
		return lrs.routeToDefaultAgent(request), nil
	}

	log.Printf("🔄 Routing lead %s with score %.2f", request.LeadID, request.LeadScore)

	// Find the best agent based on routing rules
	bestAgent, confidence, reason, ruleApplied := lrs.findBestAgent(request)

	if bestAgent == nil {
		return nil, fmt.Errorf("no available agents found for lead %s", request.LeadID)
	}

	// Update workload tracking
	lrs.updateAgentWorkload(bestAgent.ID, request)

	// Find alternative agents
	alternatives := lrs.findAlternativeAgents(request, bestAgent.ID)

	result := &LeadRoutingResult{
		AssignedAgentID:   bestAgent.ID,
		Agent:             bestAgent,
		RoutingReason:     reason,
		Confidence:        confidence,
		AlternativeAgents: alternatives,
		RoutingTime:       time.Now(),
		RuleApplied:       ruleApplied,
	}

	log.Printf("✅ Lead %s routed to agent %s (%s)", request.LeadID, bestAgent.Name, reason)
	return result, nil
}

// AddAgent adds a new agent to the routing system
func (lrs *LeadRoutingService) AddAgent(agent *Agent) {
	agent.CreatedAt = time.Now()
	agent.UpdatedAt = time.Now()
	lrs.agents[agent.ID] = agent

	// Initialize workload tracking
	lrs.workloadTracker.AgentWorkloads[agent.ID] = &AgentWorkload{
		AgentID:           agent.ID,
		ActiveLeads:       0,
		LeadsToday:        0,
		ConversionsToday:  0,
		ResponseTime:      15.0, // Default 15 minutes
		SatisfactionScore: 0.8,  // Default score
		LastActivity:      time.Now(),
	}

	// Enable routing if we now have multiple agents
	if len(lrs.agents) > 1 {
		lrs.enabled = true
		log.Printf("🚀 Lead routing enabled - multiple agents detected")
	}

	log.Printf("👤 Agent %s added to routing system", agent.Name)
}

// EnableRouting manually enables the routing system
func (lrs *LeadRoutingService) EnableRouting() {
	lrs.enabled = true
	log.Printf("🚀 Lead routing manually enabled")
}

// DisableRouting disables the routing system (fallback to default agent)
func (lrs *LeadRoutingService) DisableRouting() {
	lrs.enabled = false
	log.Printf("⏸️ Lead routing disabled - using default agent")
}

// Private methods

func (lrs *LeadRoutingService) initializeDefaultAgent() {
	defaultAgent := &Agent{
		ID:               "agent_1",
		Name:             "Primary Agent",
		Email:            "agent@elitepropertyshowings.com",
		Phone:            "(555) 123-4567",
		Active:           true,
		Specializations:  []string{"luxury", "investment", "first_time_buyer", "commercial"},
		GeographicAreas:  []string{"77001", "77002", "77003", "77004"}, // Houston zip codes
		MaxDailyLeads:    50,
		CurrentWorkload:  0,
		PerformanceScore: 0.95,
		Languages:        []string{"english", "spanish"},
		Availability:     lrs.getDefaultAvailability(),
		Preferences:      make(map[string]interface{}),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	lrs.agents[defaultAgent.ID] = defaultAgent

	// Initialize workload tracking
	lrs.workloadTracker.AgentWorkloads[defaultAgent.ID] = &AgentWorkload{
		AgentID:           defaultAgent.ID,
		ActiveLeads:       0,
		LeadsToday:        0,
		ConversionsToday:  0,
		ResponseTime:      10.0,
		SatisfactionScore: 0.95,
		LastActivity:      time.Now(),
	}
}

func (lrs *LeadRoutingService) initializeDefaultRoutingRules() {
	// High-value lead rule
	highValueRule := RoutingRule{
		ID:       "high_value_leads",
		Name:     "Route High-Value Leads to Top Performers",
		Priority: 100,
		Conditions: []RoutingCondition{
			{Field: "is_high_value", Operator: "equals", Value: true},
		},
		Action: RoutingAction{
			Type: "performance_match",
			Parameters: map[string]interface{}{
				"min_performance_score": 0.8,
			},
			FallbackType: "round_robin",
		},
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Luxury property rule
	luxuryRule := RoutingRule{
		ID:       "luxury_properties",
		Name:     "Route Luxury Properties to Specialists",
		Priority: 90,
		Conditions: []RoutingCondition{
			{Field: "property_price", Operator: "greater_than", Value: 750000},
		},
		Action: RoutingAction{
			Type: "skill_match",
			Parameters: map[string]interface{}{
				"required_skills": []string{"luxury"},
			},
			FallbackType: "performance_match",
		},
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Geographic routing rule
	geoRule := RoutingRule{
		ID:       "geographic_routing",
		Name:     "Route by Geographic Expertise",
		Priority: 70,
		Conditions: []RoutingCondition{
			{Field: "zip_code", Operator: "not_empty", Value: nil},
		},
		Action: RoutingAction{
			Type: "geographic_match",
			Parameters: map[string]interface{}{
				"prefer_local_expertise": true,
			},
			FallbackType: "round_robin",
		},
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	lrs.routingRules = []RoutingRule{highValueRule, luxuryRule, geoRule}

	// Sort by priority
	sort.Slice(lrs.routingRules, func(i, j int) bool {
		return lrs.routingRules[i].Priority > lrs.routingRules[j].Priority
	})
}

func (lrs *LeadRoutingService) routeToDefaultAgent(request LeadRoutingRequest) *LeadRoutingResult {
	defaultAgent := lrs.agents["agent_1"]

	return &LeadRoutingResult{
		AssignedAgentID:   defaultAgent.ID,
		Agent:             defaultAgent,
		RoutingReason:     "Single agent mode - routed to primary agent",
		Confidence:        1.0,
		AlternativeAgents: []string{},
		RoutingTime:       time.Now(),
		RuleApplied:       "default_single_agent",
	}
}

func (lrs *LeadRoutingService) findBestAgent(request LeadRoutingRequest) (*Agent, float32, string, string) {
	// Apply routing rules in priority order
	for _, rule := range lrs.routingRules {
		if !rule.Enabled {
			continue
		}

		if lrs.evaluateRuleConditions(rule.Conditions, request) {
			agent, confidence, reason := lrs.executeRoutingAction(rule.Action, request)
			if agent != nil {
				return agent, confidence, reason, rule.Name
			}
		}
	}

	// Fallback to round-robin if no rules match
	agent, confidence, reason := lrs.roundRobinRouting(request)
	return agent, confidence, reason, "fallback_round_robin"
}

func (lrs *LeadRoutingService) evaluateRuleConditions(conditions []RoutingCondition, request LeadRoutingRequest) bool {
	for _, condition := range conditions {
		if !lrs.evaluateCondition(condition, request) {
			return false
		}
	}
	return true
}

func (lrs *LeadRoutingService) evaluateCondition(condition RoutingCondition, request LeadRoutingRequest) bool {
	var fieldValue interface{}

	// Get field value from request
	switch condition.Field {
	case "lead_score":
		fieldValue = request.LeadScore
	case "property_price":
		fieldValue = request.PropertyPrice
	case "is_high_value":
		fieldValue = request.IsHighValue
	case "zip_code":
		fieldValue = request.ZipCode
	case "urgency":
		fieldValue = request.Urgency
	default:
		return false
	}

	// Evaluate based on operator
	switch condition.Operator {
	case "equals":
		return fieldValue == condition.Value
	case "greater_than":
		if fv, ok := fieldValue.(float32); ok {
			if cv, ok := condition.Value.(float64); ok {
				return fv > float32(cv)
			}
		}
		if fv, ok := fieldValue.(int); ok {
			if cv, ok := condition.Value.(float64); ok {
				return float64(fv) > cv
			}
		}
	case "less_than":
		if fv, ok := fieldValue.(float32); ok {
			if cv, ok := condition.Value.(float64); ok {
				return fv < float32(cv)
			}
		}
	case "not_empty":
		if str, ok := fieldValue.(string); ok {
			return str != ""
		}
	}

	return false
}

func (lrs *LeadRoutingService) executeRoutingAction(action RoutingAction, request LeadRoutingRequest) (*Agent, float32, string) {
	switch action.Type {
	case "performance_match":
		return lrs.performanceBasedRouting(request, action.Parameters)
	case "skill_match":
		return lrs.skillBasedRouting(request, action.Parameters)
	case "geographic_match":
		return lrs.geographicRouting(request, action.Parameters)
	case "round_robin":
		return lrs.roundRobinRouting(request)
	default:
		return lrs.roundRobinRouting(request)
	}
}

func (lrs *LeadRoutingService) performanceBasedRouting(request LeadRoutingRequest, params map[string]interface{}) (*Agent, float32, string) {
	minScore := 0.7 // Default minimum performance score
	if ms, ok := params["min_performance_score"].(float64); ok {
		minScore = ms
	}

	var bestAgent *Agent
	var bestScore float32

	for _, agent := range lrs.agents {
		if !agent.Active || agent.PerformanceScore < float32(minScore) {
			continue
		}

		if lrs.isAgentAvailable(agent) && agent.PerformanceScore > bestScore {
			bestAgent = agent
			bestScore = agent.PerformanceScore
		}
	}

	if bestAgent != nil {
		return bestAgent, 0.9, fmt.Sprintf("Performance-based routing (score: %.2f)", bestScore)
	}

	return nil, 0, "No agents meet performance criteria"
}

func (lrs *LeadRoutingService) skillBasedRouting(request LeadRoutingRequest, params map[string]interface{}) (*Agent, float32, string) {
	requiredSkills := []string{}
	if skills, ok := params["required_skills"].([]string); ok {
		requiredSkills = skills
	}

	for _, agent := range lrs.agents {
		if !agent.Active || !lrs.isAgentAvailable(agent) {
			continue
		}

		if lrs.agentHasSkills(agent, requiredSkills) {
			return agent, 0.85, fmt.Sprintf("Skill-based routing (skills: %v)", requiredSkills)
		}
	}

	return nil, 0, "No agents with required skills available"
}

func (lrs *LeadRoutingService) geographicRouting(request LeadRoutingRequest, params map[string]interface{}) (*Agent, float32, string) {
	if request.ZipCode == "" {
		return nil, 0, "No location information for geographic routing"
	}

	for _, agent := range lrs.agents {
		if !agent.Active || !lrs.isAgentAvailable(agent) {
			continue
		}

		if lrs.agentCoversArea(agent, request.ZipCode) {
			return agent, 0.8, fmt.Sprintf("Geographic routing (area: %s)", request.ZipCode)
		}
	}

	return nil, 0, "No agents cover the requested geographic area"
}

func (lrs *LeadRoutingService) roundRobinRouting(request LeadRoutingRequest) (*Agent, float32, string) {
	// Find agent with lowest current workload
	var bestAgent *Agent
	lowestWorkload := 9999

	for _, agent := range lrs.agents {
		if !agent.Active || !lrs.isAgentAvailable(agent) {
			continue
		}

		if agent.CurrentWorkload < lowestWorkload {
			bestAgent = agent
			lowestWorkload = agent.CurrentWorkload
		}
	}

	if bestAgent != nil {
		return bestAgent, 0.7, fmt.Sprintf("Round-robin routing (workload: %d)", lowestWorkload)
	}

	return nil, 0, "No available agents"
}

func (lrs *LeadRoutingService) findAlternativeAgents(request LeadRoutingRequest, excludeAgentID string) []string {
	var alternatives []string

	for agentID, agent := range lrs.agents {
		if agentID != excludeAgentID && agent.Active && lrs.isAgentAvailable(agent) {
			alternatives = append(alternatives, agentID)
		}
	}

	return alternatives
}

func (lrs *LeadRoutingService) updateAgentWorkload(agentID string, request LeadRoutingRequest) {
	if workload, exists := lrs.workloadTracker.AgentWorkloads[agentID]; exists {
		workload.ActiveLeads++
		workload.LeadsToday++
		workload.LastActivity = time.Now()
	}

	if agent, exists := lrs.agents[agentID]; exists {
		agent.CurrentWorkload++
		agent.UpdatedAt = time.Now()
	}

	lrs.workloadTracker.LastUpdated = time.Now()
}

// Helper methods

func (lrs *LeadRoutingService) isAgentAvailable(agent *Agent) bool {
	// Check if agent is under their daily lead limit
	if workload, exists := lrs.workloadTracker.AgentWorkloads[agent.ID]; exists {
		if workload.LeadsToday >= agent.MaxDailyLeads {
			return false
		}
	}

	// In production, would check actual availability schedule
	return true
}

func (lrs *LeadRoutingService) agentHasSkills(agent *Agent, requiredSkills []string) bool {
	for _, required := range requiredSkills {
		found := false
		for _, skill := range agent.Specializations {
			if skill == required {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (lrs *LeadRoutingService) agentCoversArea(agent *Agent, zipCode string) bool {
	for _, area := range agent.GeographicAreas {
		if area == zipCode {
			return true
		}
	}
	return false
}

func (lrs *LeadRoutingService) getDefaultAvailability() AgentAvailability {
	businessHours := []TimeSlot{{StartTime: "09:00", EndTime: "18:00"}}

	return AgentAvailability{
		Monday:        businessHours,
		Tuesday:       businessHours,
		Wednesday:     businessHours,
		Thursday:      businessHours,
		Friday:        businessHours,
		Saturday:      []TimeSlot{{StartTime: "10:00", EndTime: "16:00"}},
		Sunday:        []TimeSlot{},
		Timezone:      "America/Chicago", // Houston timezone
		VacationDates: []string{},
	}
}

// Public API methods for future use

// GetRoutingStatus returns the current status of the routing system
func (lrs *LeadRoutingService) GetRoutingStatus() map[string]interface{} {
	return map[string]interface{}{
		"enabled":       lrs.enabled,
		"agent_count":   len(lrs.agents),
		"active_agents": lrs.getActiveAgentCount(),
		"routing_rules": len(lrs.routingRules),
		"last_updated":  lrs.workloadTracker.LastUpdated,
	}
}

func (lrs *LeadRoutingService) getActiveAgentCount() int {
	count := 0
	for _, agent := range lrs.agents {
		if agent.Active {
			count++
		}
	}
	return count
}

// GetAgentWorkloads returns current agent workload information
func (lrs *LeadRoutingService) GetAgentWorkloads() map[string]*AgentWorkload {
	return lrs.workloadTracker.AgentWorkloads
}

// UpdateAgentPerformance updates an agent's performance metrics
func (lrs *LeadRoutingService) UpdateAgentPerformance(agentID string, score float32, responseTime float32) {
	if agent, exists := lrs.agents[agentID]; exists {
		agent.PerformanceScore = score
		agent.UpdatedAt = time.Now()
	}

	if workload, exists := lrs.workloadTracker.AgentWorkloads[agentID]; exists {
		workload.ResponseTime = responseTime
		workload.LastActivity = time.Now()
	}

	log.Printf("📊 Updated performance for agent %s: score=%.2f, response_time=%.1f", agentID, score, responseTime)
}
