<script setup>
defineProps({
  results: { type: Array, required: true }, // 每行对象的字段由 columns 决定
  title: { type: String, default: '发布结果' },
  columns: {
    type: Array,
    default: () => [
      { key: 'ck', label: 'CK', class: 'mono' },
      { key: 'title', label: '标题', class: 'ellipsis' },
      { key: 'value', label: '结果', class: 'ellipsis mono' },
    ],
  },
})
const emit = defineEmits(['close', 'copy-all'])
</script>

<template>
  <div class="overlay" @click.self="emit('close')">
    <div class="modal">
      <div class="modal-header">
        <span class="modal-title">{{ title }}</span>
        <span class="modal-count mono">{{ results.length }} 条</span>
        <div class="spacer" />
        <button class="btn-outline" @click="emit('copy-all')">复制全部</button>
        <button class="btn-close" @click="emit('close')">×</button>
      </div>
      <div class="modal-body">
        <div v-if="results.length === 0" class="empty">暂无成功结果</div>
        <div v-for="(r, i) in results" :key="i" class="result-row">
          <span class="mono muted">{{ r.time }}</span>
          <span v-for="col in columns" :key="col.key" :class="col.class">{{ r[col.key] }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.35);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 200;
}
.modal {
  width: 560px;
  max-height: 70vh;
  display: flex;
  flex-direction: column;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 10px;
  box-shadow: var(--shadow);
  overflow: hidden;
}
.modal-header {
  flex: 0 0 auto;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 14px;
  border-bottom: 1px solid var(--border);
}
.modal-title {
  font-size: 14px;
  font-weight: 700;
}
.modal-count {
  color: var(--muted);
  font-size: 12px;
}
.spacer {
  flex: 1;
}
.btn-outline {
  height: 28px;
  padding: 0 10px;
  border-radius: 5px;
  border: 1px solid var(--border-strong);
  background: var(--surface);
  font-size: 12px;
  cursor: pointer;
  color: var(--text);
}
.btn-outline:hover {
  border-color: var(--accent);
  color: var(--accent);
}
.btn-close {
  width: 28px;
  height: 28px;
  border: none;
  background: transparent;
  color: var(--muted);
  font-size: 16px;
  cursor: pointer;
  border-radius: 5px;
}
.btn-close:hover {
  color: var(--text);
  background: var(--surface-2);
}
.modal-body {
  flex: 1;
  min-height: 0;
  overflow: auto;
  padding: 6px 14px;
}
.result-row {
  display: grid;
  grid-template-columns: 76px 140px 1fr 1fr;
  gap: 10px;
  padding: 7px 0;
  border-bottom: 1px solid var(--border);
  font-size: 12.5px;
  align-items: center;
}
.result-row:last-child {
  border-bottom: none;
}
.muted {
  color: var(--muted);
}
.ellipsis {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.empty {
  padding: 30px 0;
  text-align: center;
  color: var(--muted);
  font-size: 12.5px;
}
</style>
