package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"chrisgross-ctrl-project/internal/config"
	"github.com/PuerkitoBio/goquery"
)

type ScraperService struct {
	config  *config.Config
	client  *http.Client
	apiKey  string
	baseURL string
}

type PropertyListing struct {
	ID           string            `json:"id"`
	URL          string            `json:"url"`
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	Price        string            `json:"price"`
	PriceFloat   float64           `json:"price_float"`
	Address      string            `json:"address"`
	City         string            `json:"city"`
	State        string            `json:"state"`
	ZipCode      string            `json:"zip_code"`
	Bedrooms     int               `json:"bedrooms"`
	Bathrooms    float32           `json:"bathrooms"`
	SquareFeet   int               `json:"square_feet"`
	Images       []string          `json:"images"`
	MLSId        string            `json:"mls_id"`
	PropertyType string            `json:"property_type"`
	YearBuilt    int               `json:"year_built"`
	Status       string            `json:"status"`
	ListingType  string            `json:"listing_type"`
	DaysOnMarket int               `json:"days_on_market"`
	Attributes   map[string]string `json:"attributes"`
}

type MLSSearchParams struct {
	Location     string   `json:"location"`
	MinPrice     int      `json:"min_price,omitempty"`
	MaxPrice     int      `json:"max_price,omitempty"`
	MinBedrooms  int      `json:"min_bedrooms,omitempty"`
	MaxBedrooms  int      `json:"max_bedrooms,omitempty"`
	MinBathrooms float32  `json:"min_bathrooms,omitempty"`
	MaxBathrooms float32  `json:"max_bathrooms,omitempty"`
	PropertyType []string `json:"property_type,omitempty"`
	Status       []string `json:"status,omitempty"`
	Limit        int      `json:"limit,omitempty"`
}

func NewScraperService(config *config.Config) *ScraperService {
	return &ScraperService{
		config:  config,
		client:  &http.Client{Timeout: 90 * time.Second},
		apiKey:  config.ScraperAPIKey,
		baseURL: "http://api.scraperapi.com",
	}
}

func (s *ScraperService) ScrapePropertyListings(targetURL string, params MLSSearchParams) ([]PropertyListing, error) {
	if s.apiKey == "" {
		return nil, fmt.Errorf("ScraperAPI key not configured")
	}

	log.Printf("🏠 Scraping properties from: %s", targetURL)

	html, err := s.fetchWithScraperAPI(targetURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch page: %v", err)
	}

	properties := s.extractPropertiesFromHTML(html, targetURL)
	filteredProperties := s.filterProperties(properties, params)

	log.Printf("✅ Scraped %d properties (%d after filtering)", len(properties), len(filteredProperties))

	return filteredProperties, nil
}

func (s *ScraperService) fetchWithScraperAPI(targetURL string) (string, error) {
	params := url.Values{}
	params.Add("api_key", s.apiKey)
	params.Add("url", targetURL)
	params.Add("country_code", "us")
	params.Add("render", "true")

	fullURL := fmt.Sprintf("%s?%s", s.baseURL, params.Encode())

	maxRetries := 3
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt*2) * time.Second)
			log.Printf("🔄 Retry attempt %d for %s", attempt+1, targetURL)
		}

		req, err := http.NewRequest("GET", fullURL, nil)
		if err != nil {
			lastErr = err
			continue
		}

		req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; PropertyHub/1.0)")

		resp, err := s.client.Do(req)
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

		log.Printf("✅ Received %d bytes of HTML", len(body))
		return string(body), nil
	}

	return "", lastErr
}

