import type { AnalysisReport } from "../types/analysis"

export const sampleReport: AnalysisReport = {
  meta: {
    providerName: "ММПКХ",
    previousMonth: "2026-01",
    currentMonth: "2026-02",
    previousRecords: 4,
    currentRecords: 6,
    previousServices: 4,
    currentServices: 6,
    previousHouses: 1,
    currentHouses: 1,
  },
  summary: {
    tariffChanges: 0,
    appearedServices: 2,
    disappearedServices: 0,
    appearedHouses: 0,
    disappearedHouses: 0,
    anomalies: 1,
  },
  tariffChanges: [],
  serviceChanges: [
    {
      type: "appeared",
      serviceKey: "ВОДООТВЕДЕНИЕ (КАНАЛИЗАЦИЯ)(К)::ВО ГВС. ПОЛНОЕ БЛАГОУСТРОЙСТВО ВАННА 1650-1700 ННЗ (М3)",
      vidUsl: "Водоотведение (канализация)(К)",
      nameUsl: "ВО ГВС. ПОЛНОЕ БЛАГОУСТРОЙСТВО ВАННА 1650-1700 ННЗ (м3)",
    },
    {
      type: "appeared",
      serviceKey: "ВОДООТВЕДЕНИЕ (КАНАЛИЗАЦИЯ)(К)::ВО ХВС. ПОЛНОЕ БЛАГОУСТРОЙСТВО ВАННА 1650-1700 ННЗ (М3)",
      vidUsl: "Водоотведение (канализация)(К)",
      nameUsl: "ВО ХВС. ПОЛНОЕ БЛАГОУСТРОЙСТВО ВАННА 1650-1700 ННЗ (м3)",
    },
  ],
  houseChanges: [],
  anomalies: [
    {
      serviceKey: "ВОДООТВЕДЕНИЕ (КАНАЛИЗАЦИЯ)(К)::ВО. ПОЛНОЕ БЛАГОУСТРОЙСТВО ВАННА 1650-1700 ННЗ (М3)",
      vidUsl: "Водоотведение (канализация)(К)",
      nameUsl: "ВО. ПОЛНОЕ БЛАГОУСТРОЙСТВО ВАННА 1650-1700 ННЗ (м3)",
      previousAmount: 987.05,
      currentAmount: 0,
      deltaAmount: -987.05,
      deltaPercent: 100,
      thresholdPercent: 20,
    },
  ],
}
