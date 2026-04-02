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
  grid-template-columns: repeat(5, minmax(0, 1fr));
  width: 100%;
}

.card {
  background: #ffffff;
  border: 1px solid rgba(0, 0, 0, 0.06);
  border-radius: 16px;
  display: grid;
  gap: 0.25rem;
  min-height: 6.5rem;
  min-width: 0;
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

@media (max-width: 1100px) {
  .grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

@media (max-width: 760px) {
  .grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 520px) {
  .grid {
    grid-template-columns: 1fr;
  }
}
</style>
