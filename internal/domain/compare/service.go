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
			Type:           "appeared",
			ServiceKey:     key,
			VidUsl:         currSvc.VidUsl,
			NameUsl:        currSvc.NameUsl,
			HouseAddresses: sortedHouseAddresses(currSvc.HouseAddresses),
		})
	}

	for key, prevSvc := range prev.Services {
		if _, ok := curr.Services[key]; ok {
			continue
		}
		report.ServiceChanges = append(report.ServiceChanges, domain.ServiceChange{
			Type:           "disappeared",
			ServiceKey:     key,
			VidUsl:         prevSvc.VidUsl,
			NameUsl:        prevSvc.NameUsl,
			HouseAddresses: sortedHouseAddresses(prevSvc.HouseAddresses),
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
			Services: sortedServices(currHouse.Services),
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
			Services: sortedServices(prevHouse.Services),
		})
	}

	threshold := math.Abs(settings.AmountChangePercent)
	lineKeys := make(map[string]struct{}, len(prev.LineItems)+len(curr.LineItems))
	for key := range prev.LineItems {
		lineKeys[key] = struct{}{}
	}
	for key := range curr.LineItems {
		lineKeys[key] = struct{}{}
	}
	for key := range lineKeys {
		prevLine, prevOK := prev.LineItems[key]
		currLine, currOK := curr.LineItems[key]
		if !prevOK || !currOK {
			continue
		}
		if anomaly, ok := buildLineAnomaly(key, prevLine, prevOK, currLine, currOK, threshold); ok {
			report.Anomalies = append(report.Anomalies, anomaly)
		}
	}

	sortReport(&report)
	fillSummary(&report)

	return report
}

func buildLineAnomaly(lineKey string, prev domain.LineItemAggregate, prevOK bool, curr domain.LineItemAggregate, currOK bool, threshold float64) (domain.AccrualAnomaly, bool) {
	if !prevOK || !currOK {
		return domain.AccrualAnomaly{}, false
	}

	prevAmount := 0.0
	currAmount := 0.0
	base := curr
	if prevOK {
		prevAmount = prev.TotalAccrual
	}
	if currOK {
		currAmount = curr.TotalAccrual
	}

	if prevAmount == 0 || currAmount == 0 {
		return domain.AccrualAnomaly{}, false
	}

	deltaPercent := math.Abs((currAmount - prevAmount) / prevAmount * 100)
	if deltaPercent < threshold {
		return domain.AccrualAnomaly{}, false
	}

	return domain.AccrualAnomaly{
		LineKey:          lineKey,
		ServiceKey:       base.ServiceKey,
		Address:          base.Address,
		VidUsl:           base.VidUsl,
		NameUsl:          base.NameUsl,
		PreviousAmount:   prevAmount,
		CurrentAmount:    currAmount,
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
		if a.Address != b.Address {
			return compareStrings(a.Address, b.Address)
		}
		return compareStrings(a.ServiceKey, b.ServiceKey)
	})
}

func sortedHouseAddresses(addresses map[string]string) []string {
	if len(addresses) == 0 {
		return []string{}
	}
	values := make([]string, 0, len(addresses))
	for _, address := range addresses {
		values = append(values, address)
	}
	slices.Sort(values)
	return values
}

func sortedServices(services map[string]domain.ServiceRef) []domain.ServiceRef {
	if len(services) == 0 {
		return []domain.ServiceRef{}
	}
	values := make([]domain.ServiceRef, 0, len(services))
	for _, service := range services {
		values = append(values, service)
	}
	slices.SortFunc(values, func(a, b domain.ServiceRef) int {
		return compareStrings(a.ServiceKey, b.ServiceKey)
	})
	return values
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
