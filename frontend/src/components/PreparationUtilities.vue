<script setup lang="ts">
import { ref } from "vue"
import {
  convertNovatekCSVToDBF,
  mergeRIRODNHotWaterRows,
  pickDBFFile,
  pickCSVFile,
  pickDBFExportPath,
  removeRIRZeroRows,
} from "../lib/backend"

defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  close: []
}>()

type UtilityDefinition = {
  id: string
  title: string
  description: string
  inputLabel: string
  actionLabel: string
}

const utilities: UtilityDefinition[] = [
  {
    id: "novatek",
    title: "Файл Новатэк",
    description: "Преобразование CSV с заголовками в DBF для последующего сравнения.",
    inputLabel: "Исходный CSV файл",
    actionLabel: "Преобразовать в DBF",
  },
  {
    id: "rir-zero",
    title: "РИР нулевые",
    description: "Удаление строк, где одновременно пусты TARIF, TARIF_DEC и NORM_USL.",
    inputLabel: "Исходный DBF файл",
    actionLabel: "Удалить пустые строки",
  },
  {
    id: "rir-odn",
    title: "РИР ГВС ОДН",
    description: "Перенос сумм NACHISL, KOPLATE и PERERASCH в строку ГВС: компонент на ХВ ОДН по заданному тарифу.",
    inputLabel: "Исходный DBF файл",
    actionLabel: "Объединить строки",
  },
]

const novatekPath = ref("")
const rirPath = ref("")
const rirOdnPath = ref("")
const rirOdnTariff = ref("32,730")
const errorMessage = ref("")
const successMessage = ref("")
const isConverting = ref(false)
const isFilteringRIR = ref(false)
const isMergingRIRODN = ref(false)

async function browseNovatekCSV() {
  errorMessage.value = ""
  try {
    const path = await pickCSVFile()
    if (path) {
      novatekPath.value = path
    }
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : "Не удалось выбрать CSV файл."
  }
}

async function runNovatekConversion() {
  errorMessage.value = ""
  successMessage.value = ""

  if (!novatekPath.value.trim()) {
    errorMessage.value = "Укажите CSV файл Новатэк."
    return
  }

  const defaultName = novatekPath.value.replace(/\.[^.]+$/, "") || "novatek"
  let savePath = ""
  try {
    savePath = await pickDBFExportPath(`${defaultName}.dbf`)
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : "Не удалось выбрать путь сохранения."
    return
  }

  if (!savePath) {
    return
  }

  isConverting.value = true
  try {
    const result = await convertNovatekCSVToDBF(novatekPath.value, savePath)
    successMessage.value = `DBF файл сохранён: ${result.path}`
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : "Не удалось преобразовать CSV в DBF."
  } finally {
    isConverting.value = false
  }
}

async function browseRIRDBF() {
  errorMessage.value = ""
  try {
    const path = await pickDBFFile()
    if (path) {
      rirPath.value = path
    }
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : "Не удалось выбрать DBF файл."
  }
}

async function runRIRCleanup() {
  errorMessage.value = ""
  successMessage.value = ""

  if (!rirPath.value.trim()) {
    errorMessage.value = "Укажите DBF файл для обработки РИР."
    return
  }

  const defaultName = rirPath.value.replace(/\.dbf$/i, "_cleaned.dbf") || "rir_cleaned.dbf"
  let savePath = ""
  try {
    savePath = await pickDBFExportPath(defaultName)
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : "Не удалось выбрать путь сохранения."
    return
  }

  if (!savePath) {
    return
  }

  isFilteringRIR.value = true
  try {
    const result = await removeRIRZeroRows(rirPath.value, savePath)
    successMessage.value = `Очищенный DBF файл сохранён: ${result.path}`
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : "Не удалось удалить пустые строки."
  } finally {
    isFilteringRIR.value = false
  }
}

async function browseRIRODNDBF() {
  errorMessage.value = ""
  try {
    const path = await pickDBFFile()
    if (path) {
      rirOdnPath.value = path
    }
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : "Не удалось выбрать DBF файл."
  }
}

