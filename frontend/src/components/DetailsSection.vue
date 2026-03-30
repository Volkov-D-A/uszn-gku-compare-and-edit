<script setup lang="ts">
defineProps<{
  title: string
  columns: string[]
  rows: Array<Array<string | number | null>>
}>()
</script>

<template>
  <section class="section">
    <h2>{{ title }}</h2>
    <table v-if="rows.length > 0">
      <thead>
        <tr>
          <th v-for="column in columns" :key="column">{{ column }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(row, index) in rows" :key="index">
          <td v-for="(cell, cellIndex) in row" :key="cellIndex">{{ cell ?? "—" }}</td>
        </tr>
      </tbody>
    </table>
    <p v-else>Изменений нет.</p>
  </section>
</template>

<style scoped>
.section {
  display: grid;
  gap: 0.75rem;
}

table {
  border-collapse: collapse;
  width: 100%;
}

th,
td {
  border-bottom: 1px solid #ddd7ca;
  padding: 0.65rem;
  text-align: left;
}
</style>
