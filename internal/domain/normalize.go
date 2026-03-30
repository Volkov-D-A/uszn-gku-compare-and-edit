package domain

import (
	"fmt"
	"strings"
)

func NormalizeText(value string) string {
	return strings.Join(strings.Fields(strings.ToUpper(strings.TrimSpace(value))), " ")
}

func BuildServiceKey(record ChargeRecord) string {
	return NormalizeText(record.VidUsl) + "::" + NormalizeText(record.NameUsl)
}

func BuildHouseKey(record ChargeRecord) string {
	if key := NormalizeText(record.FiasDom); key != "" {
		return key
	}

	parts := []string{
		NormalizeText(record.City),
		NormalizeText(record.SocrCity),
		NormalizeText(record.Street),
		NormalizeText(record.SocrStr),
		NormalizeText(record.Dom),
		NormalizeText(record.Korp),
	}

	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			filtered = append(filtered, part)
		}
	}

	return strings.Join(filtered, "|")
}

func BuildHouseAddress(record ChargeRecord) string {
	parts := []string{
		strings.TrimSpace(record.SocrCity),
		strings.TrimSpace(record.City),
		strings.TrimSpace(record.SocrStr),
		strings.TrimSpace(record.Street),
		"д. " + strings.TrimSpace(record.Dom),
	}

	if strings.TrimSpace(record.Korp) != "" {
		parts = append(parts, "корп. "+strings.TrimSpace(record.Korp))
	}

	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" && part != "д." {
			filtered = append(filtered, part)
		}
	}

	return strings.Join(filtered, ", ")
}

func BuildMonthLabel(record ChargeRecord) string {
	year := strings.TrimSpace(record.YearS)
	month := strings.TrimSpace(record.MonthS)
	if year == "" || month == "" {
		return ""
	}
	return fmt.Sprintf("%s-%s", year, month)
}
