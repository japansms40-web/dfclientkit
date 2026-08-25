<script setup>
defineProps({
  appName: { type: String, required: true },
  logoText: { type: String, required: true },
  theme: { type: String, required: true }, // 'light' | 'dark'
  version: { type: String, required: true },
  pillText: { type: String, required: true },
  pillKind: { type: String, required: true }, // 'muted' | 'running' | 'success' | 'error'
})
const emit = defineEmits(['toggle-theme', 'minimize', 'maximize', 'close'])
</script>

<template>
  <div class="titlebar">
    <div class="logo">{{ logoText }}</div>
    <span class="app-name">{{ appName }}</span>
    <span class="version-pill">{{ version }}</span>
    <div class="spacer" />
    <div class="pill" :class="`pill-${pillKind}`">
      <span class="pill-dot" />
      <span class="pill-text">{{ pillText }}</span>
    </div>
    <button class="icon-btn theme-toggle" :title="theme === 'light' ? '切换为深色' : '切换为浅色'" @click="emit('toggle-theme')">
      <svg v-if="theme === 'light'" width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4">
        <path d="M14 9.2A6 6 0 016.8 2 6.5 6.5 0 1014 9.2z" />
      </svg>
      <svg v-else width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4">
        <circle cx="8" cy="8" r="3.2" /><path d="M8 1v1.6M8 13.4V15M1 8h1.6M13.4 8H15M3 3l1.1 1.1M11.9 11.9L13 13M13 3l-1.1 1.1M4.1 11.9L3 13" />
      </svg>
    </button>
    <div class="window-controls">
      <button class="icon-btn" title="最小化" @click="emit('minimize')">
        <svg width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4">
          <path d="M3 12h10" stroke-linecap="round" />
        </svg>
      </button>
      <button class="icon-btn" title="最大化 / 还原" @click="emit('maximize')">
        <svg width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4">
          <rect x="3.5" y="3.5" width="9" height="9" rx="1.5" />
        </svg>
      </button>
      <button class="icon-btn close-btn" title="关闭" @click="emit('close')">
        <svg width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4">
          <path d="M4 4l8 8M12 4l-8 8" stroke-linecap="round" />
        </svg>
      </button>
    </div>
  </div>
</template>

<style scoped>
.titlebar {
  --wails-draggable: drag;
  height: 52px;
  flex: 0 0 52px;
  margin: 10px 10px 0;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 0 10px 0 14px;
  background: var(--nav);
  border: 1px solid var(--border);
  border-radius: 14px;
  box-shadow: 0 4px 16px rgba(16, 22, 32, 0.1);
}
.logo {
  width: 24px;
  height: 24px;
  border-radius: 6px;
  background: var(--accent);
  color: #fff;
  font-size: 12px;
  font-weight: 700;
  font-family: ui-monospace, "SF Mono", Consolas, monospace;
  display: flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 24px;
}
.app-name {
  font-size: 14.5px;
  font-weight: 700;
}
.version-pill {
  padding: 3px 8px;
  border-radius: 10px;
  background: var(--surface);
  border: 1px solid var(--border);
  color: var(--muted);
  font-size: 11px;
  font-weight: 500;
  font-family: ui-monospace, "SF Mono", Consolas, monospace;
}
.spacer {
  flex: 1;
}
.pill {
  display: flex;
  align-items: center;
  gap: 7px;
  padding: 5px 11px;
  border-radius: 14px;
  font-size: 12px;
  font-weight: 500;
}
.pill-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: currentColor;
}
.pill-muted {
  background: var(--surface);
  color: var(--muted);
}
.pill-running {
  background: rgba(63, 185, 80, 0.14);
  color: var(--ok);
}
.pill-running .pill-dot {
  animation: pulse 1.4s ease-in-out infinite;
}
.pill-success {
  background: rgba(31, 111, 235, 0.12);
  color: var(--accent);
}
.pill-error {
  background: rgba(248, 81, 73, 0.14);
  color: var(--err);
}
.icon-btn {
  --wails-draggable: no-drag;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 8px;
  border: 1px solid var(--border);
  background: transparent;
  color: var(--muted);
  cursor: pointer;
  flex: 0 0 28px;
}
.icon-btn:hover {
  color: var(--text);
  border-color: var(--border-strong);
  background: var(--surface);
}
.window-controls {
  display: flex;
  align-items: center;
  gap: 6px;
}
.close-btn:hover {
  color: #fff;
  border-color: var(--err);
  background: var(--err);
}
</style>
