/// <reference types="vite/client" />

import type { AnalysisReport, AnalysisSettings } from "./types/analysis"

declare global {
  interface Window {
    go?: {
      main?: {
        App?: {
          AnalyzeSupplierFiles: (
            prevPath: string,
            currPath: string,
            settings: AnalysisSettings,
          ) => Promise<AnalysisReport>
          PickDBFFile: () => Promise<string>
          PickCSVFile: () => Promise<string>
          PickExportPath: (defaultName: string) => Promise<string>
          PickDBFExportPath: (defaultName: string) => Promise<string>
          ExportAnalysisXLSX: (
            report: AnalysisReport,
            savePath: string,
          ) => Promise<{ path: string }>
          ConvertNovatekCSVToDBF: (
            csvPath: string,
            savePath: string,
          ) => Promise<{ path: string }>
        }
      }
    }
  }
}

export {}
