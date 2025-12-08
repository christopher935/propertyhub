package services

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
	
	"github.com/nfnt/resize"
)

// PhotoProcessingService handles concurrent image processing
type PhotoProcessingService struct {
	inputDir    string
	outputDir   string
	workerCount int
	jobQueue    chan PhotoJob
	resultQueue chan PhotoResult
	wg          sync.WaitGroup
	mutex       sync.RWMutex

	// Configuration
	maxWidth         int
	maxHeight        int
	quality          int
	generateOptimized bool
	generateThumbs   bool

	// Statistics
	totalJobs      int
	completedJobs  int
	failedJobs     int
	bytesProcessed int64
	bytesOptimized int64
}

// PhotoJob represents a single photo processing task
type PhotoJob struct {
	ID         string
	InputPath  string
	OutputPath string
	Operations []PhotoOperation
	Priority   int // 1 = high, 5 = low
}

// PhotoOperation defines image processing operations
type PhotoOperation struct {
	Type   string // "resize", "optimize", "convert", "thumbnail"
	Params map[string]interface{}
}

// PhotoResult contains the result of photo processing
type PhotoResult struct {
	JobID         string
	Success       bool
	Error         error
	OriginalSize  int64
	OptimizedSize int64
	OutputPaths   []string
	ProcessTime   time.Duration
}

// ImageSize represents image dimensions
type ImageSize struct {
	Width  int
	Height int
	Name   string // e.g., "thumbnail", "medium", "large"
}

// NewPhotoProcessingService creates a new photo processing service
func NewPhotoProcessingService(inputDir, outputDir string) *PhotoProcessingService {
	workerCount := runtime.NumCPU()
	if workerCount > 8 {
		workerCount = 8 // Limit to prevent excessive resource usage
	}

	return &PhotoProcessingService{
		inputDir:       inputDir,
		outputDir:      outputDir,
		workerCount:    workerCount,
		jobQueue:       make(chan PhotoJob, 100),
		resultQueue:    make(chan PhotoResult, 100),
		maxWidth:         1920,
		maxHeight:        1080,
		quality:          85,
		generateOptimized: true,
		generateThumbs:   true,
	}
}

// Start initializes and starts the photo processing workers
func (p *PhotoProcessingService) Start() {
	fmt.Printf("🖼️  Starting photo processing service with %d workers\n", p.workerCount)

	// Start workers
	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}

	// Start result processor
	go p.resultProcessor()

	fmt.Println("🎨 Photo processing workers started successfully")
}

// Stop gracefully shuts down the photo processing service
func (p *PhotoProcessingService) Stop() {
	fmt.Println("🛑 Stopping photo processing service...")
	close(p.jobQueue)
	p.wg.Wait()
	close(p.resultQueue)
	fmt.Println("✅ Photo processing service stopped")
}

// worker processes photos from the job queue
func (p *PhotoProcessingService) worker(id int) {
	defer p.wg.Done()

	for job := range p.jobQueue {
		startTime := time.Now()
		result := p.processPhoto(job)
		result.ProcessTime = time.Since(startTime)

		// Send result to result processor
		select {
		case p.resultQueue <- result:
		case <-time.After(5 * time.Second):
			fmt.Printf("⚠️  Worker %d: Result queue timeout for job %s\n", id, job.ID)
		}
	}
}

// resultProcessor handles processing results
func (p *PhotoProcessingService) resultProcessor() {
	for result := range p.resultQueue {
		p.mutex.Lock()
		if result.Success {
			p.completedJobs++
			p.bytesProcessed += result.OriginalSize
			p.bytesOptimized += result.OptimizedSize

			compressionRatio := float64(0)
			if result.OriginalSize > 0 {
				compressionRatio = (float64(result.OriginalSize-result.OptimizedSize) / float64(result.OriginalSize)) * 100
			}

			fmt.Printf("✅ Processed %s: %d bytes -> %d bytes (%.1f%% reduction) in %v\n",
				result.JobID, result.OriginalSize, result.OptimizedSize, compressionRatio, result.ProcessTime)
		} else {
			p.failedJobs++
			fmt.Printf("❌ Failed to process %s: %v\n", result.JobID, result.Error)
		}
		p.mutex.Unlock()
	}
}

