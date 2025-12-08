package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/flosch/pongo2/v6"
)

func main() {
	log.Println("🧪 Testing Pongo2 template parsing...")

	// Register filters
	registerFilters()

	templateDir := "web/templates"
	loader := pongo2.MustNewLocalFileSystemLoader(templateDir)
	templateSet := pongo2.NewSet("templates", loader)
	templateSet.Debug = true

	var errors []string
	var passed int

	err := filepath.Walk(templateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".html") {
			return nil
		}

		relPath, _ := filepath.Rel(templateDir, path)
		
		_, parseErr := templateSet.FromFile(relPath)
		if parseErr != nil {
			errors = append(errors, fmt.Sprintf("❌ %s: %v", relPath, parseErr))
		} else {
			passed++
			log.Printf("✅ %s", relPath)
		}

		return nil
	})

	if err != nil {
		log.Fatalf("Failed to walk templates: %v", err)
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("📊 Results: %d passed", passed)
	if len(errors) > 0 {
		fmt.Printf(", %d failed\n", len(errors))
		fmt.Println(strings.Repeat("=", 60))
		for _, e := range errors {
			fmt.Println(e)
		}
		os.Exit(1)
	} else {
		fmt.Println(", 0 failed")
		fmt.Println(strings.Repeat("=", 60))
		fmt.Println("🎉 All templates parsed successfully!")
	}
}

func registerFilters() {
	pongo2.RegisterFilter("formatPrice", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		switch v := in.Interface().(type) {
		case int:
			return pongo2.AsValue(fmt.Sprintf("$%s", humanize.Comma(int64(v)))), nil
		case int64:
			return pongo2.AsValue(fmt.Sprintf("$%s", humanize.Comma(v))), nil
		case float64:
			return pongo2.AsValue(fmt.Sprintf("$%s", humanize.CommafWithDigits(v, 2))), nil
		default:
			return pongo2.AsValue(fmt.Sprintf("%v", v)), nil
		}
	})

	pongo2.RegisterFilter("formatDate", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		if t, ok := in.Interface().(time.Time); ok {
			return pongo2.AsValue(t.Format("January 2, 2006")), nil
		}
		return in, nil
	})

	pongo2.RegisterFilter("formatNumber", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		switch v := in.Interface().(type) {
		case int:
			return pongo2.AsValue(humanize.Comma(int64(v))), nil
		case int64:
			return pongo2.AsValue(humanize.Comma(v)), nil
		case float64:
			return pongo2.AsValue(humanize.Commaf(v)), nil
		default:
			return pongo2.AsValue(fmt.Sprintf("%v", v)), nil
		}
	})

	pongo2.RegisterFilter("safe", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		return pongo2.AsSafeValue(in.String()), nil
	})
}
