package dbf

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"uszn-gku-compare-and-edit/internal/domain"
)

type fieldDescriptor struct {
	Name    string
	Type    byte
	Length  int
	Decimal int
}

func ReadCharges(path string) ([]domain.ChargeRecord, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	if len(data) < 32 {
		return nil, fmt.Errorf("file is too short")
	}

	recordCount := int(binary.LittleEndian.Uint32(data[4:8]))
	headerLength := int(binary.LittleEndian.Uint16(data[8:10]))
	recordLength := int(binary.LittleEndian.Uint16(data[10:12]))
	if headerLength <= 32 || recordLength <= 1 || len(data) < headerLength+recordCount*recordLength {
		return nil, fmt.Errorf("invalid DBF header")
	}

	fields, err := parseFieldDescriptors(data[:headerLength])
	if err != nil {
		return nil, err
	}

	records := make([]domain.ChargeRecord, 0, recordCount)
	offset := headerLength
	for i := 0; i < recordCount; i++ {
		recordBytes := data[offset : offset+recordLength]
		offset += recordLength
		if len(recordBytes) == 0 || recordBytes[0] == 0x2A {
			continue
		}

		row := parseRow(recordBytes[1:], fields)
		record, err := mapChargeRecord(row)
		if err != nil {
			return nil, fmt.Errorf("record %d: %w", i+1, err)
		}
		records = append(records, record)
	}

	return records, nil
}

func parseFieldDescriptors(header []byte) ([]fieldDescriptor, error) {
	fields := make([]fieldDescriptor, 0)
	for offset := 32; offset < len(header); offset += 32 {
		if header[offset] == 0x0D {
			return fields, nil
		}
		if offset+32 > len(header) {
			break
		}

		name := strings.TrimRight(string(header[offset:offset+11]), "\x00 ")
		fields = append(fields, fieldDescriptor{
			Name:    name,
			Type:    header[offset+11],
			Length:  int(header[offset+16]),
			Decimal: int(header[offset+17]),
		})
	}

	return nil, fmt.Errorf("field descriptor terminator not found")
}

func parseRow(raw []byte, fields []fieldDescriptor) map[string]string {
	row := make(map[string]string, len(fields))
	offset := 0
	for _, field := range fields {
		if offset+field.Length > len(raw) {
			break
		}
		chunk := raw[offset : offset+field.Length]
		offset += field.Length

		switch field.Type {
		case 'C':
			row[field.Name] = strings.TrimSpace(decodeCP866(chunk))
		default:
			row[field.Name] = strings.TrimSpace(string(chunk))
		}
	}

	return row
}

func mapChargeRecord(row map[string]string) (domain.ChargeRecord, error) {
	tarif, err := parseNumber(row["TARIF"])
	if err != nil {
		return domain.ChargeRecord{}, fmt.Errorf("parse TARIF: %w", err)
	}
	nachisl, err := parseNumber(row["NACHISL"])
	if err != nil {
		return domain.ChargeRecord{}, fmt.Errorf("parse NACHISL: %w", err)
	}

	return domain.ChargeRecord{
		Postav:   row["POSTAV"],
		YearS:    row["YEAR_S"],
		MonthS:   row["MONTH_S"],
		VidUsl:   row["VID_USL"],
		NameUsl:  row["NAME_USL"],
		FiasDom:  row["FIAS_DOM"],
		City:     row["CITY"],
		SocrCity: row["SOCR_CITY"],
		Street:   row["STREET"],
		SocrStr:  row["SOCR_STR"],
		Dom:      row["DOM"],
		Korp:     row["KORP"],
		Tarif:    tarif,
		Nachisl:  nachisl,
	}, nil
}

func parseNumber(raw string) (float64, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return 0, nil
	}
	parsed, err := strconv.ParseFloat(strings.ReplaceAll(value, ",", "."), 64)
	if err != nil {
		return 0, err
	}
	return math.Round(parsed*1_000_000) / 1_000_000, nil
}

func decodeCP866(raw []byte) string {
	runes := make([]rune, 0, len(raw))
	for _, b := range raw {
		if b < 0x80 {
			runes = append(runes, rune(b))
			continue
		}
		runes = append(runes, cp866Table[b-0x80])
	}
	return string(runes)
}

var cp866Table = [...]rune{
	'А', 'Б', 'В', 'Г', 'Д', 'Е', 'Ж', 'З',
	'И', 'Й', 'К', 'Л', 'М', 'Н', 'О', 'П',
	'Р', 'С', 'Т', 'У', 'Ф', 'Х', 'Ц', 'Ч',
	'Ш', 'Щ', 'Ъ', 'Ы', 'Ь', 'Э', 'Ю', 'Я',
	'а', 'б', 'в', 'г', 'д', 'е', 'ж', 'з',
	'и', 'й', 'к', 'л', 'м', 'н', 'о', 'п',
	'░', '▒', '▓', '│', '┤', '╡', '╢', '╖',
	'╕', '╣', '║', '╗', '╝', '╜', '╛', '┐',
	'└', '┴', '┬', '├', '─', '┼', '╞', '╟',
	'╚', '╔', '╩', '╦', '╠', '═', '╬', '╧',
	'╨', '╤', '╥', '╙', '╘', '╒', '╓', '╫',
	'╪', '┘', '┌', '█', '▄', '▌', '▐', '▀',
	'р', 'с', 'т', 'у', 'ф', 'х', 'ц', 'ч',
	'ш', 'щ', 'ъ', 'ы', 'ь', 'э', 'ю', 'я',
	'Ё', 'ё', 'Є', 'є', 'Ї', 'ї', 'Ў', 'ў',
	'°', '∙', '·', '√', '№', '¤', '■', '\u00a0',
}
