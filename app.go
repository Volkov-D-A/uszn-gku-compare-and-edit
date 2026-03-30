package main

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"uszn-gku-compare-and-edit/internal/dbf"
	"uszn-gku-compare-and-edit/internal/domain"
	"uszn-gku-compare-and-edit/internal/domain/aggregate"
	"uszn-gku-compare-and-edit/internal/domain/compare"
	"uszn-gku-compare-and-edit/internal/export"
)

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) AnalyzeSupplierFiles(prevPath string, currPath string, settings domain.AnalysisSettings) (domain.AnalysisReport, error) {
	prevRecords, err := dbf.ReadCharges(prevPath)
	if err != nil {
		return domain.AnalysisReport{}, fmt.Errorf("read previous DBF: %w", err)
	}

	currRecords, err := dbf.ReadCharges(currPath)
	if err != nil {
		return domain.AnalysisReport{}, fmt.Errorf("read current DBF: %w", err)
	}

	prevSnapshot := aggregate.BuildSnapshot(prevRecords)
	currSnapshot := aggregate.BuildSnapshot(currRecords)

	return compare.Snapshots(prevSnapshot, currSnapshot, settings), nil
}

func (a *App) ExportAnalysisXLSX(report domain.AnalysisReport, savePath string) (domain.ExportResult, error) {
	if err := export.WriteAnalysisXLSX(savePath, report); err != nil {
		return domain.ExportResult{}, fmt.Errorf("write xlsx: %w", err)
	}

	return domain.ExportResult{
		Path: savePath,
	}, nil
}

func (a *App) PickDBFFile() (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("application context is not ready")
	}

	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Выберите DBF файл",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "DBF files",
				Pattern:     "*.dbf",
			},
		},
	})
}

func (a *App) PickExportPath(defaultName string) (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("application context is not ready")
	}

	name := filepath.Base(defaultName)
	if filepath.Ext(name) == "" {
		name += ".xlsx"
	}

	return runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Сохранить XLSX отчёт",
		DefaultFilename: name,
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Excel files",
				Pattern:     "*.xlsx",
			},
		},
	})
}
