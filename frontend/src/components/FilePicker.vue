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
  gap: 0.4rem;
}

span {
  color: #7a7a7a;
  font-size: 0.82rem;
}

.row {
  display: grid;
  gap: 0.75rem;
  grid-template-columns: minmax(0, 1fr) auto;
}

input {
  background: #fbfbfb;
  border: 1px solid rgba(0, 0, 0, 0.08);
  border-radius: 16px;
  min-width: 0;
  padding: 0.85rem 1rem;
}

button {
  background: #f4f4f4;
  border: 1px solid rgba(0, 0, 0, 0.08);
  border-radius: 16px;
  color: #1f1f1f;
  cursor: pointer;
  min-height: 3.25rem;
  padding: 0.7rem 1rem;
}

button:disabled,
input:disabled {
  cursor: not-allowed;
  opacity: 0.7;
}
</style>
