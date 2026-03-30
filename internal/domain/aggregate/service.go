package aggregate

import "uszn-gku-compare-and-edit/internal/domain"

func BuildSnapshot(records []domain.ChargeRecord) domain.ProviderSnapshot {
	snapshot := domain.ProviderSnapshot{
		Services: make(map[string]domain.ServiceAggregate),
		Houses:   make(map[string]domain.HouseAggregate),
	}

	for i, record := range records {
		if i == 0 {
			snapshot.ProviderName = record.Postav
			snapshot.Month = domain.BuildMonthLabel(record)
		}
		snapshot.RecordCount++

		serviceKey := domain.BuildServiceKey(record)
		houseKey := domain.BuildHouseKey(record)

		service := snapshot.Services[serviceKey]
		if service.ServiceKey == "" {
			service = domain.ServiceAggregate{
				ServiceKey: serviceKey,
				VidUsl:     record.VidUsl,
				NameUsl:    record.NameUsl,
				Tariff:     record.Tarif,
			}
		} else if service.Tariff != record.Tarif {
			service.ConflictingTariff = true
		}
		service.TotalAccrual += record.Nachisl
		snapshot.Services[serviceKey] = service

		house := snapshot.Houses[houseKey]
		if house.HouseKey == "" {
			house = domain.HouseAggregate{
				HouseKey: houseKey,
				Address:  domain.BuildHouseAddress(record),
			}
		}
		house.TotalAccrual += record.Nachisl
		snapshot.Houses[houseKey] = house
	}

	return snapshot
}
