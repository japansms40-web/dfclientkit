<script setup>
defineProps({
  page: { type: String, required: true },
  items: { type: Array, required: true }, // [{ key, cn, en, icon? }]
  bottomItems: { type: Array, default: () => [] }, // 渲染在下方（比如"使用说明"）
})
const emit = defineEmits(['navigate'])
</script>

<template>
  <div class="nav">
    <div class="nav-title">工作台 WORKSPACE</div>

    <button v-for="item in items" :key="item.key" class="nav-item" :class="{ active: page === item.key }" @click="emit('navigate', item.key)">
      <component :is="item.icon" v-if="item.icon" />
      <span class="nav-label">
        <span class="nav-label-cn">{{ item.cn }}</span>
        <span class="nav-label-en">{{ item.en }}</span>
      </span>
    </button>

    <div class="nav-spacer" />

    <button v-for="item in bottomItems" :key="item.key" class="nav-item" :class="{ active: page === item.key }" @click="emit('navigate', item.key)">
      <component :is="item.icon" v-if="item.icon" />
      <span class="nav-label">
        <span class="nav-label-cn">{{ item.cn }}</span>
        <span class="nav-label-en">{{ item.en }}</span>
      </span>
    </button>
  </div>
</template>

<style scoped>
.nav {
  width: 182px;
  flex: 0 0 182px;
  background: var(--nav);
  border-right: 1px solid var(--border);
  padding: 10px 8px;
  display: flex;
  flex-direction: column;
  gap: 2px;
  overflow-y: auto;
}
.nav-title {
  font-size: 10px;
  letter-spacing: 0.1em;
  color: var(--muted);
  padding: 2px 8px 6px;
  text-transform: uppercase;
}
.nav-spacer {
  flex: 1;
}
.nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  height: 40px;
  flex: 0 0 40px;
  padding: 0 10px;
  border-radius: 6px;
  border: none;
  background: transparent;
  color: var(--text);
  cursor: pointer;
  font-family: inherit;
  text-align: left;
  border-left: 3px solid transparent;
}
.nav-item svg {
  flex: 0 0 15px;
}
.nav-item:hover {
  background: var(--surface);
}
.nav-item.active {
  background: var(--accent-weak);
  color: var(--accent);
  border-left-color: var(--accent);
}
.nav-label {
  display: flex;
  flex-direction: column;
  line-height: 1.1;
  overflow: hidden;
}
.nav-label-cn {
  font-size: 13px;
  font-weight: 600;
  white-space: nowrap;
}
.nav-label-en {
  font-size: 9px;
  letter-spacing: 0.05em;
  color: var(--muted);
  white-space: nowrap;
}
.nav-item.active .nav-label-en {
  color: var(--accent);
  opacity: 0.75;
}
</style>
