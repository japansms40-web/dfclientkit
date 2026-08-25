<script setup>
import { nextTick, ref, watch } from 'vue'

const props = defineProps({
  lines: { type: Array, required: true }, // {time, tag, kind, msg, highlight}
  autoScroll: { type: Boolean, required: true },
})
const emit = defineEmits(['update:autoScroll', 'copy', 'export', 'clear'])

const bodyEl = ref(null)

watch(
  () => props.lines.length,
  async () => {
    if (!props.autoScroll) return
    await nextTick()
    if (bodyEl.value) bodyEl.value.scrollTop = bodyEl.value.scrollHeight
  }
)

const KIND_COLOR = {
  start: '#58A6FF',
  info: '#58A6FF',
  success: '#3FB950',
  failure: '#F85149',
  retry: '#D29922',
}
function tagColor(kind) {
  return KIND_COLOR[kind] || '#8B949E'
}
function msgColor(line) {
  return line.highlight ? tagColor(line.kind) : '#C9D1D9'
}
</script>

<template>
  <div class="log-section">
    <div class="log-header">
      <span class="log-title">运行日志 RUNTIME LOG</span>
      <span class="log-count mono">{{ lines.length }} 行</span>
      <div class="spacer" />
      <label class="log-check">
        <input
          type="checkbox"
          class="sr-only"
          :checked="autoScroll"
          @change="emit('update:autoScroll', $event.target.checked)"
        />
        <span class="log-checkbox-box" :class="{ checked: autoScroll }">
          <svg v-if="autoScroll" width="9" height="9" viewBox="0 0 16 16" fill="none" stroke="#0b0e12" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round">
            <path d="M3 8.2l3.3 3.3L13 4.5" />
          </svg>
        </span>
        自动滚动
      </label>
      <span class="log-action" @click="emit('copy')">复制</span>
      <span class="log-action" @click="emit('export')">导出</span>
      <span class="log-action log-action-danger" @click="emit('clear')">清空</span>
    </div>
    <div ref="bodyEl" class="log-body mono">
      <div v-for="(l, i) in lines" :key="i" class="log-line">
        <span class="log-time">{{ l.time }}</span>
        <span class="log-tag" :style="{ color: tagColor(l.kind) }">{{ l.tag }}</span>
        <span class="log-msg" :style="{ color: msgColor(l) }">{{ l.msg }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.log-section {
  height: 224px;
  flex: 0 0 224px;
  background: var(--log-bg);
  border-top: 1px solid var(--border);
  display: flex;
  flex-direction: column;
}
.log-header {
  height: 30px;
  flex: 0 0 30px;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 12px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  font-size: 11.5px;
  color: #8b949e;
}
.log-title {
  letter-spacing: 0.06em;
}
.log-count {
  color: #6e7681;
  font-size: 11px;
}
.spacer {
  flex: 1;
}
.log-check {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
}
.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}
.log-checkbox-box {
  flex: 0 0 13px;
  width: 13px;
  height: 13px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid #3d444d;
  border-radius: 3px;
  background: transparent;
}
.log-checkbox-box.checked {
  background: #e6edf3;
  border-color: #e6edf3;
}
.log-action {
  cursor: pointer;
}
.log-action:hover {
  color: #e6edf3;
}
.log-action-danger:hover {
  color: #f85149;
}
.log-body {
  flex: 1;
  overflow: auto;
  padding: 8px 12px;
  font-size: 12px;
  line-height: 1.62;
}
.log-line {
  display: flex;
  gap: 8px;
  white-space: pre;
}
.log-time {
  color: #e3b341;
}
</style>
