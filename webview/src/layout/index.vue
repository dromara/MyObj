<template>
  <div class="layout-container">
    <!-- 背景图案 -->
    <BackgroundPattern :pattern="backgroundPattern" />
    
    <el-container direction="vertical" class="layout-content">
      <Header />
      <el-container class="main-container">
        <Sidebar />
        <AppMain />
      </el-container>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { Header, Sidebar, AppMain } from './components'

// 背景图案设置（可以从设置页面配置）
const backgroundPattern = ref<'none' | 'grid' | 'dots' | 'gradient' | 'waves' | 'particles'>('none')

// 监听背景图案变化事件
const handleBackgroundPatternChange = (event: Event) => {
  const customEvent = event as CustomEvent<{ pattern: string }>
  if (customEvent.detail?.pattern) {
    const pattern = customEvent.detail.pattern
    if (['none', 'grid', 'dots', 'gradient', 'waves', 'particles'].includes(pattern)) {
      backgroundPattern.value = pattern as any
    }
  }
}

// 从 localStorage 加载背景图案设置
const loadBackgroundPattern = () => {
  const saved = localStorage.getItem('backgroundPattern')
  if (saved && ['none', 'grid', 'dots', 'gradient', 'waves', 'particles'].includes(saved)) {
    backgroundPattern.value = saved as any
  }
}

// 组件挂载时加载设置并监听事件
onMounted(() => {
  loadBackgroundPattern()
  
  // 添加事件监听器（使用 capture 确保能捕获到事件）
  window.addEventListener('background-pattern-changed', handleBackgroundPatternChange, true)
})

// 清理事件监听器
onBeforeUnmount(() => {
  window.removeEventListener('background-pattern-changed', handleBackgroundPatternChange, true)
})
</script>

<style scoped>
.layout-container {
  width: 100%;
  height: 100vh;
  background: var(--bg-color);
  overflow: hidden;
  position: relative;
}

.layout-content {
  position: relative;
  z-index: 1;
}

.layout-container :deep(.el-container) {
  height: 100%;
}

.main-container {
  height: calc(100vh - 64px);
  overflow: hidden;
}
</style>
