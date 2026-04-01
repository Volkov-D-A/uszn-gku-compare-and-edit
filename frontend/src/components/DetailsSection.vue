<script setup lang="ts">
defineProps<{
  title: string
  columns: string[]
  rows: Array<Array<string | number | null>>
}>()
</script>

<template>
  <section class="section">
    <div class="head">
      <h2>{{ title }}</h2>
      <span>{{ rows.length }} записей</span>
    </div>
    <div class="table-wrap" v-if="rows.length > 0">
      <table>
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
    </div>
    <p v-else class="empty">Изменений нет.</p>
  </section>
</template>

<style scoped>
.section {
  background: #ffffff;
  border: 1px solid rgba(0, 0, 0, 0.06);
  border-radius: 16px;
  display: grid;
  gap: 0.8rem;
  padding: 1rem 1.05rem;
}

.head {
  align-items: baseline;
  display: flex;
  gap: 1rem;
  justify-content: space-between;
}

.head span {
  color: #7a7a7a;
  font-size: 0.8rem;
  text-transform: uppercase;
}

h2 {
  font-size: 1.05rem;
  margin: 0;
}

.table-wrap {
  overflow: auto;
}

table {
  border-collapse: collapse;
  width: 100%;
}

th {
  color: #7a7a7a;
  font-size: 0.78rem;
  text-transform: uppercase;
}

th,
td {
  border-bottom: 1px solid rgba(0, 0, 0, 0.06);
  padding: 0.68rem 0.3rem;
  text-align: left;
}

.empty {
  color: #7a7a7a;
  margin: 0;
}
</style>
