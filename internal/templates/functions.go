package templates

import (
	"fmt"
	"html/template"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func addCommas(n int) string {
	s := strconv.Itoa(n)
	if n < 0 {
		return "-" + addCommas(-n)
	}
	if len(s) <= 3 {
		return s
	}
	return addCommas(n/1000) + "," + s[len(s)-3:]
}

func addCommasFloat(f float64) string {
	intPart := int(f)
	fracPart := f - float64(intPart)
	result := addCommas(intPart)
	if fracPart > 0 {
		result += fmt.Sprintf(".%02d", int(fracPart*100))
	}
	return result
}

// GetFuncMap returns all template functions for use in Gin templates
func GetFuncMap() template.FuncMap {
	return template.FuncMap{
		// HTML/CSS/URL safety functions
		"safeHTML": func(html string) template.HTML {
			return template.HTML(html)
		},
		"safeCSS": func(css string) template.CSS {
			return template.CSS(css)
		},
		"safeURL": func(rawURL string) template.URL {
			if u, err := url.Parse(rawURL); err == nil {
				return template.URL(u.String())
			}
			return template.URL("")
		},
		"urlEncode": func(s string) string {
			return url.QueryEscape(s)
		},

		// String manipulation functions
		"upper": func(s string) string {
			return strings.ToUpper(s)
		},
		"lower": func(s string) string {
			return strings.ToLower(s)
		},
		"title": func(s string) string {
			return strings.Title(s)
		},
		"substr": func(s string, start, length int) string {
			if start < 0 || start >= len(s) {
				return ""
			}
			end := start + length
			if end > len(s) {
				end = len(s)
			}
			return s[start:end]
		},
		"truncate": func(s string, length int) string {
			if len(s) <= length {
				return s
			}
			return s[:length] + "..."
		},
		"replace": func(s, old, new string) string {
			return strings.ReplaceAll(s, old, new)
		},
		"split": func(s, sep string) []string {
			return strings.Split(s, sep)
		},
		"join": func(arr []string, sep string) string {
			return strings.Join(arr, sep)
		},
		"contains": func(s, substr string) bool {
			return strings.Contains(s, substr)
		},
		"hasPrefix": func(s, prefix string) bool {
			return strings.HasPrefix(s, prefix)
		},
		"hasSuffix": func(s, suffix string) bool {
			return strings.HasSuffix(s, suffix)
		},
		"trim": func(s string) string {
			return strings.TrimSpace(s)
		},

		// Number formatting functions
		"formatPrice": func(price interface{}) string {
			switch v := price.(type) {
			case int:
				return "$" + addCommas(v)
			case int64:
				return "$" + addCommas(int(v))
			case float64:
				return "$" + addCommasFloat(v)
			case float32:
				return "$" + addCommasFloat(float64(v))
			case string:
				if f, err := strconv.ParseFloat(v, 64); err == nil {
					return "$" + addCommasFloat(f)
				}
				return v
			default:
				return fmt.Sprintf("$%v", v)
			}
		},
		"formatNumber": func(n interface{}) string {
			switch v := n.(type) {
			case int:
				return addCommas(v)
			case int64:
				return addCommas(int(v))
			case float64:
				return addCommasFloat(v)
			case float32:
				return addCommasFloat(float64(v))
			default:
				return fmt.Sprintf("%v", v)
			}
		},
		"formatPercent": func(n interface{}) string {
			switch v := n.(type) {
			case float64:
				return fmt.Sprintf("%.1f%%", v)
			case float32:
				return fmt.Sprintf("%.1f%%", v)
			case int:
				return fmt.Sprintf("%d%%", v)
			default:
				return fmt.Sprintf("%v%%", v)
			}
		},
		"add": func(a, b int) int {
			return a + b
		},
		"subtract": func(a, b int) int {
			return a - b
		},
		"multiply": func(a, b int) int {
			return a * b
		},
		"divide": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},

		// Date/Time formatting functions
		"formatDate": func(t interface{}) string {
			switch v := t.(type) {
			case time.Time:
				return v.Format("January 2, 2006")
			case *time.Time:
				if v != nil {
					return v.Format("January 2, 2006")
				}
				return ""
			default:
				return fmt.Sprintf("%v", v)
			}
		},
		"formatDateTime": func(t interface{}) string {
			switch v := t.(type) {
			case time.Time:
				return v.Format("January 2, 2006 at 3:04 PM")
			case *time.Time:
				if v != nil {
					return v.Format("January 2, 2006 at 3:04 PM")
				}
				return ""
			default:
				return fmt.Sprintf("%v", v)
			}
		},
		"formatTime": func(t interface{}) string {
			switch v := t.(type) {
			case time.Time:
				return v.Format("3:04 PM")
			case *time.Time:
				if v != nil {
					return v.Format("3:04 PM")
				}
				return ""
			default:
				return fmt.Sprintf("%v", v)
			}
		},
		"formatShortDate": func(t interface{}) string {
			switch v := t.(type) {
			case time.Time:
				return v.Format("Jan 2, 2006")
			case *time.Time:
				if v != nil {
					return v.Format("Jan 2, 2006")
				}
				return ""
			default:
				return fmt.Sprintf("%v", v)
			}
		},
		"formatISO": func(t interface{}) string {
			switch v := t.(type) {
			case time.Time:
				return v.Format(time.RFC3339)
			case *time.Time:
				if v != nil {
					return v.Format(time.RFC3339)
				}
				return ""
			default:
				return ""
			}
		},
		"timeAgo": func(t interface{}) string {
			var tm time.Time
			switch v := t.(type) {
			case time.Time:
				tm = v
			case *time.Time:
				if v != nil {
					tm = *v
				} else {
					return ""
				}
			default:
				return ""
			}

			duration := time.Since(tm)
			if duration.Hours() < 1 {
				minutes := int(duration.Minutes())
				if minutes == 0 {
					return "just now"
				}
				return fmt.Sprintf("%d minutes ago", minutes)
			} else if duration.Hours() < 24 {
				hours := int(duration.Hours())
				return fmt.Sprintf("%d hours ago", hours)
			} else if duration.Hours() < 168 { // 7 days
				days := int(duration.Hours() / 24)
				return fmt.Sprintf("%d days ago", days)
			} else if duration.Hours() < 730 { // ~30 days
				weeks := int(duration.Hours() / 168)
				return fmt.Sprintf("%d weeks ago", weeks)
			} else {
				months := int(duration.Hours() / 730)
				return fmt.Sprintf("%d months ago", months)
			}
		},
		"currentYear": func() int {
			return time.Now().Year()
		},
		"now": func() time.Time {
			return time.Now()
		},

		// Array/Slice functions
		"len": func(arr interface{}) int {
			switch v := arr.(type) {
			case []interface{}:
				return len(v)
			case []string:
				return len(v)
			case []int:
				return len(v)
			default:
				return 0
			}
		},
		"first": func(arr interface{}) interface{} {
			switch v := arr.(type) {
			case []interface{}:
				if len(v) > 0 {
					return v[0]
				}
			case []string:
				if len(v) > 0 {
					return v[0]
				}
			}
			return nil
		},
		"last": func(arr interface{}) interface{} {
			switch v := arr.(type) {
			case []interface{}:
				if len(v) > 0 {
					return v[len(v)-1]
				}
			case []string:
				if len(v) > 0 {
					return v[len(v)-1]
				}
			}
			return nil
		},
		"slice": func(arr interface{}, start, end int) interface{} {
			switch v := arr.(type) {
			case []interface{}:
				if start < 0 || start >= len(v) {
					return []interface{}{}
				}
				if end > len(v) {
					end = len(v)
				}
				return v[start:end]
			case []string:
				if start < 0 || start >= len(v) {
					return []string{}
				}
				if end > len(v) {
					end = len(v)
				}
				return v[start:end]
			}
			return nil
		},

		// Comparison functions
		"eq": func(a, b interface{}) bool {
			return a == b
		},
		"ne": func(a, b interface{}) bool {
			return a != b
		},
		"lt": func(a, b int) bool {
			return a < b
		},
		"le": func(a, b int) bool {
			return a <= b
		},
		"gt": func(a, b int) bool {
			return a > b
		},
		"ge": func(a, b int) bool {
			return a >= b
		},

		// Logical functions
		"and": func(a, b bool) bool {
			return a && b
		},
		"or": func(a, b bool) bool {
			return a || b
		},
		"not": func(a bool) bool {
			return !a
		},

		// Default/fallback functions
		"default": func(def interface{}, val interface{}) interface{} {
			if val == nil || val == "" {
				return def
			}
			return val
		},
		"coalesce": func(vals ...interface{}) interface{} {
			for _, v := range vals {
				if v != nil && v != "" {
					return v
				}
			}
			return nil
		},

		// Property-specific functions
		"propertyStatus": func(status string) string {
			statusMap := map[string]string{
				"for_sale":   "For Sale",
				"for_rent":   "For Rent",
				"sold":       "Sold",
				"rented":     "Rented",
				"pending":    "Pending",
				"contingent": "Contingent",
				"active":     "Active",
				"inactive":   "Inactive",
			}
			if s, ok := statusMap[status]; ok {
				return s
			}
			return status
		},
		"propertyType": func(propType string) string {
			typeMap := map[string]string{
				"single_family": "Single Family Home",
				"condo":         "Condominium",
				"townhouse":     "Townhouse",
				"multi_family":  "Multi-Family",
				"land":          "Land",
				"commercial":    "Commercial",
			}
			if t, ok := typeMap[propType]; ok {
				return t
			}
			return propType
		},

		// Badge/Status color functions
		"statusColor": func(status string) string {
			colorMap := map[string]string{
				"active":     "green",
				"pending":    "yellow",
				"sold":       "blue",
				"rented":     "blue",
				"inactive":   "gray",
				"cancelled":  "red",
				"completed":  "green",
				"scheduled":  "blue",
				"confirmed":  "green",
				"unconfirmed": "yellow",
			}
			if c, ok := colorMap[strings.ToLower(status)]; ok {
				return c
			}
			return "gray"
		},
	}
}