func (s *ScraperService) extractPropertiesFromHTML(html string, sourceURL string) []PropertyListing {
	properties := []PropertyListing{}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		log.Printf("⚠️  Failed to parse HTML: %v", err)
		return properties
	}

	expectedListingType := s.getExpectedListingType(sourceURL)
	log.Printf("🔍 Parsing JSON-LD for URL type: %s", expectedListingType)

	totalItems := 0
	placesFound := 0
	propertiesExtracted := 0
	filtered := map[string]int{
		"wrong_type":      0,
		"recent_showing":  0,
		"missing_data":    0,
	}

	doc.Find("script[type='application/ld+json']").Each(func(i int, script *goquery.Selection) {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(script.Text()), &data); err != nil {
			return
		}

		if graph, ok := data["@graph"].([]interface{}); ok {
			totalItems += len(graph)

			for _, item := range graph {
				if itemMap, ok := item.(map[string]interface{}); ok {

					isPlace := false
					if typeVal := itemMap["@type"]; typeVal != nil {
						switch v := typeVal.(type) {
						case string:
							isPlace = strings.Contains(v, "Place") || strings.Contains(v, "Residence")
						case []interface{}:
							for _, t := range v {
								if tStr, ok := t.(string); ok {
									if strings.Contains(tStr, "Place") || strings.Contains(tStr, "Residence") {
										isPlace = true
										break
									}
								}
							}
						}
					}

					if isPlace {
						placesFound++

						prop := s.extractPropertyFromJSONLD(itemMap, expectedListingType)
						if prop != nil {
							properties = append(properties, *prop)
							propertiesExtracted++
						} else {
							// Track why filtered
							listingType := s.getListingTypeFromItem(itemMap)
							if listingType == "Recent Showings" {
								filtered["recent_showing"]++
							} else if listingType != expectedListingType && expectedListingType != "Any" {
								filtered["wrong_type"]++
							} else {
								filtered["missing_data"]++
							}
						}
					}
				}
			}
		}
	})

	log.Printf("📊 JSON-LD Analysis for %s:", expectedListingType)
	log.Printf("   Total items: %d | Places: %d | Extracted: %d", totalItems, placesFound, propertiesExtracted)
	log.Printf("   Filtered: wrong_type=%d, showings=%d, missing_data=%d", 
		filtered["wrong_type"], filtered["recent_showing"], filtered["missing_data"])

	properties = s.enrichPropertiesWithPrices(doc, properties)

	log.Printf("📊 Found %d properties via JSON-LD", len(properties))
	return properties
}

func (s *ScraperService) getExpectedListingType(sourceURL string) string {
	lower := strings.ToLower(sourceURL)
	if strings.Contains(lower, "rent-by-agent") || strings.Contains(lower, "/rent") {
		return "For Rent"
	}
	if strings.Contains(lower, "forsale-by-agent") || strings.Contains(lower, "/forsale") {
		return "For Sale"
	}
	if strings.Contains(lower, "sold-by-agent") || strings.Contains(lower, "/sold") {
		return "Recently Sold"
	}
	if strings.Contains(lower, "rented-by-agent") || strings.Contains(lower, "/rented") {
		return "Recently Rented"
	}
	return "Any"
}

func (s *ScraperService) getListingTypeFromItem(data map[string]interface{}) string {
	if addProps, ok := data["additionalProperty"].([]interface{}); ok {
		for _, ap := range addProps {
			if apMap, ok := ap.(map[string]interface{}); ok {
				if name, _ := apMap["name"].(string); name == "listingType" {
					if val, ok := apMap["value"].(string); ok {
						return val
					}
				}
			}
		}
	}
	return ""
}