async function runRIRODNMerge() {
  errorMessage.value = ""
  successMessage.value = ""

  if (!rirOdnPath.value.trim()) {
    errorMessage.value = "Укажите DBF файл для обработки РИР ГВС ОДН."
    return
  }

  if (!rirOdnTariff.value.trim()) {
    errorMessage.value = "Укажите тариф для поиска строки ГВС: компонент на ХВ ОДН."
    return
  }

  const defaultName = rirOdnPath.value.replace(/\.dbf$/i, "_odn_merged.dbf") || "rir_odn_merged.dbf"
  let savePath = ""
  try {
    savePath = await pickDBFExportPath(defaultName)
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : "Не удалось выбрать путь сохранения."
    return
  }

  if (!savePath) {
    return
  }

  isMergingRIRODN.value = true
  try {
    const result = await mergeRIRODNHotWaterRows(rirOdnPath.value, savePath, rirOdnTariff.value)
    successMessage.value = `Обработанный DBF файл сохранён: ${result.path}`
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : "Не удалось объединить строки ГВС ОДН."
  } finally {
    isMergingRIRODN.value = false
  }
}
</script>

<template>
  <div v-if="open" class="modal-overlay" @click.self="emit('close')">
    <section class="modal" aria-label="Утилиты подготовки" aria-modal="true" role="dialog">
      <div class="panel-head">
        <div>
          <h2>Утилиты подготовки</h2>
          <p>Интерфейсная заготовка для отдельных операций перед сравнением файлов.</p>
        </div>
        <div class="modal-actions">
          <span class="panel-chip">3 утилиты</span>
          <button class="close-button" type="button" @click="emit('close')">Закрыть</button>
        </div>
      </div>

      <section v-if="errorMessage || successMessage" class="messages">
        <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
        <p v-if="successMessage" class="success">{{ successMessage }}</p>
      </section>

      <div class="utilities-grid">
        <article v-for="utility in utilities" :key="utility.id" class="utility-card">
          <div class="utility-card__head">
            <div>
              <h3>{{ utility.title }}</h3>
              <p>{{ utility.description }}</p>
            </div>
          </div>

          <label class="utility-field">
            <span>{{ utility.inputLabel }}</span>
            <div class="utility-field__row">
              <input
                :value="
                  utility.id === 'novatek'
                    ? novatekPath
                    : utility.id === 'rir-zero'
                      ? rirPath
                      : utility.id === 'rir-odn'
                        ? rirOdnPath
                        : ''
                "
                type="text"
                :disabled="
                  (utility.id === 'novatek' && isConverting) ||
                  (utility.id === 'rir-zero' && isFilteringRIR) ||
                  (utility.id === 'rir-odn' && isMergingRIRODN) ||
                  (utility.id !== 'novatek' && utility.id !== 'rir-zero' && utility.id !== 'rir-odn')
                "
                :placeholder="
                  utility.id === 'novatek'
                    ? 'Выберите CSV файл с заголовками'
                    : utility.id === 'rir-zero'
                      ? 'Выберите DBF файл с услугами'
                    : utility.id === 'rir-odn'
                      ? 'Выберите DBF файл для обработки ГВС ОДН'
                    : 'Выбор файла появится на следующем этапе'
                "
                @input="
                  utility.id === 'novatek'
                    ? (novatekPath = ($event.target as HTMLInputElement).value)
                    : utility.id === 'rir-zero'
                      ? (rirPath = ($event.target as HTMLInputElement).value)
                      : utility.id === 'rir-odn'
                        ? (rirOdnPath = ($event.target as HTMLInputElement).value)
                    : undefined
                "
              />
              <button
                type="button"
                :disabled="
                  (utility.id === 'novatek' && isConverting) ||
                  (utility.id === 'rir-zero' && isFilteringRIR) ||
                  (utility.id === 'rir-odn' && isMergingRIRODN) ||
                  (utility.id !== 'novatek' && utility.id !== 'rir-zero' && utility.id !== 'rir-odn')
                "
                @click="
                  utility.id === 'novatek'
                    ? browseNovatekCSV()
                    : utility.id === 'rir-zero'
                      ? browseRIRDBF()
                      : utility.id === 'rir-odn'
                        ? browseRIRODNDBF()
                    : undefined
                "
              >
                Обзор
              </button>
            </div>
          </label>

          <label v-if="utility.id === 'rir-odn'" class="utility-field">
            <span>Тариф строки ГВС: компонент на ХВ ОДН</span>
            <input v-model="rirOdnTariff" type="text" :disabled="isMergingRIRODN" placeholder="32,730" />
          </label>

          <button
            class="secondary-button"
            type="button"
            :disabled="
              (utility.id === 'novatek' && isConverting) ||
              (utility.id === 'rir-zero' && isFilteringRIR) ||
              (utility.id === 'rir-odn' && isMergingRIRODN) ||
              (utility.id !== 'novatek' && utility.id !== 'rir-zero' && utility.id !== 'rir-odn')
            "
            @click="
              utility.id === 'novatek'
                ? runNovatekConversion()
                : utility.id === 'rir-zero'
                  ? runRIRCleanup()
                  : utility.id === 'rir-odn'
                    ? runRIRODNMerge()
                : undefined
            "
          >
            {{
              utility.id === "novatek" && isConverting
                ? "Преобразуем..."
                : utility.id === "rir-zero" && isFilteringRIR
                  ? "Обрабатываем..."
                  : utility.id === "rir-odn" && isMergingRIRODN
                    ? "Обрабатываем..."
                  : utility.actionLabel
            }}
          </button>
        </article>
      </div>
    </section>
  </div>
