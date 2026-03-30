import { sampleReport } from "../dev/sampleReport"
import type { AnalysisReport, AnalysisSettings } from "../types/analysis"

type ExportResult = {
  path: string
}

function getBackend() {
  return window.go?.main?.App
}

export async function analyzeSupplierFiles(
  prevPath: string,
  currPath: string,
  settings: AnalysisSettings,
): Promise<AnalysisReport> {
  const backend = getBackend()
  if (backend) {
    return backend.AnalyzeSupplierFiles(prevPath, currPath, settings)
  }

  if (import.meta.env.DEV) {
    await wait(150)
    return sampleReport
  }

  throw new Error("Wails bridge is not connected.")
}

export async function exportAnalysisXLSX(
  report: AnalysisReport,
  savePath: string,
): Promise<ExportResult> {
  const backend = getBackend()
  if (backend) {
    return backend.ExportAnalysisXLSX(report, savePath)
  }

  if (import.meta.env.DEV) {
    await wait(100)
    return { path: savePath || "report.xlsx" }
  }

  throw new Error("Wails bridge is not connected.")
}

export async function pickDBFFile(): Promise<string> {
  const backend = getBackend()
  if (backend) {
    return backend.PickDBFFile()
  }

  if (import.meta.env.DEV) {
    await wait(50)
    return ""
  }

  throw new Error("Wails bridge is not connected.")
}

export async function pickExportPath(defaultName: string): Promise<string> {
  const backend = getBackend()
  if (backend) {
    return backend.PickExportPath(defaultName)
  }

  if (import.meta.env.DEV) {
    await wait(50)
    return defaultName
  }

  throw new Error("Wails bridge is not connected.")
}

function wait(ms: number) {
  return new Promise<void>((resolve) => {
    window.setTimeout(resolve, ms)
  })
}
