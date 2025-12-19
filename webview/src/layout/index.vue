<template>
  <div class="layout-container">
    <el-container direction="vertical">
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
</script>

<style scoped>
.layout-container {
  width: 100%;
  height: 100vh;
  background: var(--bg-color);
  overflow: hidden;
}

.layout-container :deep(.el-container) {
  height: 100%;
}

.main-container {
  height: calc(100vh - 64px);
  overflow: hidden;
}
</style>
