<template>
  <el-scrollbar ref="scrollContainerRef" :vertical="false" class="scroll-container" @wheel.prevent="handleScroll">
    <slot />
  </el-scrollbar>
</template>

<script setup lang="ts">
  import type { RouteLocationNormalized } from 'vue-router'

  const tagAndTagSpacing = ref(4)
  const scrollContainerRef = ref<any>()
  const scrollWrapper = computed(() => scrollContainerRef.value?.$refs.wrapRef as HTMLElement)

  const emit = defineEmits<{
    scroll: []
  }>()

  const handleScroll = (e: WheelEvent) => {
    const eventDelta = (e as any).wheelDelta || -e.deltaY * 40
    const $scrollWrapper = scrollWrapper.value
    if ($scrollWrapper) {
      $scrollWrapper.scrollLeft = $scrollWrapper.scrollLeft + eventDelta / 4
    }
  }

  const emitScroll = () => {
    emit('scroll')
  }

  onMounted(() => {
    scrollWrapper.value?.addEventListener('scroll', emitScroll, true)
  })

  onBeforeUnmount(() => {
    scrollWrapper.value?.removeEventListener('scroll', emitScroll)
  })

  /**
   * 移动到目标标签
   */
  const moveToTarget = (currentTag: RouteLocationNormalized, visitedViews: RouteLocationNormalized[]) => {
    const $container = scrollContainerRef.value?.$el as HTMLElement
    if (!$container) return

    const $containerWidth = $container.offsetWidth
    const $scrollWrapper = scrollWrapper.value
    if (!$scrollWrapper) return

    let firstTag: RouteLocationNormalized | null = null
    let lastTag: RouteLocationNormalized | null = null

    if (visitedViews.length > 0) {
      firstTag = visitedViews[0]
      lastTag = visitedViews[visitedViews.length - 1]
    }

    if (firstTag === currentTag) {
      $scrollWrapper.scrollLeft = 0
    } else if (lastTag === currentTag) {
      $scrollWrapper.scrollLeft = $scrollWrapper.scrollWidth - $containerWidth
    } else {
      const tagListDom = document.getElementsByClassName('tags-view-item') as HTMLCollectionOf<HTMLElement>
      const currentIndex = visitedViews.findIndex(item => item === currentTag)
      let prevTag: HTMLElement | null = null
      let nextTag: HTMLElement | null = null

      for (const k in tagListDom) {
        if (k !== 'length' && Object.hasOwnProperty.call(tagListDom, k)) {
          const dom = tagListDom[k]
          const path = dom.getAttribute('data-path')
          if (path === visitedViews[currentIndex - 1]?.path) {
            prevTag = dom
          }
          if (path === visitedViews[currentIndex + 1]?.path) {
            nextTag = dom
          }
        }
      }

      if (prevTag && nextTag) {
        const afterNextTagOffsetLeft = nextTag.offsetLeft + nextTag.offsetWidth + tagAndTagSpacing.value
        const beforePrevTagOffsetLeft = prevTag.offsetLeft - tagAndTagSpacing.value

        if (afterNextTagOffsetLeft > $scrollWrapper.scrollLeft + $containerWidth) {
          $scrollWrapper.scrollLeft = afterNextTagOffsetLeft - $containerWidth
        } else if (beforePrevTagOffsetLeft < $scrollWrapper.scrollLeft) {
          $scrollWrapper.scrollLeft = beforePrevTagOffsetLeft
        }
      }
    }
  }

  defineExpose({
    moveToTarget
  })
</script>

<style scoped>
  .scroll-container {
    white-space: nowrap;
    position: relative;
    overflow: hidden;
    width: 100%;
  }

  .scroll-container :deep(.el-scrollbar__bar) {
    bottom: 0px;
  }

  .scroll-container :deep(.el-scrollbar__wrap) {
    height: 49px;
  }
</style>