func (s *ScraperService) extractPropertyFromJSONLD(data map[string]interface{}, expectedType string) *PropertyListing {
	prop := &PropertyListing{}

	if name, ok := data["name"].(string); ok {
		prop.Title = name
	}

	if address, ok := data["address"].(map[string]interface{}); ok {
		prop.Address, _ = address["streetAddress"].(string)
		prop.City, _ = address["addressLocality"].(string)
		prop.State, _ = address["addressRegion"].(string)
		prop.ZipCode, _ = address["postalCode"].(string)
	}

	if beds, ok := data["numberOfBedrooms"].(float64); ok {
		prop.Bedrooms = int(beds)
	}

	if baths, ok := data["numberOfBathroomsTotal"].(float64); ok {
		prop.Bathrooms = float32(baths)
	}

	if floorSize, ok := data["floorSize"].(map[string]interface{}); ok {
		if sqft, ok := floorSize["value"].(float64); ok {
			prop.SquareFeet = int(sqft)
		}
	}

	if urlVal, ok := data["url"].(string); ok {
		prop.URL = urlVal
	}

	if images, ok := data["image"].([]interface{}); ok {
		for _, img := range images {
			if imgStr, ok := img.(string); ok {
				prop.Images = append(prop.Images, imgStr)
			}
		}
	}

	if addProps, ok := data["additionalProperty"].([]interface{}); ok {
		for _, addProp := range addProps {
			if propMap, ok := addProp.(map[string]interface{}); ok {
				name, _ := propMap["name"].(string)
				value := propMap["value"]

				switch name {
				case "MLS Number":
					prop.MLSId, _ = value.(string)
				case "listingType":
					prop.ListingType, _ = value.(string)
				case "Days on Market":
					if days, ok := value.(float64); ok {
						prop.DaysOnMarket = int(days)
					}
				case "Year Built":
					if year, ok := value.(float64); ok {
						prop.YearBuilt = int(year)
					}
				case "Property Type":
					prop.PropertyType, _ = value.(string)
				case "Square Footage":
					if sqftStr, ok := value.(string); ok {
						cleaned := strings.ReplaceAll(sqftStr, ",", "")
						fmt.Sscanf(cleaned, "%d", &prop.SquareFeet)
					}
				}
			}
		}
	}

	// FILTER 1: Skip "Recent Showings" (not actual listings)
	if prop.ListingType == "Recent Showings" {
		return nil
	}

	// FILTER 2: Only import properties matching the URL we're scraping
	if expectedType != "Any" && prop.ListingType != expectedType {
		return nil
	}

	// FILTER 3: Must have minimum data
	if prop.Address == "" || prop.ListingType == "" {
		return nil
	}

	// Map listing type to status
	switch prop.ListingType {
	case "For Rent":
		prop.Status = "for_lease"
	case "For Sale":
		prop.Status = "for_sale"
	case "Recently Sold":
		prop.Status = "sold"
	case "Recently Rented":
		prop.Status = "rented"
	default:
		prop.Status = "active"
	}

	return prop
}

func (s *ScraperService) enrichPropertiesWithPrices(doc *goquery.Document, properties []PropertyListing) []PropertyListing {
	pricesAdded := 0

	doc.Find("script[type='application/ld+json']").Each(func(i int, script *goquery.Selection) {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(script.Text()), &data); err != nil {
			return
		}

		if graph, ok := data["@graph"].([]interface{}); ok {
			for _, item := range graph {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if typeVal, ok := itemMap["@type"].(string); ok && typeVal == "Product" {
						name, _ := itemMap["name"].(string)

						for j := range properties {
							if strings.Contains(name, properties[j].Address) {
								if offers, ok := itemMap["offers"].(map[string]interface{}); ok {
									if price, ok := offers["price"].(float64); ok {
										properties[j].PriceFloat = price
										properties[j].Price = fmt.Sprintf("$%.0f", price)
										pricesAdded++
									}
								}
								break
							}
						}
					}
				}
			}
		}
	})

	log.Printf("💰 Added prices to %d properties", pricesAdded)
	return properties
}

func (s *ScraperService) filterProperties(properties []PropertyListing, params MLSSearchParams) []PropertyListing {
	var filtered []PropertyListing

	for _, property := range properties {
		if params.MinPrice > 0 && property.PriceFloat < float64(params.MinPrice) {
			continue
		}
		if params.MaxPrice > 0 && property.PriceFloat > float64(params.MaxPrice) {
			continue
		}
		if params.MinBedrooms > 0 && property.Bedrooms < params.MinBedrooms {
			continue
		}
		if params.MaxBedrooms > 0 && property.Bedrooms > params.MaxBedrooms {
			continue
		}

		filtered = append(filtered, property)

		if params.Limit > 0 && len(filtered) >= params.Limit {
			break
		}
	}

	return filtered
}

func (s *ScraperService) ValidateURL(targetURL string) error {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %v", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL must use http or https scheme")
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("URL must have a valid host")
	}

	return nil
}

func (s *ScraperService) GetScrapeStatus() map[string]interface{} {
	return map[string]interface{}{
		"service":     "scraper",
		"api_key_set": s.apiKey != "",
		"base_url":    s.baseURL,
		"timeout":     s.client.Timeout.Seconds(),
		"timestamp":   time.Now().Unix(),
	}
}
