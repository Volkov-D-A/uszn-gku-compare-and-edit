package domain

type AnalysisSettings struct {
	AmountChangePercent float64 `json:"amountChangePercent"`
}

type ChargeRecord struct {
	Postav   string
	Lshet    string
	YearS    string
	MonthS   string
	VidUsl   string
	NameUsl  string
	FiasDom  string
	City     string
	SocrCity string
	Street   string
	SocrStr  string
	Dom      string
	Korp     string
	Tarif    float64
	Nachisl  float64
}

type ProviderSnapshot struct {
	ProviderName string
	Month        string
	RecordCount  int
	Services     map[string]ServiceAggregate
	Houses       map[string]HouseAggregate
	LineItems    map[string]LineItemAggregate
}

type ServiceAggregate struct {
	ServiceKey        string
	VidUsl            string
	NameUsl           string
	Tariff            float64
	TotalAccrual      float64
	ConflictingTariff bool
	HouseAddresses    map[string]string
}

type HouseAggregate struct {
	HouseKey     string
	Address      string
	TotalAccrual float64
	Services     map[string]ServiceRef
}

type LineItemAggregate struct {
	LineKey      string
	ServiceKey   string
	Lshet        string
	Address      string
	VidUsl       string
	NameUsl      string
	TotalAccrual float64
}

type ServiceRef struct {
	ServiceKey string `json:"serviceKey"`
	VidUsl     string `json:"vidUsl"`
	NameUsl    string `json:"nameUsl"`
}

type AnalysisMeta struct {
	ProviderName     string `json:"providerName"`
	PreviousMonth    string `json:"previousMonth"`
	CurrentMonth     string `json:"currentMonth"`
	PreviousRecords  int    `json:"previousRecords"`
	CurrentRecords   int    `json:"currentRecords"`
	PreviousServices int    `json:"previousServices"`
	CurrentServices  int    `json:"currentServices"`
	PreviousHouses   int    `json:"previousHouses"`
	CurrentHouses    int    `json:"currentHouses"`
}

type SummaryCounts struct {
	TariffChanges       int `json:"tariffChanges"`
	AppearedServices    int `json:"appearedServices"`
	DisappearedServices int `json:"disappearedServices"`
	AppearedHouses      int `json:"appearedHouses"`
	DisappearedHouses   int `json:"disappearedHouses"`
	Anomalies           int `json:"anomalies"`
}

type TariffChange struct {
	ServiceKey     string  `json:"serviceKey"`
	VidUsl         string  `json:"vidUsl"`
	NameUsl        string  `json:"nameUsl"`
	PreviousTariff float64 `json:"previousTariff"`
	CurrentTariff  float64 `json:"currentTariff"`
}

type ServiceChange struct {
	Type           string   `json:"type"`
	ServiceKey     string   `json:"serviceKey"`
	VidUsl         string   `json:"vidUsl"`
	NameUsl        string   `json:"nameUsl"`
	HouseAddresses []string `json:"houseAddresses"`
}

type HouseChange struct {
	Type     string       `json:"type"`
	HouseKey string       `json:"houseKey"`
	Address  string       `json:"address"`
	Services []ServiceRef `json:"services"`
}

type AccrualAnomaly struct {
	LineKey          string   `json:"lineKey"`
	ServiceKey       string   `json:"serviceKey"`
	Address          string   `json:"address"`
	VidUsl           string   `json:"vidUsl"`
	NameUsl          string   `json:"nameUsl"`
	PreviousAmount   float64  `json:"previousAmount"`
	CurrentAmount    float64  `json:"currentAmount"`
	DeltaPercent     *float64 `json:"deltaPercent"`
	ThresholdPercent float64  `json:"thresholdPercent"`
}

type AnalysisReport struct {
	Meta           AnalysisMeta     `json:"meta"`
	Summary        SummaryCounts    `json:"summary"`
	TariffChanges  []TariffChange   `json:"tariffChanges"`
	ServiceChanges []ServiceChange  `json:"serviceChanges"`
	HouseChanges   []HouseChange    `json:"houseChanges"`
	Anomalies      []AccrualAnomaly `json:"anomalies"`
}

type ExportResult struct {
	Path string `json:"path"`
}
