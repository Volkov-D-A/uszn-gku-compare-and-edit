package compare

import (
	"math"
	"slices"

	"uszn-gku-compare-and-edit/internal/domain"
)

func Snapshots(prev domain.ProviderSnapshot, curr domain.ProviderSnapshot, settings domain.AnalysisSettings) domain.AnalysisReport {
	report := domain.AnalysisReport{
		TariffChanges:  []domain.TariffChange{},
		ServiceChanges: []domain.ServiceChange{},
		HouseChanges:   []domain.HouseChange{},
		Anomalies:      []domain.AccrualAnomaly{},
		Meta: domain.AnalysisMeta{
			ProviderName:     firstNonEmpty(curr.ProviderName, prev.ProviderName),
			PreviousMonth:    prev.Month,
			CurrentMonth:     curr.Month,
			PreviousRecords:  prev.RecordCount,
			CurrentRecords:   curr.RecordCount,
			PreviousServices: len(prev.Services),
			CurrentServices:  len(curr.Services),
			PreviousHouses:   len(prev.Houses),
			CurrentHouses:    len(curr.Houses),
		},
	}

	for key, prevSvc := range prev.Services {
		currSvc, ok := curr.Services[key]
		if !ok {
			continue
		}
		if prevSvc.Tariff != currSvc.Tariff {
			report.TariffChanges = append(report.TariffChanges, domain.TariffChange{
				ServiceKey:     key,
				VidUsl:         currSvc.VidUsl,
				NameUsl:        currSvc.NameUsl,
				PreviousTariff: prevSvc.Tariff,
				CurrentTariff:  currSvc.Tariff,
			})
		}
	}

	for key, currSvc := range curr.Services {
		if _, ok := prev.Services[key]; ok {
			continue
		}
		report.ServiceChanges = append(report.ServiceChanges, domain.ServiceChange{
			Type:       "appeared",
			ServiceKey: key,
			VidUsl:     currSvc.VidUsl,
			NameUsl:    currSvc.NameUsl,
		})
	}

	for key, prevSvc := range prev.Services {
		if _, ok := curr.Services[key]; ok {
			continue
		}
		report.ServiceChanges = append(report.ServiceChanges, domain.ServiceChange{
			Type:       "disappeared",
			ServiceKey: key,
			VidUsl:     prevSvc.VidUsl,
			NameUsl:    prevSvc.NameUsl,
		})
	}

	for key, currHouse := range curr.Houses {
		if _, ok := prev.Houses[key]; ok {
			continue
		}
		report.HouseChanges = append(report.HouseChanges, domain.HouseChange{
			Type:     "appeared",
			HouseKey: key,
			Address:  currHouse.Address,
		})
	}

	for key, prevHouse := range prev.Houses {
		if _, ok := curr.Houses[key]; ok {
			continue
		}
		report.HouseChanges = append(report.HouseChanges, domain.HouseChange{
			Type:     "disappeared",
			HouseKey: key,
			Address:  prevHouse.Address,
		})
	}

	threshold := math.Abs(settings.AmountChangePercent)
	for key, prevSvc := range prev.Services {
		currSvc, ok := curr.Services[key]
		if !ok {
			continue
		}
		if anomaly, ok := buildAnomaly(prevSvc, currSvc, threshold); ok {
			anomaly.ServiceKey = key
			report.Anomalies = append(report.Anomalies, anomaly)
		}
	}

	sortReport(&report)
	fillSummary(&report)

	return report
}

func buildAnomaly(prev domain.ServiceAggregate, curr domain.ServiceAggregate, threshold float64) (domain.AccrualAnomaly, bool) {
	deltaAmount := curr.TotalAccrual - prev.TotalAccrual
	if prev.TotalAccrual == 0 {
		if curr.TotalAccrual == 0 {
			return domain.AccrualAnomaly{}, false
		}
		return domain.AccrualAnomaly{
			VidUsl:           curr.VidUsl,
			NameUsl:          curr.NameUsl,
			PreviousAmount:   prev.TotalAccrual,
			CurrentAmount:    curr.TotalAccrual,
			DeltaAmount:      deltaAmount,
			DeltaPercent:     nil,
			ThresholdPercent: threshold,
		}, true
	}

	deltaPercent := math.Abs(deltaAmount / prev.TotalAccrual * 100)
	if deltaPercent < threshold {
		return domain.AccrualAnomaly{}, false
	}

	return domain.AccrualAnomaly{
		VidUsl:           curr.VidUsl,
		NameUsl:          curr.NameUsl,
		PreviousAmount:   prev.TotalAccrual,
		CurrentAmount:    curr.TotalAccrual,
		DeltaAmount:      deltaAmount,
		DeltaPercent:     &deltaPercent,
		ThresholdPercent: threshold,
	}, true
}

func fillSummary(report *domain.AnalysisReport) {
	for _, change := range report.ServiceChanges {
		if change.Type == "appeared" {
			report.Summary.AppearedServices++
		}
		if change.Type == "disappeared" {
			report.Summary.DisappearedServices++
		}
	}

	for _, change := range report.HouseChanges {
		if change.Type == "appeared" {
			report.Summary.AppearedHouses++
		}
		if change.Type == "disappeared" {
			report.Summary.DisappearedHouses++
		}
	}

	report.Summary.TariffChanges = len(report.TariffChanges)
	report.Summary.Anomalies = len(report.Anomalies)
}

func sortReport(report *domain.AnalysisReport) {
	slices.SortFunc(report.TariffChanges, func(a, b domain.TariffChange) int {
		return compareStrings(a.ServiceKey, b.ServiceKey)
	})
	slices.SortFunc(report.ServiceChanges, func(a, b domain.ServiceChange) int {
		if a.Type != b.Type {
			return compareStrings(a.Type, b.Type)
		}
		return compareStrings(a.ServiceKey, b.ServiceKey)
	})
	slices.SortFunc(report.HouseChanges, func(a, b domain.HouseChange) int {
		if a.Type != b.Type {
			return compareStrings(a.Type, b.Type)
		}
		return compareStrings(a.Address, b.Address)
	})
	slices.SortFunc(report.Anomalies, func(a, b domain.AccrualAnomaly) int {
		return compareStrings(a.ServiceKey, b.ServiceKey)
	})
}

func compareStrings(a, b string) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
