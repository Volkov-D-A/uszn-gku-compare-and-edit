package dbf

import (
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
)

const (
	dbfVersionDBaseIII  = 0x03
	dbfLanguageDriverRU = 0x65
	maxDBFFieldLength   = 254
)

type tableField struct {
	Name   string
	Length int
}

type dbfTable struct {
	Version        byte
	LanguageDriver byte
	Fields         []fieldDescriptor
	Records        [][]byte
}

func ConvertNovatekCSVToDBF(csvPath string, savePath string) error {
	data, err := os.ReadFile(csvPath)
	if err != nil {
		return fmt.Errorf("read csv: %w", err)
	}

	rows, err := parseDelimitedRows(data)
	if err != nil {
		return err
	}
	if len(rows) < 2 {
		return fmt.Errorf("csv must contain header and at least one data row")
	}

	header := rows[0]
	records := rows[1:]
	fields := buildFields(header, records)
	content, err := buildDBF(fields, records)
	if err != nil {
		return err
	}

	if err := os.WriteFile(savePath, content, 0o644); err != nil {
		return fmt.Errorf("write dbf: %w", err)
	}

	return nil
}

func RemoveRIRZeroRows(path string, savePath string) error {
	table, err := readDBFTable(path)
	if err != nil {
		return err
	}

	vidUslIndex := findFieldIndex(table.Fields, "VID_USL")
	tarifIndex := findFieldIndex(table.Fields, "TARIF")
	tarifDecIndex := findFieldIndex(table.Fields, "TARIF_DEC")
	normIndex := findFieldIndex(table.Fields, "NORM_USL")
	if vidUslIndex < 0 || tarifIndex < 0 || tarifDecIndex < 0 || normIndex < 0 {
		return fmt.Errorf("required fields VID_USL, TARIF, TARIF_DEC, NORM_USL were not found")
	}

	filtered := make([][]byte, 0, len(table.Records))
	for _, record := range table.Records {
		if shouldKeepRIRRecord(record, table.Fields, vidUslIndex, tarifIndex, tarifDecIndex, normIndex) {
			filtered = append(filtered, record)
		}
	}
	table.Records = filtered

	content, err := buildRawDBF(table)
	if err != nil {
		return err
	}

	if err := os.WriteFile(savePath, content, 0o644); err != nil {
		return fmt.Errorf("write dbf: %w", err)
	}

	return nil
}

