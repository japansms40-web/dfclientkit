<script setup>
defineProps({
  total: { type: Number, required: true },
  success: { type: Number, required: true },
  fail: { type: Number, required: true },
  pending: { type: Number, required: true },
  elapsed: { type: String, required: true },
})
</script>

<template>
  <div class="status-bar">
    <div class="cell">
      <span class="label">账号</span><span class="mono">{{ total }}</span>
    </div>
    <div class="cell">
      <span class="label">成功</span><span class="mono ok">{{ success }}</span>
    </div>
    <div class="cell">
      <span class="label">失败</span><span class="mono err">{{ fail }}</span>
    </div>
    <div class="cell no-border">
      <span class="label">待发</span><span class="mono">{{ pending }}</span>
    </div>
    <div class="spacer" />
    <div class="cell no-border">
      <span class="label">已用时</span><span class="mono">{{ elapsed }}</span>
    </div>
  </div>
</template>

<style scoped>
.status-bar {
  height: 36px;
  flex: 0 0 36px;
  display: flex;
  align-items: center;
  background: var(--nav);
  border-top: 1px solid var(--border);
  font-size: 12px;
}
.cell {
  display: flex;
  align-items: center;
  gap: 7px;
  padding: 0 14px;
  border-right: 1px solid var(--border);
  height: 100%;
}
.cell.no-border {
  border-right: none;
}
.label {
  color: var(--muted);
}
.ok {
  color: var(--ok);
  font-weight: 500;
}
.err {
  color: var(--err);
  font-weight: 500;
}
.spacer {
  flex: 1;
}
</style>
