<script setup lang="ts">
defineProps<{
  modelValue: number
}>()

const emit = defineEmits<{
  "update:modelValue": [value: number]
}>()

function onInput(event: Event) {
  const target = event.target as HTMLInputElement
  emit("update:modelValue", Number(target.value))
}
</script>

<template>
  <label class="field">
    <span>Порог аномалий, %</span>
    <input :value="modelValue" type="number" min="0" step="1" @input="onInput" />
  </label>
</template>

<style scoped>
.field {
  display: grid;
  gap: 0.35rem;
}

input {
  border: 1px solid #c4c4c4;
  border-radius: 8px;
  padding: 0.7rem 0.9rem;
}
</style>
