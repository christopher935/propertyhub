package handlers

import (
	"fmt"
	"net/http"
	"sort"
	
	"chrisgross-ctrl-project/internal/models"
	"chrisgross-ctrl-project/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RecommendationsHandler struct {
	db                   *gorm.DB
	propertyMatching     *services.PropertyMatchingService
	behavioralScoring    *services.BehavioralScoringEngine
}

func NewRecommendationsHandler(db *gorm.DB, behavioralEngine *services.BehavioralScoringEngine) *RecommendationsHandler {
	return &RecommendationsHandler{
		db:                db,
		propertyMatching:  services.NewPropertyMatchingService(db),
		behavioralScoring: behavioralEngine,
	}
}

type PropertyRecommendation struct {
	Property         models.Property `json:"property"`
	Score            float64         `json:"score"`
	Reason           string          `json:"reason"`
	RecommendationType string        `json:"type"`
}

func (h *RecommendationsHandler) GetPersonalizedRecommendations(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		sessionID = c.GetString("session_id")
	}
	
	limit := 6
	if limitParam := c.Query("limit"); limitParam != "" {
		fmt.Sscanf(limitParam, "%d", &limit)
	}
	
	var recommendations []PropertyRecommendation
	
	if sessionID != "" {
		recommendations = h.getBehavioralRecommendations(sessionID, limit)
	}
	
	if len(recommendations) < limit {
		fallbackProps := h.getTrendingProperties(limit - len(recommendations))
		for _, prop := range fallbackProps {
			recommendations = append(recommendations, PropertyRecommendation{
				Property:           prop,
				Score:              80.0,
				Reason:             "Trending in your area",
				RecommendationType: "trending",
			})
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"recommendations": recommendations,
		"count":           len(recommendations),
		"personalized":    sessionID != "",
	})
}

func (h *RecommendationsHandler) getBehavioralRecommendations(sessionID string, limit int) []PropertyRecommendation {
	var recommendations []PropertyRecommendation
	
	var events []models.BehavioralEvent
	h.db.Where("session_id = ?", sessionID).
		Order("created_at DESC").
		Limit(50).
		Find(&events)
	
	if len(events) == 0 {
		return recommendations
	}
	
	preferences := h.extractUserPreferences(events)
	
	var properties []models.Property
	query := h.db.Where("status = ?", "active")
	
	if preferences["min_price"] != nil && preferences["max_price"] != nil {
		query = query.Where("price BETWEEN ? AND ?", preferences["min_price"], preferences["max_price"])
	}
	
	if preferences["preferred_city"] != nil {
		query = query.Where("city = ?", preferences["preferred_city"])
	}
	
	if preferences["property_type"] != nil {
		query = query.Where("property_type = ?", preferences["property_type"])
	}
	
	query.Order("created_at DESC").Limit(limit * 2).Find(&properties)
	
	scoredProperties := []struct {
		property models.Property
		score    float64
		reason   string
	}{}
	
	for _, prop := range properties {
		score, reason := h.scorePropertyForUser(prop, preferences, events)
		if score >= 60.0 {
			scoredProperties = append(scoredProperties, struct {
				property models.Property
				score    float64
				reason   string
			}{prop, score, reason})
		}
	}
	
	sort.Slice(scoredProperties, func(i, j int) bool {
		return scoredProperties[i].score > scoredProperties[j].score
	})
	
	for i, sp := range scoredProperties {
		if i >= limit {
			break
		}
		recommendations = append(recommendations, PropertyRecommendation{
			Property:           sp.property,
			Score:              sp.score,
			Reason:             sp.reason,
			RecommendationType: "personalized",
		})
	}
	
	return recommendations
}

func (h *RecommendationsHandler) extractUserPreferences(events []models.BehavioralEvent) map[string]interface{} {
	preferences := make(map[string]interface{})
	
	var totalPrice float64
	var priceCount int
	cityViews := make(map[string]int)
	typeViews := make(map[string]int)
	
	for _, event := range events {
		if event.EventName == "property_viewed" || event.EventName == "property_detail_view" {
			if price, ok := event.Properties["price"].(float64); ok {
				totalPrice += price
				priceCount++
			}
			
			if city, ok := event.Properties["city"].(string); ok {
				cityViews[city]++
			}
			
			if propType, ok := event.Properties["property_type"].(string); ok {
				typeViews[propType]++
			}
		}
	}
	
	if priceCount > 0 {
		avgPrice := totalPrice / float64(priceCount)
		preferences["min_price"] = avgPrice * 0.7
		preferences["max_price"] = avgPrice * 1.3
	}
	
	if len(cityViews) > 0 {
		maxViews := 0
		preferredCity := ""
		for city, views := range cityViews {
			if views > maxViews {
				maxViews = views
				preferredCity = city
			}
		}
		preferences["preferred_city"] = preferredCity
	}
	
	if len(typeViews) > 0 {
		maxViews := 0
		preferredType := ""
		for pType, views := range typeViews {
			if views > maxViews {
				maxViews = views
				preferredType = pType
			}
		}
		preferences["property_type"] = preferredType
	}
	
	return preferences
}

func (h *RecommendationsHandler) scorePropertyForUser(property models.Property, preferences map[string]interface{}, events []models.BehavioralEvent) (float64, string) {
	score := 50.0
	reasons := []string{}
	
	if prefCity, ok := preferences["preferred_city"].(string); ok && property.City == prefCity {
		score += 20.0
		reasons = append(reasons, fmt.Sprintf("In your preferred area: %s", prefCity))
	}
	
	if prefType, ok := preferences["property_type"].(string); ok && property.PropertyType == prefType {
		score += 15.0
		reasons = append(reasons, "Matches your preferred type")
	}
	
	if minPrice, ok := preferences["min_price"].(float64); ok {
		if maxPrice, ok2 := preferences["max_price"].(float64); ok2 {
			if property.Price >= minPrice && property.Price <= maxPrice {
				score += 15.0
				reasons = append(reasons, "In your price range")
			}
		}
	}
	
	if property.ViewCount > 50 {
		score += 10.0
		reasons = append(reasons, "Popular property")
	}
	
	reason := "Recommended for you"
	if len(reasons) > 0 {
		reason = reasons[0]
	}
	
	return score, reason
}

func (h *RecommendationsHandler) getTrendingProperties(limit int) []models.Property {
	var properties []models.Property
	h.db.Where("status = ?", "active").
		Order("view_count DESC, created_at DESC").
		Limit(limit).
		Find(&properties)
	return properties
}

func (h *RecommendationsHandler) GetSimilarProperties(c *gin.Context) {
	propertyID := c.Param("id")
	
	var property models.Property
	if err := h.db.First(&property, propertyID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Property not found"})
		return
	}
	
	var similarProperties []models.Property
	h.db.Where("city = ? AND id != ? AND price BETWEEN ? AND ? AND property_type = ?",
		property.City,
		property.ID,
		property.Price*0.8,
		property.Price*1.2,
		property.PropertyType,
	).Order("ABS(price - ?) ASC", property.Price).
		Limit(6).
		Find(&similarProperties)
	
	c.JSON(http.StatusOK, gin.H{
		"similar_properties": similarProperties,
		"count":              len(similarProperties),
	})
}