func MergeRIRODNHotWaterRows(path string, savePath string, targetTariff string) error {
	table, err := readDBFTable(path)
	if err != nil {
		return err
	}

	accountIndex := findFieldIndex(table.Fields, "LSCHET")
	if accountIndex < 0 {
		accountIndex = findFieldIndex(table.Fields, "LSHET")
	}
	vidUslIndex := findFieldIndex(table.Fields, "VID_USL")
	tarifIndex := findFieldIndex(table.Fields, "TARIF")
	nachislIndex := findFieldIndex(table.Fields, "NACHISL")
	koplateIndex := findFieldIndex(table.Fields, "KOPLATE")
	pereraschIndex := findFieldIndex(table.Fields, "PERERASCH")
	if accountIndex < 0 || vidUslIndex < 0 || tarifIndex < 0 || nachislIndex < 0 || koplateIndex < 0 || pereraschIndex < 0 {
		return fmt.Errorf("required fields LSCHET/LSHET, VID_USL, TARIF, NACHISL, KOPLATE, PERERASCH were not found")
	}

	tariffValue, err := parseNumber(targetTariff)
	if err != nil {
		return fmt.Errorf("parse target tariff: %w", err)
	}

	sourceService := "ГВС: компонент на ТЭ ОДН"
	targetService := "ГВС: компонент на ХВ ОДН"

	targetsByAccount := make(map[string]int)
	sourcesByAccount := make(map[string][]int)

	for index, record := range table.Records {
		account := strings.TrimSpace(readCharacterField(record, table.Fields, accountIndex))
		if account == "" {
			continue
		}

		service := strings.TrimSpace(readCharacterField(record, table.Fields, vidUslIndex))
		switch service {
		case targetService:
			recordTariff, err := readNumericField(record, table.Fields, tarifIndex)
			if err != nil {
				return fmt.Errorf("record %d parse TARIF: %w", index+1, err)
			}
			if almostEqual(recordTariff, tariffValue) {
				if _, exists := targetsByAccount[account]; !exists {
					targetsByAccount[account] = index
				}
			}
		case sourceService:
			sourcesByAccount[account] = append(sourcesByAccount[account], index)
		}
	}

	for account, targetIndex := range targetsByAccount {
		sourceIndexes := sourcesByAccount[account]
		if len(sourceIndexes) == 0 {
			continue
		}

		nachislTotal, err := readNumericField(table.Records[targetIndex], table.Fields, nachislIndex)
		if err != nil {
			return fmt.Errorf("target account %s parse NACHISL: %w", account, err)
		}
		koplateTotal, err := readNumericField(table.Records[targetIndex], table.Fields, koplateIndex)
		if err != nil {
			return fmt.Errorf("target account %s parse KOPLATE: %w", account, err)
		}
		pereraschTotal, err := readNumericField(table.Records[targetIndex], table.Fields, pereraschIndex)
		if err != nil {
			return fmt.Errorf("target account %s parse PERERASCH: %w", account, err)
		}

		for _, sourceIndex := range sourceIndexes {
			sourceRecord := table.Records[sourceIndex]

			nachislValue, err := readNumericField(sourceRecord, table.Fields, nachislIndex)
			if err != nil {
				return fmt.Errorf("source account %s parse NACHISL: %w", account, err)
			}
			koplateValue, err := readNumericField(sourceRecord, table.Fields, koplateIndex)
			if err != nil {
				return fmt.Errorf("source account %s parse KOPLATE: %w", account, err)
			}
			pereraschValue, err := readNumericField(sourceRecord, table.Fields, pereraschIndex)
			if err != nil {
				return fmt.Errorf("source account %s parse PERERASCH: %w", account, err)
			}

			nachislTotal += nachislValue
			koplateTotal += koplateValue
			pereraschTotal += pereraschValue

			if err := writeNumericField(sourceRecord, table.Fields, nachislIndex, 0); err != nil {
				return fmt.Errorf("zero source NACHISL for account %s: %w", account, err)
			}
			if err := writeNumericField(sourceRecord, table.Fields, koplateIndex, 0); err != nil {
				return fmt.Errorf("zero source KOPLATE for account %s: %w", account, err)
			}
			if err := writeNumericField(sourceRecord, table.Fields, pereraschIndex, 0); err != nil {
				return fmt.Errorf("zero source PERERASCH for account %s: %w", account, err)
			}
		}

		if err := writeNumericField(table.Records[targetIndex], table.Fields, nachislIndex, nachislTotal); err != nil {
			return fmt.Errorf("write target NACHISL for account %s: %w", account, err)
		}
		if err := writeNumericField(table.Records[targetIndex], table.Fields, koplateIndex, koplateTotal); err != nil {
			return fmt.Errorf("write target KOPLATE for account %s: %w", account, err)
		}
		if err := writeNumericField(table.Records[targetIndex], table.Fields, pereraschIndex, pereraschTotal); err != nil {
			return fmt.Errorf("write target PERERASCH for account %s: %w", account, err)
		}
	}

	content, err := buildRawDBF(table)
	if err != nil {
		return err
	}

	if err := os.WriteFile(savePath, content, 0o644); err != nil {
		return fmt.Errorf("write dbf: %w", err)
	}

	return nil
}

func parseDelimitedRows(data []byte) ([][]string, error) {
	decoded := decodeCSVContent(data)
	reader := csv.NewReader(strings.NewReader(decoded))
	reader.Comma = ';'
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true

	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parse csv: %w", err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("csv is empty")
	}

	rows[0][0] = strings.TrimPrefix(rows[0][0], "\uFEFF")
	return rows, nil
}

