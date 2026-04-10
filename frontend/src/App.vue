<script setup lang="ts">
import { ref } from "vue"
import AnalysisSettings from "./components/AnalysisSettings.vue"
import FilePicker from "./components/FilePicker.vue"
import PreparationUtilities from "./components/PreparationUtilities.vue"
import SummaryCards from "./components/SummaryCards.vue"
import {
  analyzeSupplierFiles,
  exportAnalysisXLSX,
  pickDBFFile,
  pickExportPath,
} from "./lib/backend"
import type { AnalysisReport } from "./types/analysis"

const previousPath = ref("")
const currentPath = ref("")
const threshold = ref(20)
const isUtilitiesOpen = ref(false)

const report = ref<AnalysisReport | null>(null)
const errorMessage = ref("")
const successMessage = ref("")
const isAnalyzing = ref(false)
const isExporting = ref(false)

async function runAnalysis() {
  errorMessage.value = ""
  successMessage.value = ""

  if (!previousPath.value.trim() || !currentPath.value.trim()) {
    errorMessage.value = "Укажите пути к файлам предыдущего и текущего месяца."
    return
  }

  isAnalyzing.value = true
  try {
    report.value = await analyzeSupplierFiles(previousPath.value, currentPath.value, {
      amountChangePercent: threshold.value,
    })
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : "Не удалось выполнить анализ."
  } finally {
    isAnalyzing.value = false
  }
}

async function browsePrevious() {
  const path = await pickDBFFileSafe()
  if (path) {
    previousPath.value = path
  }
}

async function browseCurrent() {
  const path = await pickDBFFileSafe()
  if (path) {
    currentPath.value = path
  }
}

async function browseExport() {
  errorMessage.value = ""
  try {
    const fallbackName = currentPath.value.replace(/\.dbf$/i, "_analysis.xlsx") || "report.xlsx"
    const path = await pickExportPath(fallbackName)
    return path
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : "Не удалось выбрать путь сохранения."
    return ""
  }
}

async function runExport() {
  errorMessage.value = ""
  successMessage.value = ""

  if (!report.value) {
    errorMessage.value = "Сначала выполните анализ."
    return
  }

  const savePath = await browseExport()
  if (!savePath) {
    return
  }

  isExporting.value = true
  try {
    const result = await exportAnalysisXLSX(report.value, savePath)
    successMessage.value = `Отчёт сохранён: ${result.path}`
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : "Не удалось сохранить отчёт."
  } finally {
    isExporting.value = false
  }
}

async function pickDBFFileSafe() {
  errorMessage.value = ""
  try {
    return await pickDBFFile()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : "Не удалось выбрать DBF файл."
    return ""
  }
}
</script>

<template>
  <main class="page">
    <section class="workspace">
      <header class="topbar">
        <div>
          <p class="eyebrow">Сравнение поставщика</p>
          <h1>Анализ начислений ЖКУ</h1>
        </div>
        <button class="utility-button" type="button" @click="isUtilitiesOpen = true">Утилиты</button>
      </header>

      <section class="control-grid">
        <article class="panel panel--wide">
          <div class="panel-head">
            <div>
              <h2>Источник данных</h2>
              <p>Пути к файлам за соседние месяцы.</p>
            </div>
            <span class="panel-chip">DBF</span>
          </div>
          <div class="panel-body panel-body--stack">
            <FilePicker
              v-model="previousPath"
              label="Предыдущий месяц"
              :disabled="isAnalyzing"
              @browse="browsePrevious"
            />
            <FilePicker
              v-model="currentPath"
              label="Текущий месяц"
              :disabled="isAnalyzing"
              @browse="browseCurrent"
            />
          </div>
        </article>

        <article class="panel">
          <div class="panel-head">
            <div>
              <h2>Параметры анализа</h2>
              <p>Порог для выявления отклонений.</p>
            </div>
            <span class="panel-chip">{{ threshold }}%</span>
          </div>
          <div class="panel-body panel-body--stack">
            <AnalysisSettings v-model="threshold" />
            <div class="threshold-preview">
              <span>Текущий порог аномалий</span>
              <strong>{{ threshold }}%</strong>
            </div>
            <button class="primary-button" type="button" :disabled="isAnalyzing" @click="runAnalysis">
              {{ isAnalyzing ? "Идёт анализ..." : "Запустить анализ" }}
            </button>
          </div>
        </article>
      </section>

      <section class="messages" v-if="errorMessage || successMessage">
        <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
        <p v-if="successMessage" class="success">{{ successMessage }}</p>
      </section>

      <section v-if="report" class="results">
        <section class="meta-row">
          <article class="meta-card">
            <span>Поставщик</span>
            <strong>{{ report.meta.providerName }}</strong>
          </article>
          <article class="meta-card">
            <span>Период анализа</span>
            <strong>{{ report.meta.previousMonth }} → {{ report.meta.currentMonth }}</strong>
          </article>
          <article class="meta-card">
            <span>Объём записей</span>
            <strong>{{ report.meta.previousRecords }} → {{ report.meta.currentRecords }}</strong>
          </article>
        </section>

        <SummaryCards :summary="report.summary" />

        <section class="panel export-panel">
          <div class="panel-head">
            <div>
              <h2>Экспорт отчёта</h2>
              <p>Подробности по адресам, услугам и аномалиям сохраняются только в XLSX.</p>
            </div>
            <span class="panel-chip">XLSX</span>
          </div>
          <div class="export-layout export-layout--single">
            <button class="primary-button" type="button" :disabled="isExporting" @click="runExport">
              {{ isExporting ? "Сохраняем..." : "Сохранить отчёт в XLSX" }}
            </button>
          </div>
        </section>
      </section>
    </section>

    <PreparationUtilities :open="isUtilitiesOpen" @close="isUtilitiesOpen = false" />
  </main>