</template>

<style scoped>
.modal-overlay {
  align-items: center;
  background: rgba(18, 18, 18, 0.38);
  display: flex;
  inset: 0;
  justify-content: center;
  padding: 1rem;
  position: fixed;
  z-index: 20;
}

.modal {
  --utility-control-height: 3.1rem;
  background: #ffffff;
  border: 1px solid rgba(0, 0, 0, 0.06);
  border-radius: 24px;
  box-shadow: 0 22px 60px rgba(28, 23, 17, 0.18);
  max-height: min(90vh, 960px);
  max-width: 1120px;
  overflow: auto;
  padding: 1rem;
  width: min(100%, 1120px);
}

.panel-head {
  align-items: start;
  display: flex;
  gap: 1rem;
  justify-content: space-between;
  margin-bottom: 0.9rem;
}

.modal-actions {
  align-items: center;
  display: flex;
  gap: 0.65rem;
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

.close-button {
  background: #f4f4f4;
  border: 1px solid rgba(0, 0, 0, 0.08);
  border-radius: 999px;
  color: #1f1f1f;
  cursor: pointer;
  font: inherit;
  min-height: 2.45rem;
  padding: 0.55rem 0.95rem;
}

.utilities-grid {
  display: grid;
  gap: 0.75rem;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.messages {
  display: grid;
  gap: 0.5rem;
  margin-bottom: 0.9rem;
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

.utility-card {
  background: #faf9f7;
  border: 1px solid rgba(0, 0, 0, 0.05);
  border-radius: 18px;
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
  padding: 0.9rem;
}

.utility-card__head {
  display: grid;
  gap: 0.55rem;
}

.utility-card__head h3 {
  font-size: 0.94rem;
  margin: 0;
}

.utility-card__head p {
  color: #5f5f5f;
  font-size: 0.86rem;
  margin: 0;
}

.utility-status {
  background: #ffffff;
  border: 1px dashed rgba(0, 0, 0, 0.12);
  border-radius: 14px;
  color: #6f6f6f;
  font-size: 0.76rem;
  padding: 0.45rem 0.6rem;
}

.utility-field {
  display: grid;
  gap: 0.4rem;
}

.utility-field span {
  color: #7a7a7a;
  font-size: 0.82rem;
}

.utility-field__row {
  display: grid;
  gap: 0.65rem;
  grid-template-columns: minmax(0, 1fr) auto;
}

.utility-field input,
.utility-field button,
.secondary-button {
  border-radius: 16px;
  font: inherit;
  height: var(--utility-control-height);
  line-height: 1.2;
}

.utility-field input {
  background: #ffffff;
  border: 1px solid rgba(0, 0, 0, 0.08);
  display: block;
  min-width: 0;
  padding: 0 1rem;
}

.utility-field button {
  align-items: center;
  background: #f4f4f4;
  border: 1px solid rgba(0, 0, 0, 0.08);
  color: #1f1f1f;
  display: inline-flex;
  justify-content: center;
  padding: 0 1rem;
}

.secondary-button {
  align-items: center;
  background: #eceae5;
  border: 0;
  color: #4b4b4b;
  cursor: pointer;
  display: inline-flex;
  justify-content: center;
  margin-top: auto;
  padding: 0 1.15rem;
}

button:disabled,
input:disabled {
  cursor: not-allowed;
  opacity: 0.72;
}

@media (max-width: 1200px) {
  .utilities-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 760px) {
  .modal {
    border-radius: 20px;
    padding: 0.9rem;
  }

  .panel-head,
  .modal-actions {
    align-items: stretch;
    flex-direction: column;
  }
}
</style>