func decodeCSVContent(data []byte) string {
	if utf8.Valid(data) {
		return string(data)
	}

	windows1251 := decodeBytes(charmap.Windows1251, data)
	cp866 := decodeCP866(data)
	if scoreReadableCyrillic(windows1251) >= scoreReadableCyrillic(cp866) {
		return windows1251
	}
	return cp866
}

func decodeBytes(encoding *charmap.Charmap, data []byte) string {
	decoded, err := encoding.NewDecoder().Bytes(data)
	if err != nil {
		return ""
	}
	return string(decoded)
}

func scoreReadableCyrillic(value string) int {
	score := 0
	for _, r := range value {
		switch {
		case r >= 'А' && r <= 'я':
			score += 2
		case r == 'Ё' || r == 'ё':
			score += 2
		case r == ' ' || r == '-' || r == '.' || r == ',' || r == ';':
			score++
		case strings.ContainsRune("░▒▓│┤╡╢╖╕╣║╗╝╜╛┐└┴┬├─┼╞╟╚╔╩╦╠═╬╧╨╤╥╙╘╒╓╫╪┘┌█▄▌▐▀", r):
			score -= 3
		}
	}
	return score
}

func buildFields(header []string, records [][]string) []tableField {
	names := make(map[string]int, len(header))
	fields := make([]tableField, 0, len(header))
	for index, rawName := range header {
		name := makeUniqueFieldName(normalizeFieldName(rawName, index), names)
		maxLen := 1
		for _, record := range records {
			if index >= len(record) {
				continue
			}
			length := len(encodeCP866String(record[index]))
			if length > maxLen {
				maxLen = length
			}
		}
		if maxLen > maxDBFFieldLength {
			maxLen = maxDBFFieldLength
		}
		fields = append(fields, tableField{
			Name:   name,
			Length: maxLen,
		})
	}
	return fields
}

func normalizeFieldName(raw string, index int) string {
	trimmed := strings.ToUpper(strings.TrimSpace(raw))
	if trimmed == "" {
		return fmt.Sprintf("FIELD_%03d", index+1)
	}

	var builder strings.Builder
	for _, r := range trimmed {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			builder.WriteRune(r)
		}
	}

	name := builder.String()
	if name == "" {
		return fmt.Sprintf("FIELD_%03d", index+1)
	}
	if len(name) > 11 {
		name = name[:11]
	}
	return name
}

func makeUniqueFieldName(base string, names map[string]int) string {
	if _, exists := names[base]; !exists {
		names[base] = 1
		return base
	}

	for attempt := 2; ; attempt++ {
		suffix := fmt.Sprintf("_%d", attempt)
		prefixLimit := 11 - len(suffix)
		candidate := base
		if prefixLimit < len(candidate) {
			candidate = candidate[:prefixLimit]
		}
		candidate += suffix
		if _, exists := names[candidate]; !exists {
			names[candidate] = 1
			return candidate
		}
	}
}

func buildDBF(fields []tableField, records [][]string) ([]byte, error) {
	recordLength := 1
	for _, field := range fields {
		recordLength += field.Length
	}

	headerLength := 32 + len(fields)*32 + 1
	totalSize := headerLength + len(records)*recordLength + 1

	var buffer bytes.Buffer
	buffer.Grow(totalSize)

	header := make([]byte, 32)
	header[0] = dbfVersionDBaseIII
	now := time.Now()
	header[1] = byte(now.Year() - 1900)
	header[2] = byte(now.Month())
	header[3] = byte(now.Day())
	binary.LittleEndian.PutUint32(header[4:8], uint32(len(records)))
	binary.LittleEndian.PutUint16(header[8:10], uint16(headerLength))
	binary.LittleEndian.PutUint16(header[10:12], uint16(recordLength))
	header[29] = dbfLanguageDriverRU
	buffer.Write(header)

	for _, field := range fields {
		descriptor := make([]byte, 32)
		copy(descriptor[:11], []byte(field.Name))
		descriptor[11] = 'C'
		descriptor[16] = byte(field.Length)
		buffer.Write(descriptor)
	}
	buffer.WriteByte(0x0D)

	for _, record := range records {
		buffer.WriteByte(0x20)
		for index, field := range fields {
			value := ""
			if index < len(record) {
				value = record[index]
			}
			encoded := encodeCP866String(value)
			if len(encoded) > field.Length {
				encoded = encoded[:field.Length]
			}

			cell := bytes.Repeat([]byte(" "), field.Length)
			copy(cell, encoded)
			buffer.Write(cell)
		}
	}

	buffer.WriteByte(0x1A)
	return buffer.Bytes(), nil
}

