package services

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// AssetMinifierService handles CSS and JS minification
type AssetMinifierService struct {
	sourceDir      string
	outputDir      string
	mutex          sync.RWMutex
	processedFiles map[string]time.Time

	// Statistics
	totalFiles    int
	minifiedFiles int
	bytesOriginal int64
	bytesMinified int64
}

// MinificationResult holds the result of minification
type MinificationResult struct {
	OriginalSize     int64
	MinifiedSize     int64
	CompressionRatio float64
	Success          bool
	Error            error
}

// NewAssetMinifierService creates a new asset minifier service
func NewAssetMinifierService(sourceDir, outputDir string) *AssetMinifierService {
	return &AssetMinifierService{
		sourceDir:      sourceDir,
		outputDir:      outputDir,
		processedFiles: make(map[string]time.Time),
	}
}

// MinifyCSS minifies CSS content
func (a *AssetMinifierService) MinifyCSS(content string) string {
	// Remove comments (/* ... */)
	commentPattern := regexp.MustCompile(`/\*[\s\S]*?\*/`)
	content = commentPattern.ReplaceAllString(content, "")

	// Remove unnecessary whitespace
	whitespacePattern := regexp.MustCompile(`\s+`)
	content = whitespacePattern.ReplaceAllString(content, " ")

	// Remove whitespace around specific characters
	content = strings.ReplaceAll(content, " {", "{")
	content = strings.ReplaceAll(content, "{ ", "{")
	content = strings.ReplaceAll(content, " }", "}")
	content = strings.ReplaceAll(content, "; ", ";")
	content = strings.ReplaceAll(content, ": ", ":")
	content = strings.ReplaceAll(content, ", ", ",")

	// Remove trailing semicolons before closing braces
	trailingSemicolon := regexp.MustCompile(`;\s*}`)
	content = trailingSemicolon.ReplaceAllString(content, "}")

	// Remove unnecessary quotes from font families and other properties
	unnecessaryQuotes := regexp.MustCompile(`"([a-zA-Z0-9\-_]+)"`)
	content = unnecessaryQuotes.ReplaceAllString(content, "$1")

	// Convert hex colors to shorter format where possible (#ffffff -> #fff)
	longHex := regexp.MustCompile(`#([a-fA-F0-9])([a-fA-F0-9])([a-fA-F0-9])([a-fA-F0-9])([a-fA-F0-9])([a-fA-F0-9])`)
	content = longHex.ReplaceAllStringFunc(content, func(match string) string {
		if len(match) == 7 && match[1] == match[2] && match[3] == match[4] && match[5] == match[6] {
			return "#" + string(match[1]) + string(match[3]) + string(match[5])
		}
		return match
	})

	// Remove leading zeros from decimal values (0.5 -> .5)
	leadingZeros := regexp.MustCompile(`\b0\.`)
	content = leadingZeros.ReplaceAllString(content, ".")

	// Remove units from zero values (0px -> 0)
	zeroUnits := regexp.MustCompile(`\b0(px|em|rem|%|vh|vw|pt|pc|in|cm|mm|ex|ch|vmin|vmax)\b`)
	content = zeroUnits.ReplaceAllString(content, "0")

	return strings.TrimSpace(content)
}

// MinifyJS minifies JavaScript content (basic minification)
func (a *AssetMinifierService) MinifyJS(content string) string {
	// Remove single-line comments (// ...)
	singleLineComments := regexp.MustCompile(`//.*?$`)
	content = singleLineComments.ReplaceAllString(content, "")

	// Remove multi-line comments (/* ... */)
	multiLineComments := regexp.MustCompile(`/\*[\s\S]*?\*/`)
	content = multiLineComments.ReplaceAllString(content, "")

	// Remove unnecessary whitespace (but preserve string literals)
	lines := strings.Split(content, "\n")
	var minifiedLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			minifiedLines = append(minifiedLines, line)
		}
	}

	content = strings.Join(minifiedLines, "\n")

	// Remove extra spaces around operators and punctuation
	content = regexp.MustCompile(`\s*([{}();,=+\-*/<>!&|])\s*`).ReplaceAllString(content, "$1")

	// Remove newlines where safe
	content = strings.ReplaceAll(content, ";\n", ";")
	content = strings.ReplaceAll(content, "{\n", "{")
	content = strings.ReplaceAll(content, "\n}", "}")

	return strings.TrimSpace(content)
}

