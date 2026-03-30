<script setup lang="ts">
import { computed, ref } from "vue"
import AnalysisSettings from "./components/AnalysisSettings.vue"
import DetailsSection from "./components/DetailsSection.vue"
import FilePicker from "./components/FilePicker.vue"
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
const exportPath = ref("")
const threshold = ref(20)

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
    if (!exportPath.value.trim()) {
      exportPath.value = currentPath.value.replace(/\.dbf$/i, "_analysis.xlsx")
    }
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
    const fallbackName =
      exportPath.value.trim() ||
      currentPath.value.replace(/\.dbf$/i, "_analysis.xlsx") ||
      "report.xlsx"
    const path = await pickExportPath(fallbackName)
    if (path) {
      exportPath.value = path
    }
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : "Не удалось выбрать путь сохранения."
  }
}

async function runExport() {
  errorMessage.value = ""
  successMessage.value = ""

  if (!report.value) {
    errorMessage.value = "Сначала выполните анализ."
    return
  }
  if (!exportPath.value.trim()) {
    errorMessage.value = "Укажите путь для сохранения XLSX."
    return
  }

  isExporting.value = true
  try {
    const result = await exportAnalysisXLSX(report.value, exportPath.value)
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

const tariffRows = computed(() =>
  report.value?.tariffChanges.map((item) => [
    item.vidUsl,
    item.nameUsl,
    item.previousTariff,
    item.currentTariff,
  ]) ?? [],
)

const houseRows = computed(() =>
  report.value?.houseChanges.map((item) => [item.type === "appeared" ? "Появился" : "Исчез", item.address]) ?? [],
)

const anomalyRows = computed(() =>
  report.value?.anomalies.map((item) => [
    item.vidUsl,
    item.nameUsl,
    item.previousAmount,
    item.currentAmount,
    item.deltaAmount,
    item.deltaPercent === null ? "—" : `${item.deltaPercent.toFixed(2)}%`,
  ]) ?? [],
)

const serviceTitleRows = computed(() =>
  report.value?.serviceChanges.map((item) => [
    item.type === "appeared" ? "Появилась" : "Исчезла",
    item.vidUsl,
    item.nameUsl,
  ]) ?? [],
)
</script>

<template>
  <main class="page">
    <header class="hero">
      <p>Сравнение начислений ЖКУ по поставщику</p>
      <h1>DBF Compare</h1>
      <span class="badge">MVP</span>
    </header>

    <section class="panel">
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
      <AnalysisSettings v-model="threshold" />
      <button type="button" :disabled="isAnalyzing" @click="runAnalysis">
        {{ isAnalyzing ? "Сравниваем..." : "Сравнить" }}
      </button>
      <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
      <p v-if="successMessage" class="success">{{ successMessage }}</p>
      <p class="hint">
        В dev-режиме без Wails показывается демонстрационный отчёт по тестовым DBF.
      </p>
    </section>

    <section v-if="report" class="results">
      <section class="meta">
        <div>
          <span>Поставщик</span>
          <strong>{{ report.meta.providerName }}</strong>
        </div>
        <div>
          <span>Период</span>
          <strong>{{ report.meta.previousMonth }} -> {{ report.meta.currentMonth }}</strong>
        </div>
        <div>
          <span>Записей</span>
          <strong>{{ report.meta.previousRecords }} -> {{ report.meta.currentRecords }}</strong>
        </div>
      </section>

      <SummaryCards :summary="report.summary" />
      <section class="export-panel">
        <FilePicker
          v-model="exportPath"
          label="Путь для XLSX"
          button-label="Выбрать"
          :disabled="isExporting"
          @browse="browseExport"
        />
        <button type="button" :disabled="isExporting" @click="runExport">
          {{ isExporting ? "Сохраняем..." : "Сохранить XLSX" }}
        </button>
      </section>
      <DetailsSection title="Тарифы" :columns="['VID_USL', 'NAME_USL', 'Было', 'Стало']" :rows="tariffRows" />
      <DetailsSection title="Услуги" :columns="['Тип', 'VID_USL', 'NAME_USL']" :rows="serviceTitleRows" />
      <DetailsSection title="Дома" :columns="['Тип', 'Адрес']" :rows="houseRows" />
      <DetailsSection title="Аномалии" :columns="['VID_USL', 'NAME_USL', 'Было', 'Стало', 'Дельта', 'Дельта %']" :rows="anomalyRows" />
    </section>
  </main>
</template>

<style scoped>
.page {
  background:
    radial-gradient(circle at top left, rgba(183, 149, 84, 0.35), transparent 28rem),
    linear-gradient(180deg, #f9f6ef 0%, #efe7d6 100%);
  color: #28231b;
  font-family: Georgia, "Times New Roman", serif;
  min-height: 100vh;
  padding: 2rem;
}

.hero {
  margin-bottom: 1.5rem;
}

.hero p {
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.hero h1 {
  font-size: clamp(2.5rem, 5vw, 4rem);
  margin: 0.25rem 0 0;
}

.badge {
  background: rgba(145, 109, 43, 0.12);
  border: 1px solid rgba(145, 109, 43, 0.18);
  border-radius: 999px;
  display: inline-block;
  margin-top: 1rem;
  padding: 0.3rem 0.7rem;
}

.panel,
.results {
  display: grid;
  gap: 1rem;
}

.panel {
  background: rgba(255, 252, 245, 0.82);
  border: 1px solid rgba(120, 95, 49, 0.15);
  border-radius: 20px;
  margin-bottom: 1.5rem;
  padding: 1.25rem;
}

.meta,
.export-panel {
  display: grid;
  gap: 1rem;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
}

.meta div {
  background: rgba(255, 252, 245, 0.7);
  border: 1px solid rgba(120, 95, 49, 0.12);
  border-radius: 16px;
  padding: 1rem;
}

.meta span {
  color: #6b5a3c;
  display: block;
  font-size: 0.92rem;
}

.meta strong {
  display: block;
  font-size: 1.1rem;
  margin-top: 0.2rem;
}

button {
  background: #916d2b;
  border: 0;
  border-radius: 10px;
  color: #fffdf7;
  cursor: pointer;
  padding: 0.85rem 1.2rem;
}

button:disabled {
  cursor: progress;
  opacity: 0.7;
}

.error {
  color: #9f2a1f;
}

.success {
  color: #24633d;
}

.hint {
  color: #6b5a3c;
  font-size: 0.92rem;
  margin: 0;
}
</style>