func buildRawDBF(table dbfTable) ([]byte, error) {
	recordLength := 1
	for _, field := range table.Fields {
		recordLength += field.Length
	}

	headerLength := 32 + len(table.Fields)*32 + 1
	totalSize := headerLength + len(table.Records)*recordLength + 1

	var buffer bytes.Buffer
	buffer.Grow(totalSize)

	header := make([]byte, 32)
	header[0] = table.Version
	if header[0] == 0 {
		header[0] = dbfVersionDBaseIII
	}
	now := time.Now()
	header[1] = byte(now.Year() - 1900)
	header[2] = byte(now.Month())
	header[3] = byte(now.Day())
	binary.LittleEndian.PutUint32(header[4:8], uint32(len(table.Records)))
	binary.LittleEndian.PutUint16(header[8:10], uint16(headerLength))
	binary.LittleEndian.PutUint16(header[10:12], uint16(recordLength))
	header[29] = table.LanguageDriver
	if header[29] == 0 {
		header[29] = dbfLanguageDriverRU
	}
	buffer.Write(header)

	for _, field := range table.Fields {
		descriptor := make([]byte, 32)
		copy(descriptor[:11], []byte(field.Name))
		descriptor[11] = field.Type
		if descriptor[11] == 0 {
			descriptor[11] = 'C'
		}
		descriptor[16] = byte(field.Length)
		descriptor[17] = byte(field.Decimal)
		buffer.Write(descriptor)
	}
	buffer.WriteByte(0x0D)

	for _, record := range table.Records {
		if len(record) != recordLength-1 {
			return nil, fmt.Errorf("invalid record length: expected %d, got %d", recordLength-1, len(record))
		}
		buffer.WriteByte(0x20)
		buffer.Write(record)
	}

	buffer.WriteByte(0x1A)
	return buffer.Bytes(), nil
}

func readDBFTable(path string) (dbfTable, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return dbfTable{}, fmt.Errorf("read file: %w", err)
	}
	if len(data) < 32 {
		return dbfTable{}, fmt.Errorf("file is too short")
	}

	recordCount := int(binary.LittleEndian.Uint32(data[4:8]))
	headerLength := int(binary.LittleEndian.Uint16(data[8:10]))
	recordLength := int(binary.LittleEndian.Uint16(data[10:12]))
	if headerLength <= 32 || recordLength <= 1 || len(data) < headerLength+recordCount*recordLength {
		return dbfTable{}, fmt.Errorf("invalid DBF header")
	}

	fields, err := parseFieldDescriptors(data[:headerLength])
	if err != nil {
		return dbfTable{}, err
	}

	records := make([][]byte, 0, recordCount)
	offset := headerLength
	for i := 0; i < recordCount; i++ {
		recordBytes := data[offset : offset+recordLength]
		offset += recordLength
		if len(recordBytes) == 0 || recordBytes[0] == 0x2A {
			continue
		}

		row := make([]byte, recordLength-1)
		copy(row, recordBytes[1:])
		records = append(records, row)
	}

	return dbfTable{
		Version:        data[0],
		LanguageDriver: data[29],
		Fields:         fields,
		Records:        records,
	}, nil
}

func findFieldIndex(fields []fieldDescriptor, name string) int {
	for index, field := range fields {
		if field.Name == name {
			return index
		}
	}
	return -1
}