// MinifyFile minifies a single file
func (a *AssetMinifierService) MinifyFile(inputPath string) (*MinificationResult, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	// Read input file
	content, err := os.ReadFile(inputPath)
	if err != nil {
		return &MinificationResult{Success: false, Error: err}, err
	}

	originalSize := int64(len(content))
	a.totalFiles++

	// Get file extension
	ext := strings.ToLower(filepath.Ext(inputPath))

	var minifiedContent string
	switch ext {
	case ".css":
		minifiedContent = a.MinifyCSS(string(content))
	case ".js":
		minifiedContent = a.MinifyJS(string(content))
	default:
		// For non-minifiable files, just copy
		minifiedContent = string(content)
	}

	minifiedSize := int64(len(minifiedContent))

	// Calculate compression ratio
	compressionRatio := float64(0)
	if originalSize > 0 {
		compressionRatio = (float64(originalSize-minifiedSize) / float64(originalSize)) * 100
	}

	// Generate output path
	relPath, err := filepath.Rel(a.sourceDir, inputPath)
	if err != nil {
		return &MinificationResult{Success: false, Error: err}, err
	}

	// Add .min suffix for minified files
	if ext == ".css" || ext == ".js" {
		name := strings.TrimSuffix(filepath.Base(relPath), ext)
		dir := filepath.Dir(relPath)
		relPath = filepath.Join(dir, name+".min"+ext)
	}

	outputPath := filepath.Join(a.outputDir, relPath)

	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return &MinificationResult{Success: false, Error: err}, err
	}

	// Write minified file
	if err := os.WriteFile(outputPath, []byte(minifiedContent), 0644); err != nil {
		return &MinificationResult{Success: false, Error: err}, err
	}

	// Update statistics
	a.processedFiles[inputPath] = time.Now()
	a.minifiedFiles++
	a.bytesOriginal += originalSize
	a.bytesMinified += minifiedSize

	return &MinificationResult{
		OriginalSize:     originalSize,
		MinifiedSize:     minifiedSize,
		CompressionRatio: compressionRatio,
		Success:          true,
	}, nil
}

// MinifyDirectory minifies all CSS and JS files in a directory
func (a *AssetMinifierService) MinifyDirectory() error {
	return filepath.Walk(a.sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process CSS and JS files
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".css" && ext != ".js" {
			return nil
		}

		// Skip already minified files
		if strings.Contains(filepath.Base(path), ".min.") {
			return nil
		}

		// Check if file was recently processed
		if lastProcessed, exists := a.processedFiles[path]; exists {
			if info.ModTime().Before(lastProcessed) {
				return nil // File hasn't been modified since last processing
			}
		}

		// Minify file
		result, err := a.MinifyFile(path)
		if err != nil {
			fmt.Printf("Warning: Failed to minify %s: %v\n", path, err)
			return nil // Continue processing other files
		}

		if result.Success {
			fmt.Printf("✅ Minified %s: %d bytes -> %d bytes (%.1f%% reduction)\n",
				filepath.Base(path), result.OriginalSize, result.MinifiedSize, result.CompressionRatio)
		}

		return nil
	})
}

// GetStats returns minification statistics
func (a *AssetMinifierService) GetStats() map[string]interface{} {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	totalCompressionRatio := float64(0)
	if a.bytesOriginal > 0 {
		totalCompressionRatio = (float64(a.bytesOriginal-a.bytesMinified) / float64(a.bytesOriginal)) * 100
	}

	return map[string]interface{}{
		"total_files":             a.totalFiles,
		"minified_files":          a.minifiedFiles,
		"bytes_original":          a.bytesOriginal,
		"bytes_minified":          a.bytesMinified,
		"bytes_saved":             a.bytesOriginal - a.bytesMinified,
		"total_compression_ratio": totalCompressionRatio,
		"processed_files":         len(a.processedFiles),
	}
}

// WatchDirectory starts watching the source directory for changes
func (a *AssetMinifierService) WatchDirectory() error {
	// This is a simplified version - in production, use fsnotify for real file watching
	ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
	defer ticker.Stop()

	fmt.Println("📦 Asset minifier watching directory:", a.sourceDir)

	for {
		select {
		case <-ticker.C:
			if err := a.MinifyDirectory(); err != nil {
				fmt.Printf("Error during directory minification: %v\n", err)
			}
		}
	}
}

// CleanupOldFiles removes minified files that no longer have source files
func (a *AssetMinifierService) CleanupOldFiles() error {
	return filepath.Walk(a.outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only check .min.css and .min.js files
		name := filepath.Base(path)
		if !strings.Contains(name, ".min.") {
			return nil
		}

		// Determine source file path
		ext := filepath.Ext(path)
		sourceName := strings.ReplaceAll(name, ".min"+ext, ext)
		relPath, err := filepath.Rel(a.outputDir, filepath.Dir(path))
		if err != nil {
			return err
		}

		sourcePath := filepath.Join(a.sourceDir, relPath, sourceName)

		// Check if source file exists
		if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
			fmt.Printf("🗑️  Removing orphaned minified file: %s\n", path)
			return os.Remove(path)
		}

		return nil
	})
}

// CreateAssetManifest creates a manifest of all processed assets
func (a *AssetMinifierService) CreateAssetManifest() error {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	manifest := make(map[string]interface{})
	manifest["generated_at"] = time.Now().Format(time.RFC3339)
	manifest["stats"] = a.GetStats()

	files := make(map[string]map[string]interface{})
	for filePath, processedTime := range a.processedFiles {
		relPath, err := filepath.Rel(a.sourceDir, filePath)
		if err != nil {
			continue
		}

		files[relPath] = map[string]interface{}{
			"processed_at": processedTime.Format(time.RFC3339),
			"has_minified": true,
		}
	}
	manifest["files"] = files

	// Write manifest to output directory
	manifestPath := filepath.Join(a.outputDir, "asset-manifest.json")
	manifestContent := fmt.Sprintf(`{
		"generated_at": "%s",
		"stats": %+v,
		"files": %+v
	}`, manifest["generated_at"], manifest["stats"], manifest["files"])

	return os.WriteFile(manifestPath, []byte(manifestContent), 0644)
}
