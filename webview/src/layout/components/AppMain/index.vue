<template>
  <el-main class="layout-main">
    <router-view v-slot="{ Component }">
      <transition name="fade-scale" mode="out-in">
        <component :is="Component" />
      </transition>
    </router-view>
  </el-main>
</template>

<script setup lang="ts">
// AppMain component for main content area
</script>

<style scoped>
.layout-main {
  padding: 24px;
  background: var(--bg-color);
  overflow-y: auto;
  height: 100%;
  flex: 1;
}

.fade-scale-enter-active,
.fade-scale-leave-active {
  transition: opacity 0.3s ease, transform 0.3s ease;
}

.fade-scale-enter-from {
  opacity: 0;
  transform: scale(0.98);
}

.fade-scale-leave-to {
  opacity: 0;
  transform: scale(1.02);
}
</style>

