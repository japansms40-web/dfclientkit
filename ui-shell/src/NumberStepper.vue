<script setup>
defineProps({
  label: { type: String, required: true },
  unit: { type: String, default: '' },
  modelValue: { type: [Number, String], required: true },
  editable: { type: Boolean, default: true },
  min: { type: Number, default: 0 },
})
const emit = defineEmits(['update:modelValue'])

function onInput(e) {
  const digits = e.target.value.replace(/[^0-9]/g, '')
  e.target.value = digits
  emit('update:modelValue', digits === '' ? 0 : Number(digits))
}
</script>

<template>
  <div class="stepper-row">
    <span class="stepper-label">{{ label }}</span>
    <input
      v-if="editable"
      class="stepper-box mono"
      type="text"
      inputmode="numeric"
      :value="modelValue"
      @input="onInput"
    />
    <span v-else class="stepper-box mono stepper-readonly">{{ modelValue }}</span>
    <span class="stepper-unit">{{ unit }}</span>
  </div>
</template>

<style scoped>
.stepper-row {
  display: flex;
  align-items: center;
  gap: 8px;
}
.stepper-label {
  flex: 1;
  font-size: 12.5px;
  text-align: right;
  color: var(--muted);
}
.stepper-box {
  width: 74px;
  height: 28px;
  border: 1px solid var(--border-strong);
  border-radius: 4px;
  background: var(--surface);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12.5px;
  color: var(--text);
  text-align: center;
  padding: 0;
}
.stepper-box:focus {
  outline: none;
  border-color: var(--accent);
}
.stepper-readonly {
  background: transparent;
  border-color: var(--border);
  color: var(--muted);
}
.stepper-unit {
  width: 20px;
  font-size: 11.5px;
  color: var(--muted);
}
</style>
