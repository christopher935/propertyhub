package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
)

type RouteValidator struct {
	routes map[string]string
}

func NewRouteValidator() *RouteValidator {
	return &RouteValidator{
		routes: make(map[string]string),
	}
}

func (rv *RouteValidator) CheckRoute(method, path, location string) error {
	key := fmt.Sprintf("%s %s", strings.ToUpper(method), path)
	if existingLoc, exists := rv.routes[key]; exists {
		return fmt.Errorf("DUPLICATE ROUTE DETECTED: %s\n  First registered at: %s\n  Duplicate at: %s",
			key, existingLoc, location)
	}
	rv.routes[key] = location
	return nil
}

func ValidateGinRoutes(engine *gin.Engine) error {
	rv := NewRouteValidator()

	for _, route := range engine.Routes() {
		if err := rv.CheckRoute(route.Method, route.Path, route.Handler); err != nil {
			return err
		}
	}

	log.Printf("✅ Route validation passed: %d unique routes registered", len(rv.routes))
	return nil
}
