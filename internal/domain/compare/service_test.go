package compare

import (
	"testing"

	"uszn-gku-compare-and-edit/internal/dbf"
	"uszn-gku-compare-and-edit/internal/domain"
	"uszn-gku-compare-and-edit/internal/domain/aggregate"
)

func TestSnapshotsWithFixtureData(t *testing.T) {
	prevRecords, err := dbf.ReadCharges("../../../test_data/chrg_356_92_202601.dbf")
	if err != nil {
		t.Fatalf("read previous charges: %v", err)
	}
	currRecords, err := dbf.ReadCharges("../../../test_data/chrg_356_92_202602.dbf")
	if err != nil {
		t.Fatalf("read current charges: %v", err)
	}

	report := Snapshots(
		aggregate.BuildSnapshot(prevRecords),
		aggregate.BuildSnapshot(currRecords),
		domain.AnalysisSettings{AmountChangePercent: 20},
	)

	if report.Meta.ProviderName != "ММПКХ" {
		t.Fatalf("unexpected provider: %q", report.Meta.ProviderName)
	}
	if report.Summary.TariffChanges != 0 {
		t.Fatalf("unexpected tariff changes: %d", report.Summary.TariffChanges)
	}
	if report.Summary.AppearedServices != 2 {
		t.Fatalf("unexpected appeared services: %d", report.Summary.AppearedServices)
	}
	if report.Summary.DisappearedServices != 0 {
		t.Fatalf("unexpected disappeared services: %d", report.Summary.DisappearedServices)
	}
	if report.Summary.AppearedHouses != 0 || report.Summary.DisappearedHouses != 0 {
		t.Fatalf("unexpected house changes: %+v", report.Summary)
	}
	if report.Summary.Anomalies != 1 {
		t.Fatalf("unexpected anomalies: %d", report.Summary.Anomalies)
	}
}
