<script setup lang="ts">
defineProps<{
  label: string
  modelValue: string
  buttonLabel?: string
  disabled?: boolean
}>()

const emit = defineEmits<{
  "update:modelValue": [value: string]
  browse: []
}>()

function onChange(event: Event) {
  const target = event.target as HTMLInputElement
  emit("update:modelValue", target.value)
}
</script>

<template>
  <label class="field">
    <span>{{ label }}</span>
    <div class="row">
      <input :value="modelValue" type="text" @input="onChange" placeholder="Введите путь" :disabled="disabled" />
      <button type="button" :disabled="disabled" @click="emit('browse')">
        {{ buttonLabel ?? "Обзор" }}
      </button>
    </div>
  </label>
</template>

<style scoped>
.field {
  display: grid;
  gap: 0.35rem;
}

.row {
  display: grid;
  gap: 0.75rem;
  grid-template-columns: minmax(0, 1fr) auto;
}

input {
  border: 1px solid #c4c4c4;
  border-radius: 8px;
  min-width: 0;
  padding: 0.7rem 0.9rem;
}

button {
  background: #e6dcc5;
  border: 1px solid #c8b489;
  border-radius: 8px;
  color: #3c2d16;
  cursor: pointer;
  padding: 0.7rem 1rem;
}

button:disabled,
input:disabled {
  cursor: not-allowed;
  opacity: 0.7;
}
</style>
