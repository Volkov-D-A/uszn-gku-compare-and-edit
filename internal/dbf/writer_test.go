package dbf

import (
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestConvertNovatekCSVToDBFProducesReadableCharges(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "novatek.dbf")

	err := ConvertNovatekCSVToDBF("../../test_data/923260405_0003.csv", outputPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			t.Skip("fixture ../../test_data/923260405_0003.csv is not available")
		}
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

func TestRemoveRIRZeroRowsDropsRecordsWithEmptyTariffFields(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "rir-source.dbf")
	outputPath := filepath.Join(tempDir, "rir-filtered.dbf")

	source, err := buildRawDBF(dbfTable{
		Version:        dbfVersionDBaseIII,
		LanguageDriver: dbfLanguageDriverRU,
		Fields: []fieldDescriptor{
			{Name: "POSTAV", Type: 'C', Length: 20},
			{Name: "VID_USL", Type: 'C', Length: 20},
			{Name: "TARIF", Type: 'C', Length: 10},
			{Name: "TARIF_DEC", Type: 'C', Length: 10},
			{Name: "NORM_USL", Type: 'C', Length: 10},
		},
		Records: [][]byte{
			buildRawRecord(
				fieldDescriptor{Name: "POSTAV", Type: 'C', Length: 20},
				fieldDescriptor{Name: "VID_USL", Type: 'C', Length: 20},
				fieldDescriptor{Name: "TARIF", Type: 'C', Length: 10},
				fieldDescriptor{Name: "TARIF_DEC", Type: 'C', Length: 10},
				fieldDescriptor{Name: "NORM_USL", Type: 'C', Length: 10},
			)("РИР", "Отопление", "", "", ""),
			buildRawRecord(
				fieldDescriptor{Name: "POSTAV", Type: 'C', Length: 20},
				fieldDescriptor{Name: "VID_USL", Type: 'C', Length: 20},
				fieldDescriptor{Name: "TARIF", Type: 'C', Length: 10},
				fieldDescriptor{Name: "TARIF_DEC", Type: 'C', Length: 10},
				fieldDescriptor{Name: "NORM_USL", Type: 'C', Length: 10},
			)("РИР", "Пеня", "", "", ""),
			buildRawRecord(
				fieldDescriptor{Name: "POSTAV", Type: 'C', Length: 20},
				fieldDescriptor{Name: "VID_USL", Type: 'C', Length: 20},
				fieldDescriptor{Name: "TARIF", Type: 'C', Length: 10},
				fieldDescriptor{Name: "TARIF_DEC", Type: 'C', Length: 10},
				fieldDescriptor{Name: "NORM_USL", Type: 'C', Length: 10},
			)("РИР", "Газ", "12.50", "12.500000", "0.3500"),
		},
	})
	if err != nil {
		t.Fatalf("build source dbf: %v", err)
	}
	if err := os.WriteFile(sourcePath, source, 0o644); err != nil {
		t.Fatalf("write source dbf: %v", err)
	}

	before, err := countBlankRIRRows(sourcePath)
	if err != nil {
		t.Fatalf("count blank rows before filtering: %v", err)
	}
	if before != 2 {
		t.Fatalf("expected two blank rows before filtering, got %d", before)
	}

	err = RemoveRIRZeroRows(sourcePath, outputPath)
	if err != nil {
		t.Fatalf("remove blank rir rows: %v", err)
	}

	after, err := countBlankRIRRows(outputPath)
	if err != nil {
		t.Fatalf("count blank rows after filtering: %v", err)
	}
	if after != 1 {
		t.Fatalf("expected one blank row after filtering for Пеня, got %d", after)
	}

	filteredTable, err := readDBFTable(outputPath)
	if err != nil {
		t.Fatalf("read filtered table: %v", err)
	}
	if len(filteredTable.Records) != 2 {
		t.Fatalf("unexpected filtered record count: %d", len(filteredTable.Records))
	}

	foundPenya := false
	for _, record := range filteredTable.Records {
		if got := readCharacterField(record, filteredTable.Fields, findFieldIndex(filteredTable.Fields, "VID_USL")); got == "Пеня" {
			foundPenya = true
			break
		}
	}
	if !foundPenya {
		t.Fatal("expected Пеня row to be preserved")
	}
}