</template>

<style scoped>
.page {
  background: #f6f6f5;
  color: #121212;
  display: block;
  font-family: "Avenir Next", "Segoe UI", sans-serif;
  font-size: 14px;
  min-height: 100vh;
  padding: 1rem;
}

.workspace,
.results {
  display: grid;
  gap: 0.75rem;
  margin: 0 auto;
  max-width: 1200px;
  width: 100%;
}

.topbar {
  align-items: start;
  background: #ffffff;
  border: 1px solid rgba(0, 0, 0, 0.06);
  border-radius: 28px;
  display: flex;
  gap: 1rem;
  justify-content: space-between;
  padding: 1.1rem 1.2rem;
}

.utility-button {
  align-self: center;
  background: #f1efe9;
  border: 1px solid rgba(0, 0, 0, 0.08);
  border-radius: 999px;
  color: #1f1f1f;
  cursor: pointer;
  min-height: 2.8rem;
  padding: 0.7rem 1.1rem;
  white-space: nowrap;
}

.eyebrow {
  font-size: 0.78rem;
  letter-spacing: 0.12em;
  margin: 0 0 0.3rem;
  text-transform: uppercase;
}

.topbar h1 {
  font-size: clamp(1.6rem, 3vw, 2rem);
  margin: 0;
}

.control-grid,
.meta-row {
  display: grid;
  gap: 0.75rem;
}

.panel,
.meta-card {
  background: #ffffff;
  border: 1px solid rgba(0, 0, 0, 0.06);
  border-radius: 20px;
  box-shadow: 0 4px 16px rgba(28, 23, 17, 0.03);
}

.control-grid {
  grid-template-columns: minmax(0, 1.45fr) minmax(260px, 0.55fr);
}

.panel {
  padding: 0.95rem 1rem;
}

.panel--wide {
  min-width: 0;
}

.panel-head {
  align-items: start;
  display: flex;
  gap: 1rem;
  justify-content: space-between;
  margin-bottom: 0.7rem;
}

.panel-head h2 {
  font-size: 0.95rem;
  margin: 0;
}

.panel-head p {
  font-size: 0.88rem;
  margin: 0.25rem 0 0;
}

.panel-chip {
  background: #f3f3f3;
  border-radius: 999px;
  color: #5d5d5d;
  font-size: 0.78rem;
  padding: 0.35rem 0.7rem;
}

.panel-body {
  display: grid;
  gap: 0.75rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.panel-body--stack {
  grid-template-columns: 1fr;
}

.threshold-preview {
  align-items: end;
  background: #f7f7f7;
  border-radius: 16px;
  display: flex;
  justify-content: space-between;
  padding: 0.75rem 0.85rem;
}

.threshold-preview span {
  color: #7a7a7a;
}

.threshold-preview strong {
  font-size: 1.45rem;
}

.primary-button {
  background: #1f1f1f;
  border: 0;
  border-radius: 999px;
  color: #ffffff;
  cursor: pointer;
  min-height: 2.9rem;
  padding: 0.75rem 1.15rem;
}

.primary-button:disabled {
  cursor: progress;
  opacity: 0.7;
}

.messages {
  display: grid;
  gap: 0.5rem;
}

.error {
  background: #fff1f2;
  border: 1px solid #ffd5da;
  border-radius: 18px;
  color: #a52e43;
  margin: 0;
  padding: 0.9rem 1rem;
}

.success {
  background: #effcf3;
  border: 1px solid #cdeed5;
  border-radius: 18px;
  color: #2e7750;
  margin: 0;
  padding: 0.9rem 1rem;
}

.hint {
  margin: 0;
}

.meta-row {
  grid-template-columns: repeat(3, minmax(220px, 1fr));
}

.meta-card {
  display: grid;
  gap: 0.3rem;
  min-width: 0;
  padding: 0.8rem 0.9rem;
}

.meta-card span {
  color: #7a7a7a;
  font-size: 0.78rem;
  text-transform: uppercase;
}

.meta-card strong {
  font-size: 1.1rem;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.export-layout {
  align-items: end;
  display: grid;
  gap: 0.75rem;
  grid-template-columns: minmax(0, 1fr) auto;
}

.export-layout--single {
  grid-template-columns: 1fr;
}

.details-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

@media (max-width: 1200px) {
  .control-grid,
  .meta-row {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 760px) {
  .topbar {
    flex-direction: column;
  }

  .panel-body {
    grid-template-columns: 1fr;
  }

  .export-layout {
    grid-template-columns: 1fr;
  }
}
</style>