func shouldKeepRIRRecord(
	record []byte,
	fields []fieldDescriptor,
	vidUslIndex int,
	tarifIndex int,
	tarifDecIndex int,
	normIndex int,
) bool {
	if strings.EqualFold(readCharacterField(record, fields, vidUslIndex), "Пеня") {
		return true
	}

	return !isBlankField(record, fields, tarifIndex) ||
		!isBlankField(record, fields, tarifDecIndex) ||
		!isBlankField(record, fields, normIndex)
}

func readCharacterField(record []byte, fields []fieldDescriptor, targetIndex int) string {
	offset := 0
	for index, field := range fields {
		nextOffset := offset + field.Length
		if nextOffset > len(record) {
			return ""
		}
		if index == targetIndex {
			raw := record[offset:nextOffset]
			if field.Type == 'C' {
				return strings.TrimSpace(decodeCP866(raw))
			}
			return strings.TrimSpace(string(raw))
		}
		offset = nextOffset
	}
	return ""
}

func isBlankField(record []byte, fields []fieldDescriptor, targetIndex int) bool {
	offset := 0
	for index, field := range fields {
		nextOffset := offset + field.Length
		if nextOffset > len(record) {
			return true
		}
		if index == targetIndex {
			return len(bytes.TrimSpace(record[offset:nextOffset])) == 0
		}
		offset = nextOffset
	}
	return true
}

func readNumericField(record []byte, fields []fieldDescriptor, targetIndex int) (float64, error) {
	return parseNumber(readRawField(record, fields, targetIndex))
}

func readRawField(record []byte, fields []fieldDescriptor, targetIndex int) string {
	offset := 0
	for index, field := range fields {
		nextOffset := offset + field.Length
		if nextOffset > len(record) {
			return ""
		}
		if index == targetIndex {
			return strings.TrimSpace(string(record[offset:nextOffset]))
		}
		offset = nextOffset
	}
	return ""
}

func writeNumericField(record []byte, fields []fieldDescriptor, targetIndex int, value float64) error {
	offset, field, ok := fieldOffset(fields, targetIndex)
	if !ok {
		return fmt.Errorf("field index %d not found", targetIndex)
	}
	if offset+field.Length > len(record) {
		return fmt.Errorf("field %s is out of bounds", field.Name)
	}

	formatted := formatNumericValue(value, field.Length, field.Decimal)
	if len(formatted) > field.Length {
		return fmt.Errorf("value %q does not fit in field %s", formatted, field.Name)
	}

	cell := bytes.Repeat([]byte(" "), field.Length)
	copy(cell[field.Length-len(formatted):], []byte(formatted))
	copy(record[offset:offset+field.Length], cell)
	return nil
}

func fieldOffset(fields []fieldDescriptor, targetIndex int) (int, fieldDescriptor, bool) {
	offset := 0
	for index, field := range fields {
		if index == targetIndex {
			return offset, field, true
		}
		offset += field.Length
	}
	return 0, fieldDescriptor{}, false
}

func formatNumericValue(value float64, width int, decimals int) string {
	rounded := math.Round(value*1_000_000) / 1_000_000
	if decimals > 0 {
		return strconv.FormatFloat(rounded, 'f', decimals, 64)
	}
	if math.Abs(rounded-math.Round(rounded)) < 0.0000001 {
		return strconv.FormatInt(int64(math.Round(rounded)), 10)
	}
	return strconv.FormatFloat(rounded, 'f', -1, 64)
}

func almostEqual(a float64, b float64) bool {
	return math.Abs(a-b) < 0.000001
}

func encodeCP866String(value string) []byte {
	out := make([]byte, 0, len(value))
	for _, r := range value {
		switch {
		case r < 0x80:
			out = append(out, byte(r))
		default:
			encoded, ok := encodeCP866Rune(r)
			if !ok {
				out = append(out, '?')
				continue
			}
			out = append(out, encoded)
		}
	}
	return out
}

func encodeCP866Rune(target rune) (byte, bool) {
	for index, candidate := range cp866Table {
		if candidate == target {
			return byte(index + 0x80), true
		}
	}
	return 0, false
}
