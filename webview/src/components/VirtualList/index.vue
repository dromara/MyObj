<template>
  <div
    ref="containerRefElement"
    class="virtual-list-container"
    :style="{ height: `${containerHeight}px`, overflow: 'auto' }"
  >
    <div class="virtual-list-wrapper" :style="{ height: `${totalHeight}px`, position: 'relative' }">
      <div class="virtual-list-content" :style="{ transform: `translateY(${offsetY}px)` }">
        <slot v-for="index in visibleItems" :key="index" :index="index" :item="items[index]" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { useVirtualScroll } from '@/composables'

  interface Props<T = any> {
    /** 数据列表 */
    items: T[]
    /** 每项高度（像素） */
    itemHeight: number
    /** 容器高度（像素） */
    containerHeight: number
    /** 缓冲区大小 */
    buffer?: number
  }

  const props = withDefaults(defineProps<Props>(), {
    buffer: 3
  })

  const virtualScroll = useVirtualScroll({
    itemHeight: props.itemHeight,
    containerHeight: props.containerHeight,
    total: props.items.length,
    buffer: props.buffer
  })

  const visibleItems = computed(() => virtualScroll.visibleItems.value)
  const totalHeight = computed(() => virtualScroll.totalHeight.value)
  const offsetY = computed(() => virtualScroll.offsetY.value)
  // 容器引用，用于绑定到模板（在模板中使用）
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const containerRefElement = virtualScroll.containerRef
</script>

<style scoped>
  .virtual-list-container {
    position: relative;
    overflow: auto;
    -webkit-overflow-scrolling: touch;
  }

  .virtual-list-wrapper {
    position: relative;
    width: 100%;
  }

  .virtual-list-content {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    will-change: transform;
  }
</style>
