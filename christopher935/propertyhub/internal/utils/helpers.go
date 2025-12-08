package utils

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

// GenerateID generates a unique ID string
func GenerateID() uint {
	// Generate a random uint for ID
	rand.Seed(time.Now().UnixNano())
	return uint(rand.Uint32())
}

// GenerateBookingReference generates a unique booking reference string
func GenerateBookingReference() string {
	now := time.Now()
	rand.Seed(now.UnixNano())
	randomNum := rand.Intn(9999)
	return fmt.Sprintf("SHW-%s-%04d", now.Format("20060102"), randomNum)
}

// ConvertToJSON converts any value to JSON string
func ConvertToJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(data)
}
