package dbf

import (
	"path/filepath"
	"testing"
)

func TestConvertNovatekCSVToDBFProducesReadableCharges(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "novatek.dbf")

	err := ConvertNovatekCSVToDBF("../../test_data/923260405_0003.csv", outputPath)
	if err != nil {
		t.Fatalf("convert csv to dbf: %v", err)
	}

	records, err := ReadCharges(outputPath)
	if err != nil {
		t.Fatalf("read converted dbf: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("unexpected record count: %d", len(records))
	}

	record := records[0]
	if record.Postav != "НОВАТЭК-Челябинск" {
		t.Fatalf("unexpected provider: %q", record.Postav)
	}
	if record.City != "Новогорный" {
		t.Fatalf("unexpected city: %q", record.City)
	}
	if record.SocrCity != "п" {
		t.Fatalf("unexpected city prefix: %q", record.SocrCity)
	}
	if record.Dom != "3" {
		t.Fatalf("unexpected house number: %q", record.Dom)
	}
	if record.NameUsl != "Плита с ГВС или эл. водонагрев. (без отопления)" {
		t.Fatalf("unexpected service name: %q", record.NameUsl)
	}
	if record.Tarif != 0 {
		t.Fatalf("unexpected tariff: %v", record.Tarif)
	}
	if record.Nachisl != 0 {
		t.Fatalf("unexpected accrual: %v", record.Nachisl)
	}
}
