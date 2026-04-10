package dbf

import (
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"os"
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
