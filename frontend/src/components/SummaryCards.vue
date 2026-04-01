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
] as const
</script>

<template>
  <div class="grid">
    <article v-for="card in cards" :key="card.key" class="card">
      <span class="label">{{ card.label }}</span>
      <strong>{{ summary[card.key] }}</strong>
    </article>
  </div>
</template>

<style scoped>
.grid {
  display: grid;
  gap: 1rem;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
}

.card {
  background: #ffffff;
  border: 1px solid rgba(0, 0, 0, 0.06);
  border-radius: 16px;
  display: grid;
  gap: 0.25rem;
  min-height: 6.5rem;
  padding: 0.9rem 1rem;
}

.label {
  color: #767676;
  font-size: 0.9rem;
}

strong {
  display: block;
  font-size: 1.7rem;
  line-height: 1;
}
</style>
