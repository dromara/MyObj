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

// 从 localStorage 加载背景图案设置
onMounted(() => {
  const saved = localStorage.getItem('backgroundPattern')
  if (saved && ['none', 'grid', 'dots', 'gradient', 'waves', 'particles'].includes(saved)) {
    backgroundPattern.value = saved as any
  }
  
  // 监听背景图案变化事件
  const handleBackgroundPatternChange = (event: CustomEvent) => {
    backgroundPattern.value = event.detail.pattern
  }
  window.addEventListener('background-pattern-changed', handleBackgroundPatternChange as EventListener)
  
  onBeforeUnmount(() => {
    window.removeEventListener('background-pattern-changed', handleBackgroundPatternChange as EventListener)
  })
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