func TestMergeRIRODNHotWaterRowsMovesSumsToColdWaterComponent(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "rir-odn-source.dbf")
	outputPath := filepath.Join(tempDir, "rir-odn-output.dbf")

	fields := []fieldDescriptor{
		{Name: "LSCHET", Type: 'C', Length: 20},
		{Name: "VID_USL", Type: 'C', Length: 40},
		{Name: "TARIF", Type: 'N', Length: 10, Decimal: 3},
		{Name: "NACHISL", Type: 'N', Length: 12, Decimal: 2},
		{Name: "KOPLATE", Type: 'N', Length: 12, Decimal: 2},
		{Name: "PERERASCH", Type: 'N', Length: 12, Decimal: 2},
	}

	source, err := buildRawDBF(dbfTable{
		Version:        dbfVersionDBaseIII,
		LanguageDriver: dbfLanguageDriverRU,
		Fields:         fields,
		Records: [][]byte{
			buildRawRecord(fields...)("1001", "ГВС: компонент на ТЭ ОДН", "0", "10.00", "7.00", "2.00"),
			buildRawRecord(fields...)("1001", "ГВС: компонент на ХВ ОДН", "32.730", "1.50", "2.50", "3.50"),
			buildRawRecord(fields...)("1002", "ГВС: компонент на ТЭ ОДН", "0", "5.00", "4.00", "3.00"),
			buildRawRecord(fields...)("1002", "ГВС: компонент на ХВ ОДН", "30.000", "6.00", "5.00", "4.00"),
		},
	})
	if err != nil {
		t.Fatalf("build source dbf: %v", err)
	}
	if err := os.WriteFile(sourcePath, source, 0o644); err != nil {
		t.Fatalf("write source dbf: %v", err)
	}

	err = MergeRIRODNHotWaterRows(sourcePath, outputPath, "32,730")
	if err != nil {
		t.Fatalf("merge rir odn rows: %v", err)
	}

	table, err := readDBFTable(outputPath)
	if err != nil {
		t.Fatalf("read merged table: %v", err)
	}

	accountIndex := findFieldIndex(table.Fields, "LSCHET")
	serviceIndex := findFieldIndex(table.Fields, "VID_USL")
	nachislIndex := findFieldIndex(table.Fields, "NACHISL")
	koplateIndex := findFieldIndex(table.Fields, "KOPLATE")
	pereraschIndex := findFieldIndex(table.Fields, "PERERASCH")

	for _, record := range table.Records {
		account := readCharacterField(record, table.Fields, accountIndex)
		service := readCharacterField(record, table.Fields, serviceIndex)

		switch {
		case account == "1001" && service == "ГВС: компонент на ХВ ОДН":
			assertNumericField(t, record, table.Fields, nachislIndex, 11.5)
			assertNumericField(t, record, table.Fields, koplateIndex, 9.5)
			assertNumericField(t, record, table.Fields, pereraschIndex, 5.5)
		case account == "1001" && service == "ГВС: компонент на ТЭ ОДН":
			assertNumericField(t, record, table.Fields, nachislIndex, 0)
			assertNumericField(t, record, table.Fields, koplateIndex, 0)
			assertNumericField(t, record, table.Fields, pereraschIndex, 0)
		case account == "1002" && service == "ГВС: компонент на ХВ ОДН":
			assertNumericField(t, record, table.Fields, nachislIndex, 6)
			assertNumericField(t, record, table.Fields, koplateIndex, 5)
			assertNumericField(t, record, table.Fields, pereraschIndex, 4)
		}
	}
}

func countBlankRIRRows(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	if len(data) < 32 {
		return 0, nil
	}

	recordCount := int(binary.LittleEndian.Uint32(data[4:8]))
	headerLength := int(binary.LittleEndian.Uint16(data[8:10]))
	recordLength := int(binary.LittleEndian.Uint16(data[10:12]))
	fields, err := parseFieldDescriptors(data[:headerLength])
	if err != nil {
		return 0, err
	}

	tarifIndex := findFieldIndex(fields, "TARIF")
	tarifDecIndex := findFieldIndex(fields, "TARIF_DEC")
	normIndex := findFieldIndex(fields, "NORM_USL")
	if tarifIndex < 0 || tarifDecIndex < 0 || normIndex < 0 {
		return 0, nil
	}

	count := 0
	offset := headerLength
	for i := 0; i < recordCount; i++ {
		recordBytes := data[offset : offset+recordLength]
		offset += recordLength
		if len(recordBytes) == 0 || recordBytes[0] == 0x2A {
			continue
		}

		row := recordBytes[1:]
		if isBlankField(row, fields, tarifIndex) &&
			isBlankField(row, fields, tarifDecIndex) &&
			isBlankField(row, fields, normIndex) {
			count++
		}
	}

	return count, nil
}

func buildRawRecord(fields ...fieldDescriptor) func(values ...string) []byte {
	return func(values ...string) []byte {
		record := make([]byte, 0)
		for index, field := range fields {
			value := ""
			if index < len(values) {
				value = values[index]
			}
			encoded := encodeCP866String(value)
			if len(encoded) > field.Length {
				encoded = encoded[:field.Length]
			}

			cell := make([]byte, field.Length)
			for i := range cell {
				cell[i] = ' '
			}
			copy(cell, encoded)
			record = append(record, cell...)
		}
		return record
	}
}

func assertNumericField(t *testing.T, record []byte, fields []fieldDescriptor, targetIndex int, want float64) {
	t.Helper()

	got, err := readNumericField(record, fields, targetIndex)
	if err != nil {
		t.Fatalf("read numeric field %d: %v", targetIndex, err)
	}
	if !almostEqual(got, want) {
		t.Fatalf("unexpected numeric value for field %s: got %v want %v", fields[targetIndex].Name, got, want)
	}
}
