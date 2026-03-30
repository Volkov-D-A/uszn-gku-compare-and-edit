<script setup lang="ts">
import type { SummaryCounts } from "../types/analysis"

defineProps<{
  summary: SummaryCounts
}>()

const cards = [
  { key: "tariffChanges", label: "Изменилось тарифов" },
  { key: "appearedServices", label: "Появилось услуг" },
  { key: "disappearedServices", label: "Исчезло услуг" },
  { key: "appearedHouses", label: "Появилось домов" },
  { key: "disappearedHouses", label: "Исчезло домов" },
  { key: "anomalies", label: "Аномалий" },
] as const
</script>

<template>
  <div class="grid">
    <article v-for="card in cards" :key="card.key" class="card">
      <strong>{{ summary[card.key] }}</strong>
      <span>{{ card.label }}</span>
    </article>
  </div>
</template>

<style scoped>
.grid {
  display: grid;
  gap: 1rem;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
}

.card {
  background: #f5f1e7;
  border-radius: 16px;
  padding: 1rem;
}

strong {
  display: block;
  font-size: 1.8rem;
}
</style>