// processPhoto processes a single photo job
func (p *PhotoProcessingService) processPhoto(job PhotoJob) PhotoResult {
	result := PhotoResult{
		JobID:       job.ID,
		Success:     false,
		OutputPaths: []string{},
	}

	// Read original image
	originalData, err := os.ReadFile(job.InputPath)
	if err != nil {
		result.Error = fmt.Errorf("failed to read input file: %v", err)
		return result
	}

	result.OriginalSize = int64(len(originalData))

	// Decode image
	img, format, err := p.decodeImage(strings.NewReader(string(originalData)))
	if err != nil {
		result.Error = fmt.Errorf("failed to decode image: %v", err)
		return result
	}

	// Process operations
	processedImg := img
	totalOptimizedSize := int64(0)

	for _, operation := range job.Operations {
		switch operation.Type {
		case "resize":
			if width, ok := operation.Params["width"].(int); ok {
				if height, ok := operation.Params["height"].(int); ok {
					processedImg = p.resizeImage(processedImg, width, height)
				}
			}
		case "optimize":
			// Optimization happens during encoding
		case "thumbnail":
			if err := p.generateThumbnail(processedImg, job.OutputPath); err != nil {
				fmt.Printf("⚠️  Failed to generate thumbnail: %v\n", err)
			}
		}
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(job.OutputPath), 0755); err != nil {
		result.Error = fmt.Errorf("failed to create output directory: %v", err)
		return result
	}

	// Save optimized image
	optimizedSize, err := p.saveOptimizedImage(processedImg, job.OutputPath, format)
	if err != nil {
		result.Error = fmt.Errorf("failed to save optimized image: %v", err)
		return result
	}

	result.OutputPaths = append(result.OutputPaths, job.OutputPath)
	totalOptimizedSize += optimizedSize

	// Generate optimized version if enabled (using high-quality JPEG compression)
	if p.generateOptimized && format != "jpeg" {
		optimizedPath := strings.TrimSuffix(job.OutputPath, filepath.Ext(job.OutputPath)) + "_optimized.jpg"
		optimizedSize, err := p.saveOptimizedJPEG(processedImg, optimizedPath)
		if err != nil {
			fmt.Printf("⚠️  Failed to generate optimized JPEG: %v\n", err)
		} else {
			result.OutputPaths = append(result.OutputPaths, optimizedPath)
			totalOptimizedSize += optimizedSize
		}
	}

	// Generate thumbnails if enabled
	if p.generateThumbs {
		thumbSizes := []ImageSize{
			{Width: 150, Height: 150, Name: "thumb"},
			{Width: 300, Height: 300, Name: "small"},
			{Width: 800, Height: 600, Name: "medium"},
		}

		for _, size := range thumbSizes {
			thumbPath := p.generateThumbPath(job.OutputPath, size.Name)
			thumbImg := p.resizeImage(processedImg, size.Width, size.Height)
			thumbSize, err := p.saveOptimizedImage(thumbImg, thumbPath, format)
			if err != nil {
				fmt.Printf("⚠️  Failed to generate %s thumbnail: %v\n", size.Name, err)
			} else {
				result.OutputPaths = append(result.OutputPaths, thumbPath)
				totalOptimizedSize += thumbSize
			}
		}
	}

	result.OptimizedSize = totalOptimizedSize
	result.Success = true
	return result
}

// AddJob adds a new photo processing job to the queue
func (p *PhotoProcessingService) AddJob(job PhotoJob) {
	p.mutex.Lock()
	p.totalJobs++
	p.mutex.Unlock()

	select {
	case p.jobQueue <- job:
		fmt.Printf("📥 Added job %s to queue (priority: %d)\n", job.ID, job.Priority)
	case <-time.After(5 * time.Second):
		fmt.Printf("⚠️  Job queue timeout for job %s\n", job.ID)
	}
}

// ProcessDirectory processes all images in a directory
func (p *PhotoProcessingService) ProcessDirectory(inputDir string) error {
	return filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Check if file is a supported image format
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" {
			return nil
		}

		// Generate output path
		relPath, err := filepath.Rel(inputDir, path)
		if err != nil {
			return err
		}
		outputPath := filepath.Join(p.outputDir, relPath)

		// Create job
		job := PhotoJob{
			ID:         filepath.Base(path),
			InputPath:  path,
			OutputPath: outputPath,
			Operations: []PhotoOperation{
				{Type: "resize", Params: map[string]interface{}{"width": p.maxWidth, "height": p.maxHeight}},
				{Type: "optimize", Params: map[string]interface{}{}},
			},
			Priority: 3, // Medium priority
		}

		if p.generateThumbs {
			job.Operations = append(job.Operations, PhotoOperation{Type: "thumbnail", Params: map[string]interface{}{}})
		}

		p.AddJob(job)
		return nil
	})
}

// decodeImage decodes an image from a reader
func (p *PhotoProcessingService) decodeImage(r io.Reader) (image.Image, string, error) {
	img, format, err := image.Decode(r)
	return img, format, err
}

