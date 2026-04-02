package dbf

import (
	"testing"

	"uszn-gku-compare-and-edit/internal/domain"
	"uszn-gku-compare-and-edit/internal/domain/aggregate"
	"uszn-gku-compare-and-edit/internal/domain/compare"
)

func TestReadChargesSupportsAlternateFieldNames(t *testing.T) {
	records, err := ReadCharges("../../test_data/9225010000.dbf")
	if err != nil {
		t.Fatalf("read alternate DBF: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("unexpected record count: %d", len(records))
	}

	record := records[0]
	if record.Lshet != "195260841" {
		t.Fatalf("unexpected account: %q", record.Lshet)
	}
	if record.SocrCity != "г" {
		t.Fatalf("unexpected city prefix: %q", record.SocrCity)
	}
	if record.City != "Озерск" {
		t.Fatalf("unexpected city: %q", record.City)
	}
	if record.Tarif != 15.45 {
		t.Fatalf("unexpected tariff: %v", record.Tarif)
	}
	if record.Nachisl != 1419.86 {
		t.Fatalf("unexpected accrual: %v", record.Nachisl)
	}
}

func TestAlternateDBFCanGoThroughComparisonPipeline(t *testing.T) {
	records, err := ReadCharges("../../test_data/9225010000.dbf")
	if err != nil {
		t.Fatalf("read alternate DBF: %v", err)
	}

	snapshot := aggregate.BuildSnapshot(records)
	report := compare.Snapshots(snapshot, snapshot, domain.AnalysisSettings{AmountChangePercent: 20})

	if report.Meta.ProviderName == "" {
		t.Fatal("provider name should not be empty")
	}
	if report.Meta.CurrentRecords != 1 || report.Meta.PreviousRecords != 1 {
		t.Fatalf("unexpected record counts: %+v", report.Meta)
	}
	if report.Summary != (domain.SummaryCounts{}) {
		t.Fatalf("unexpected differences for identical snapshots: %+v", report.Summary)
	}
}
