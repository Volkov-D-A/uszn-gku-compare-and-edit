export type AnalysisSettings = {
  amountChangePercent: number
}

export type AnalysisMeta = {
  providerName: string
  previousMonth: string
  currentMonth: string
  previousRecords: number
  currentRecords: number
  previousServices: number
  currentServices: number
  previousHouses: number
  currentHouses: number
}

export type SummaryCounts = {
  tariffChanges: number
  appearedServices: number
  disappearedServices: number
  appearedHouses: number
  disappearedHouses: number
  anomalies: number
}

export type TariffChange = {
  serviceKey: string
  vidUsl: string
  nameUsl: string
  previousTariff: number
  currentTariff: number
}

export type ServiceChange = {
  type: "appeared" | "disappeared"
  serviceKey: string
  vidUsl: string
  nameUsl: string
}

export type HouseChange = {
  type: "appeared" | "disappeared"
  houseKey: string
  address: string
}

export type AccrualAnomaly = {
  serviceKey: string
  vidUsl: string
  nameUsl: string
  previousAmount: number
  currentAmount: number
  deltaAmount: number
  deltaPercent: number | null
  thresholdPercent: number
}

export type AnalysisReport = {
  meta: AnalysisMeta
  summary: SummaryCounts
  tariffChanges: TariffChange[]
  serviceChanges: ServiceChange[]
  houseChanges: HouseChange[]
  anomalies: AccrualAnomaly[]
}
