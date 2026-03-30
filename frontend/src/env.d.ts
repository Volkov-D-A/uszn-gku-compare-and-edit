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
          PickExportPath: (defaultName: string) => Promise<string>
          ExportAnalysisXLSX: (
            report: AnalysisReport,
            savePath: string,
          ) => Promise<{ path: string }>
        }
      }
    }
  }
}

export {}
