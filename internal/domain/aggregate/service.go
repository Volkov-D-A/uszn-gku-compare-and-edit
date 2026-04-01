package aggregate

import "uszn-gku-compare-and-edit/internal/domain"

func BuildSnapshot(records []domain.ChargeRecord) domain.ProviderSnapshot {
	snapshot := domain.ProviderSnapshot{
		Services:  make(map[string]domain.ServiceAggregate),
		Houses:    make(map[string]domain.HouseAggregate),
		LineItems: make(map[string]domain.LineItemAggregate),
	}

	for i, record := range records {
		if i == 0 {
			snapshot.ProviderName = record.Postav
			snapshot.Month = domain.BuildMonthLabel(record)
		}
		snapshot.RecordCount++

		serviceKey := domain.BuildServiceKey(record)
		houseKey := domain.BuildHouseKey(record)
		lineKey := domain.BuildLineItemKey(record)
		address := domain.BuildHouseAddress(record)

		service := snapshot.Services[serviceKey]
		if service.ServiceKey == "" {
			service = domain.ServiceAggregate{
				ServiceKey:     serviceKey,
				VidUsl:         record.VidUsl,
				NameUsl:        record.NameUsl,
				Tariff:         record.Tarif,
				HouseAddresses: map[string]string{},
			}
		} else if service.Tariff != record.Tarif {
			service.ConflictingTariff = true
		}
		service.TotalAccrual += record.Nachisl
		service.HouseAddresses[houseKey] = address
		snapshot.Services[serviceKey] = service

		house := snapshot.Houses[houseKey]
		if house.HouseKey == "" {
			house = domain.HouseAggregate{
				HouseKey: houseKey,
				Address:  address,
				Services: map[string]domain.ServiceRef{},
			}
		}
		house.TotalAccrual += record.Nachisl
		house.Services[serviceKey] = domain.ServiceRef{
			ServiceKey: serviceKey,
			VidUsl:     record.VidUsl,
			NameUsl:    record.NameUsl,
		}
		snapshot.Houses[houseKey] = house

		lineItem := snapshot.LineItems[lineKey]
		if lineItem.LineKey == "" {
			lineItem = domain.LineItemAggregate{
				LineKey:    lineKey,
				ServiceKey: serviceKey,
				Lshet:      record.Lshet,
				Address:    address,
				VidUsl:     record.VidUsl,
				NameUsl:    record.NameUsl,
			}
		}
		lineItem.TotalAccrual += record.Nachisl
		snapshot.LineItems[lineKey] = lineItem
	}

	return snapshot
}
