package utils

// Mathematical utility functions
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// MinFloat returns the minimum of two float64 values
func MinFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// MaxFloat returns the maximum of two float64 values
func MaxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// Array/slice utility functions
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ContainsInt checks if an int slice contains a specific value
func ContainsInt(slice []int, item int) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