// resizeImage resizes an image maintaining aspect ratio
func (p *PhotoProcessingService) resizeImage(img image.Image, maxWidth, maxHeight int) image.Image {
	bounds := img.Bounds()
	originalWidth := bounds.Dx()
	originalHeight := bounds.Dy()

	// Calculate aspect ratio
	aspectRatio := float64(originalWidth) / float64(originalHeight)

	var newWidth, newHeight int
	if originalWidth > originalHeight {
		newWidth = maxWidth
		newHeight = int(float64(maxWidth) / aspectRatio)
		if newHeight > maxHeight {
			newHeight = maxHeight
			newWidth = int(float64(maxHeight) * aspectRatio)
		}
	} else {
		newHeight = maxHeight
		newWidth = int(float64(maxHeight) * aspectRatio)
		if newWidth > maxWidth {
			newWidth = maxWidth
			newHeight = int(float64(maxWidth) / aspectRatio)
		}
	}

	// If no resizing needed, return original
	if newWidth >= originalWidth && newHeight >= originalHeight {
		return img
	}

	// Use Lanczos3 for high-quality resizing
	return resize.Resize(uint(newWidth), uint(newHeight), img, resize.Lanczos3)
}

// saveOptimizedImage saves an image with optimization
func (p *PhotoProcessingService) saveOptimizedImage(img image.Image, outputPath, format string) (int64, error) {
	file, err := os.Create(outputPath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	switch format {
	case "jpeg", "jpg":
		err = jpeg.Encode(file, img, &jpeg.Options{Quality: p.quality})
	case "png":
		err = png.Encode(file, img)
	default:
		err = jpeg.Encode(file, img, &jpeg.Options{Quality: p.quality})
	}

	if err != nil {
		return 0, err
	}

	// Get file size
	stat, err := file.Stat()
	if err != nil {
		return 0, err
	}

	return stat.Size(), nil
}

// saveOptimizedJPEG saves an image as high-quality optimized JPEG
func (p *PhotoProcessingService) saveOptimizedJPEG(img image.Image, outputPath string) (int64, error) {
	file, err := os.Create(outputPath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	// Use high-quality JPEG compression (90% quality for optimal balance)
	options := &jpeg.Options{
		Quality: 90,
	}

	err = jpeg.Encode(file, img, options)
	if err != nil {
		return 0, err
	}

	// Get file size
	stat, err := file.Stat()
	if err != nil {
		return 0, err
	}

	return stat.Size(), nil
}

// generateThumbnail creates a thumbnail for an image
func (p *PhotoProcessingService) generateThumbnail(img image.Image, basePath string) error {
	thumbPath := p.generateThumbPath(basePath, "thumb")
	thumbImg := p.resizeImage(img, 150, 150)
	_, err := p.saveOptimizedImage(thumbImg, thumbPath, "jpeg")
	return err
}

// generateThumbPath generates a thumbnail file path
func (p *PhotoProcessingService) generateThumbPath(originalPath, sizeName string) string {
	dir := filepath.Dir(originalPath)
	ext := filepath.Ext(originalPath)
	name := strings.TrimSuffix(filepath.Base(originalPath), ext)
	return filepath.Join(dir, "thumbs", name+"_"+sizeName+ext)
}

// GetStats returns processing statistics
func (p *PhotoProcessingService) GetStats() map[string]interface{} {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	compressionRatio := float64(0)
	if p.bytesProcessed > 0 {
		compressionRatio = (float64(p.bytesProcessed-p.bytesOptimized) / float64(p.bytesProcessed)) * 100
	}

	successRate := float64(0)
	totalProcessed := p.completedJobs + p.failedJobs
	if totalProcessed > 0 {
		successRate = (float64(p.completedJobs) / float64(totalProcessed)) * 100
	}

	return map[string]interface{}{
		"total_jobs":        p.totalJobs,
		"completed_jobs":    p.completedJobs,
		"failed_jobs":       p.failedJobs,
		"pending_jobs":      p.totalJobs - p.completedJobs - p.failedJobs,
		"success_rate":      successRate,
		"bytes_processed":   p.bytesProcessed,
		"bytes_optimized":   p.bytesOptimized,
		"bytes_saved":       p.bytesProcessed - p.bytesOptimized,
		"compression_ratio": compressionRatio,
		"worker_count":      p.workerCount,
		"queue_capacity":    cap(p.jobQueue),
		"settings": map[string]interface{}{
			"max_width":          p.maxWidth,
			"max_height":         p.maxHeight,
			"quality":            p.quality,
			"generate_optimized": p.generateOptimized,
			"generate_thumbs":    p.generateThumbs,
		},
	}
}
