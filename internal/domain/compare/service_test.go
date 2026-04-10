package compare

import (
	"errors"
	"os"
	"reflect"
	"testing"

	"uszn-gku-compare-and-edit/internal/dbf"
	"uszn-gku-compare-and-edit/internal/domain"
	"uszn-gku-compare-and-edit/internal/domain/aggregate"
)

func TestSnapshotsWithFixtureData(t *testing.T) {
	prevRecords, err := dbf.ReadCharges("../../../test_data/chrg_356_92_202601.dbf")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			t.Skip("fixture ../../../test_data/chrg_356_92_202601.dbf is not available")
		}
		t.Fatalf("read previous charges: %v", err)
	}
	currRecords, err := dbf.ReadCharges("../../../test_data/chrg_356_92_202602.dbf")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			t.Skip("fixture ../../../test_data/chrg_356_92_202602.dbf is not available")
		}
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
	if len(report.ServiceChanges) != 2 || len(report.ServiceChanges[0].HouseAddresses) != 1 {
		t.Fatalf("unexpected service house addresses: %+v", report.ServiceChanges)
	}
	if report.Summary.DisappearedServices != 0 {
		t.Fatalf("unexpected disappeared services: %d", report.Summary.DisappearedServices)
	}
	if report.Summary.AppearedHouses != 0 || report.Summary.DisappearedHouses != 0 {
		t.Fatalf("unexpected house changes: %+v", report.Summary)
	}
	if report.Summary.Anomalies != 0 {
		t.Fatalf("unexpected anomalies: %d", report.Summary.Anomalies)
	}
	if len(report.Anomalies) != 0 {
		t.Fatalf("unexpected anomaly details: %+v", report.Anomalies)
	}
}

func TestSortedHouseAddressesGroupsStreetAndSortsHouses(t *testing.T) {
	addresses := map[string]string{
		"1": "Г, ОЗЕРСК, ПРОЕЗД, КАЛИНИНА, д. 8",
		"2": "Г, ОЗЕРСК, ПРОЕЗД, КАЛИНИНА, д. 4",
		"3": "Г, ОЗЕРСК, ПРОЕЗД, КАЛИНИНА, д. 6",
		"4": "Г, ОЗЕРСК, УЛ, ЛЕНИНА, д. 12",
	}

	got := sortedHouseAddresses(addresses)
	want := []string{
		"Г, ОЗЕРСК, ПРОЕЗД, КАЛИНИНА, д. 4, 6, 8",
		"Г, ОЗЕРСК, УЛ, ЛЕНИНА, д. 12",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected grouped addresses:\nwant: %#v\ngot:  %#v", want, got)
	}
}

func TestSnapshotsMarksHouseAsDisappearedWhenServicesDropToTwentyPercent(t *testing.T) {
	prev := domain.ProviderSnapshot{
		Houses: map[string]domain.HouseAggregate{
			"house-1": {
				HouseKey:     "house-1",
				Address:      "Г, ОЗЕРСК, УЛ, ЛЕНИНА, д. 12",
				ServiceCount: 5,
				Services: map[string]domain.ServiceRef{
					"svc-1": {ServiceKey: "svc-1", NameUsl: "Услуга 1"},
					"svc-2": {ServiceKey: "svc-2", NameUsl: "Услуга 2"},
					"svc-3": {ServiceKey: "svc-3", NameUsl: "Услуга 3"},
					"svc-4": {ServiceKey: "svc-4", NameUsl: "Услуга 4"},
					"svc-5": {ServiceKey: "svc-5", NameUsl: "Услуга 5"},
				},
			},
		},
		Services:  map[string]domain.ServiceAggregate{},
		LineItems: map[string]domain.LineItemAggregate{},
	}
	curr := domain.ProviderSnapshot{
		Houses: map[string]domain.HouseAggregate{
			"house-1": {
				HouseKey:     "house-1",
				Address:      "Г, ОЗЕРСК, УЛ, ЛЕНИНА, д. 12",
				ServiceCount: 1,
				Services: map[string]domain.ServiceRef{
					"svc-1": {ServiceKey: "svc-1", NameUsl: "Услуга 1"},
				},
			},
		},
		Services:  map[string]domain.ServiceAggregate{},
		LineItems: map[string]domain.LineItemAggregate{},
	}

	report := Snapshots(prev, curr, domain.AnalysisSettings{AmountChangePercent: 20})

	if len(report.HouseChanges) != 1 {
		t.Fatalf("unexpected house changes count: %+v", report.HouseChanges)
	}
	if report.HouseChanges[0].Type != houseDisappearedByServiceCount {
		t.Fatalf("unexpected house change type: %+v", report.HouseChanges[0])
	}
	if report.Summary.DisappearedHouses != 1 {
		t.Fatalf("unexpected disappeared houses summary: %+v", report.Summary)
	}
}

func TestSnapshotsMarksHouseAsDisappearedWhenRecordCountDropsToTwentyPercent(t *testing.T) {
	prevRecords, err := dbf.ReadCharges("../../../test_data/9225010000_01.dbf")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			t.Skip("fixture ../../../test_data/9225010000_01.dbf is not available")
		}
		t.Fatalf("read previous charges: %v", err)
	}
	currRecords, err := dbf.ReadCharges("../../../test_data/9225010000_02.dbf")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			t.Skip("fixture ../../../test_data/9225010000_02.dbf is not available")
		}
		t.Fatalf("read current charges: %v", err)
	}

	report := Snapshots(
		aggregate.BuildSnapshot(prevRecords),
		aggregate.BuildSnapshot(currRecords),
		domain.AnalysisSettings{AmountChangePercent: 20},
	)

	if len(report.HouseChanges) != 1 {
		t.Fatalf("unexpected house changes count: %+v", report.HouseChanges)
	}
	if report.HouseChanges[0].Type != houseDisappearedByServiceCount {
		t.Fatalf("unexpected house change type: %+v", report.HouseChanges[0])
	}
	if report.HouseChanges[0].Address != "г, Озерск, мкр, Заозерный, д. 1" {
		t.Fatalf("unexpected house address: %+v", report.HouseChanges[0])
	}
}
