package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"uszn-gku-compare-and-edit/internal/domain"
)

//go:embed all:build/frontend
var assets embed.FS

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--cli" {
		runCLI(os.Args[2:])
		return
	}

	app := NewApp()
	err := wails.Run(&options.App{
		Title:  "DBF Compare",
		Width:  1280,
		Height: 900,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.startup,
		Bind: []any{
			app,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}

func runCLI(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: go run . --cli <previous.dbf> <current.dbf> [threshold_percent] [report.xlsx]")
		return
	}

	threshold := 20.0
	if len(args) >= 3 {
		parsed, err := strconv.ParseFloat(args[2], 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid threshold: %v\n", err)
			os.Exit(1)
		}
		threshold = parsed
	}

	app := NewApp()
	report, err := app.AnalyzeSupplierFiles(args[0], args[1], domain.AnalysisSettings{
		AmountChangePercent: threshold,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "analyze error: %v\n", err)
		os.Exit(1)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		fmt.Fprintf(os.Stderr, "print report: %v\n", err)
		os.Exit(1)
	}

	if len(args) >= 4 {
		outPath, err := filepath.Abs(args[3])
		if err != nil {
			fmt.Fprintf(os.Stderr, "resolve report path: %v\n", err)
			os.Exit(1)
		}
		if _, err := app.ExportAnalysisXLSX(report, outPath); err != nil {
			fmt.Fprintf(os.Stderr, "export error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("XLSX saved to %s\n", outPath)
	}
}
