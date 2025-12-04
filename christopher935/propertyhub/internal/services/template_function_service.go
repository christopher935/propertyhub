package services

import (
	"fmt"
	"html/template"
	"strconv"
	"strings"
	"time"
	"net/url"

	"chrisgross-ctrl-project/internal/security"
	"github.com/gin-gonic/gin"
)

// TemplateFunctionService provides enterprise template function management
type TemplateFunctionService struct {
	templateSecurity *security.TemplateSecurity
}

// NewTemplateFunctionService creates a new template function service
func NewTemplateFunctionService() *TemplateFunctionService {
	return &TemplateFunctionService{
		templateSecurity: security.NewTemplateSecurity(),
	}
}

// RegisterWithGin registers all template functions with the Gin engine
func (tfs *TemplateFunctionService) RegisterWithGin(r *gin.Engine) error {
	// Start with existing secure functions from your security system
	funcMap := tfs.templateSecurity.GetSecureFuncMap()
	
	// Add enterprise real estate specific functions
	tfs.addEnterpriseRealEstateFunctions(funcMap)
	
	// Add missing functions that cause crashes
	tfs.addMissingCriticalFunctions(funcMap)
	
	// Add utility functions for PropertyHub
	tfs.addPropertyHubUtilities(funcMap)
	
	// Register with Gin
	r.SetFuncMap(funcMap)
	
	return nil
}

// addEnterpriseRealEstateFunctions adds real estate business functions
func (tfs *TemplateFunctionService) addEnterpriseRealEstateFunctions(funcMap template.FuncMap) {
	// Property pricing functions
	funcMap["formatRentPrice"] = func(rent interface{}) string {
		switch v := rent.(type) {
		case int:
			return fmt.Sprintf("$%,d/mo", v)
		case int64:
			return fmt.Sprintf("$%,d/mo", v)
		case float64:
			return fmt.Sprintf("$%,.0f/mo", v)
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return fmt.Sprintf("$%,.0f/mo", f)
			}
			return v + "/mo"
		default:
			return "$0/mo"
		}
	}
	
	funcMap["formatSalePrice"] = func(price interface{}) string {
		switch v := price.(type) {
		case int:
			return fmt.Sprintf("$%,d", v)
		case int64:
			return fmt.Sprintf("$%,d", v)
		case float64:
			return fmt.Sprintf("$%,.0f", v)
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return fmt.Sprintf("$%,.0f", f)
			}
			return v
		default:
			return "$0"
		}
	}
	
	// Property metrics functions
	funcMap["formatSquareFeet"] = func(sqft interface{}) string {
		switch v := sqft.(type) {
		case int:
			return fmt.Sprintf("%,d sq ft", v)
		case int64:
			return fmt.Sprintf("%,d sq ft", v)
		case float64:
			return fmt.Sprintf("%,.0f sq ft", v)
		default:
			return fmt.Sprintf("%v sq ft", v)
		}
	}
	
	funcMap["formatDaysOnMarket"] = func(days interface{}) string {
		switch v := days.(type) {
		case int:
			if v == 1 {
				return "1 day"
			}
			return fmt.Sprintf("%d days", v)
		case int64:
			if v == 1 {
				return "1 day"
			}
			return fmt.Sprintf("%d days", v)
		default:
			return fmt.Sprintf("%v days", v)
		}
	}
}

// addMissingCriticalFunctions adds the functions causing crashes
func (tfs *TemplateFunctionService) addMissingCriticalFunctions(funcMap template.FuncMap) {
	// formatDateTime - THE CRITICAL ONE causing crashes
	funcMap["formatDateTime"] = func(dt interface{}) string {
		switch v := dt.(type) {
		case time.Time:
			return v.Format("2006-01-02 15:04:05")
		case *time.Time:
			if v != nil {
				return v.Format("2006-01-02 15:04:05")
			}
			return ""
		case string:
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				return t.Format("2006-01-02 15:04:05")
			}
			if t, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
				return t.Format("2006-01-02 15:04:05")
			}
			return v
		default:
			return time.Now().Format("2006-01-02 15:04:05")
		}
	}
	
	// formatDate - Standard date formatting
	funcMap["formatDate"] = func(date interface{}) string {
		switch v := date.(type) {
		case time.Time:
			return v.Format("January 2, 2006")
		case *time.Time:
			if v != nil {
				return v.Format("January 2, 2006")
			}
			return ""
		case string:
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				return t.Format("January 2, 2006")
			}
			return v
		default:
			return time.Now().Format("January 2, 2006")
		}
	}
	
	// formatPrice - Override/enhance existing FormatPrice from security
	funcMap["formatPrice"] = func(price interface{}) string {
		switch v := price.(type) {
		case int:
			return fmt.Sprintf("$%,d", v)
		case int64:
			return fmt.Sprintf("$%,d", v)
		case float64:
			if v >= 1000000 {
				return fmt.Sprintf("$%.1fM", v/1000000)
			}
			return fmt.Sprintf("$%,.0f", v)
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return fmt.Sprintf("$%,.0f", f)
			}
			return v
		default:
			return "$0"
		}
	}
}

// addPropertyHubUtilities adds PropertyHub specific utilities
func (tfs *TemplateFunctionService) addPropertyHubUtilities(funcMap template.FuncMap) {
	// Current year (often needed in templates)
	funcMap["currentYear"] = func() int {
		return time.Now().Year()
	}
	
	// String utilities
	funcMap["upper"] = func(s string) string {
		return strings.ToUpper(s)
	}
	
	funcMap["lower"] = func(s string) string {
		return strings.ToLower(s)
	}
	
	funcMap["title"] = func(s string) string {
		return strings.Title(s)
	}
	
	// URL utilities  
	funcMap["urlEncode"] = func(s string) string {
		return url.QueryEscape(s)
	}
	
	funcMap["safeURL"] = func(rawURL string) template.URL {
		if u, err := url.Parse(rawURL); err == nil {
			return template.URL(u.String())
		}
		return template.URL("")
	}
	
	// HTML utilities
	funcMap["safeHTML"] = func(html string) template.HTML { 
		return template.HTML(html) 
	}
	
	funcMap["safeCSS"] = func(css string) template.CSS { 
		return template.CSS(css) 
	}
	
	// Business calculations
	funcMap["calculateROI"] = func(income, expenses interface{}) string {
		inc, _ := strconv.ParseFloat(fmt.Sprintf("%v", income), 64)
		exp, _ := strconv.ParseFloat(fmt.Sprintf("%v", expenses), 64)
		if exp > 0 {
			roi := ((inc - exp) / exp) * 100
			return fmt.Sprintf("%.1f%%", roi)
		}
		return "N/A"
	}
	
	funcMap["formatPercent"] = func(value interface{}) string {
		switch v := value.(type) {
		case float64:
			return fmt.Sprintf("%.1f%%", v)
		case int:
			return fmt.Sprintf("%d%%", v)
		default:
			return fmt.Sprintf("%v%%", v)
		}
	}
}
