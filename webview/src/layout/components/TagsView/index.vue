<template>
  <div v-if="showTagsView" class="tags-view-container">
    <ScrollPane ref="scrollPaneRef" class="tags-view-wrapper" @scroll="handleScroll">
      <router-link
        v-for="tag in visitedViews"
        :key="tag.path"
        :data-path="tag.path"
        :class="{ active: isActive(tag) }"
        :to="{ path: tag.path, query: tag.query }"
        class="tags-view-item"
        :style="activeStyle(tag)"
        @click.middle="!isAffix(tag) ? closeSelectedTag(tag) : ''"
        @contextmenu.prevent="openMenu(tag, $event)"
      >
        <span class="tags-view-item-title">{{ (tag as TagView).title || tag.name }}</span>
        <span v-if="!isAffix(tag)" @click.prevent.stop="closeSelectedTag(tag)" class="tags-view-item-close">
          <el-icon><Close /></el-icon>
        </span>
      </router-link>
    </ScrollPane>

    <!-- 右键菜单 -->
    <ul v-show="visible" :style="{ left: left + 'px', top: top + 'px' }" class="contextmenu">
      <li v-if="selectedTag" @click="refreshSelectedTag(selectedTag)">
        <el-icon><RefreshRight /></el-icon>
        {{ t('tagsView.refresh') }}
      </li>
      <li v-if="selectedTag && !isAffix(selectedTag)" @click="closeSelectedTag(selectedTag)">
        <el-icon><Close /></el-icon>
        {{ t('tagsView.closeCurrent') }}
      </li>
      <li v-if="selectedTag" @click="closeOthersTags">
        <el-icon><CircleClose /></el-icon>
        {{ t('tagsView.closeOthers') }}
      </li>
      <li v-if="selectedTag && !isFirstView()" @click="closeLeftTags">
        <el-icon><Back /></el-icon>
        {{ t('tagsView.closeLeft') }}
      </li>
      <li v-if="selectedTag && !isLastView()" @click="closeRightTags">
        <el-icon><Right /></el-icon>
        {{ t('tagsView.closeRight') }}
      </li>
      <li v-if="selectedTag" @click="closeAllTags">
        <el-icon><CircleClose /></el-icon>
        {{ t('tagsView.closeAll') }}
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
  import { useI18n } from '@/composables'
  import ScrollPane from './ScrollPane.vue'
  import type { RouteLocationNormalized } from 'vue-router'
  import { useTagsViewStore, useLayoutStore } from '@/stores'

  // 扩展 RouteLocationNormalized 类型以包含 title
  type TagView = RouteLocationNormalized & {
    title?: string
  }

  const route = useRoute()
  const router = useRouter()
  const tagsViewStore = useTagsViewStore()
  const layoutStore = useLayoutStore()
  const { t } = useI18n()

  const visible = ref(false)
  const top = ref(0)
  const left = ref(0)
  const selectedTag = ref<TagView>()
  const scrollPaneRef = ref<InstanceType<typeof ScrollPane>>()

  // 是否显示标签页（从布局设置中读取）
  const showTagsView = computed(() => layoutStore.tagsViewVisible)

  const visitedViews = computed(() => tagsViewStore.getVisitedViews)

  const isActive = (r: RouteLocationNormalized): boolean => {
    return r.path === route.path
  }

  const activeStyle = (tag: RouteLocationNormalized) => {
    if (!isActive(tag)) return {}
    return {
      backgroundColor: 'var(--tags-view-active-bg, var(--primary-color))',
      borderColor: 'var(--tags-view-active-border-color, var(--primary-color))',
      color: '#fff'
    }
  }

  const isAffix = (tag: RouteLocationNormalized) => {
    return tag?.meta && tag?.meta?.affix
  }

  const isFirstView = () => {
    try {
      return selectedTag.value?.fullPath === '/files' || selectedTag.value?.fullPath === visitedViews.value[0]?.fullPath
    } catch (err) {
      return false
    }
  }

  const isLastView = () => {
    try {
      return selectedTag.value?.fullPath === visitedViews.value[visitedViews.value.length - 1]?.fullPath
    } catch (err) {
      return false
    }
  }

  const addTags = () => {
    const { name } = route
    if (name) {
      tagsViewStore.addView(route as RouteLocationNormalized)
    }
  }

  const moveToCurrentTag = () => {
    nextTick(() => {
      for (const r of visitedViews.value) {
        if (r.path === route.path) {
          scrollPaneRef.value?.moveToTarget(r, visitedViews.value)
          if (r.fullPath !== route.fullPath) {
            tagsViewStore.updateVisitedView(route as RouteLocationNormalized)
          }
        }
      }
    })
  }

  const refreshSelectedTag = async (view?: RouteLocationNormalized) => {
    if (!view) return

    // 删除缓存视图以强制重新渲染
    tagsViewStore.delCachedView(view)

    // 使用 redirect 路由来刷新页面
    await router.replace({
      path: '/redirect' + view.path,
      query: view.query
    })

    closeMenu()
  }

  const closeSelectedTag = (view: RouteLocationNormalized) => {
    tagsViewStore.delView(view)
    if (isActive(view)) {
      toLastView()
    }
    closeMenu()
  }

  const closeRightTags = () => {
    if (selectedTag.value) {
      tagsViewStore.delRightTags(selectedTag.value)
      if (!visitedViews.value.find(i => i.fullPath === route.fullPath)) {
        toLastView()
      }
    }
    closeMenu()
  }

  const closeLeftTags = () => {
    if (selectedTag.value) {
      tagsViewStore.delLeftTags(selectedTag.value)
      if (!visitedViews.value.find(i => i.fullPath === route.fullPath)) {
        toLastView()
      }
    }
    closeMenu()
  }

  const closeOthersTags = () => {
    if (selectedTag.value) {
      router.push(selectedTag.value).catch(() => {})
      tagsViewStore.delOthersViews(selectedTag.value)
      moveToCurrentTag()
    }
    closeMenu()
  }

  const closeAllTags = () => {
    tagsViewStore.delAllViews()
    if (visitedViews.value.some(tag => tag.path === route.path)) {
      return
    }
    toLastView()
    closeMenu()
  }

  const toLastView = () => {
    const latestView = visitedViews.value.slice(-1)[0]
    if (latestView) {
      router.push(latestView.fullPath as string)
    } else {
      router.push('/files')
    }
  }

  const openMenu = (tag: RouteLocationNormalized, e: MouseEvent) => {
    const menuMinWidth = 150
    const container = document.querySelector('.tags-view-container') as HTMLElement
    if (!container) return

    const offsetLeft = container.getBoundingClientRect().left
    const offsetWidth = container.offsetWidth
    const maxLeft = offsetWidth - menuMinWidth
    const l = e.clientX - offsetLeft + 15

    if (l > maxLeft) {
      left.value = maxLeft
    } else {
      left.value = l
    }

    top.value = e.clientY
    visible.value = true
    selectedTag.value = tag
  }

  const closeMenu = () => {
    visible.value = false
  }

  const handleScroll = () => {
    closeMenu()
  }

  watch(route, () => {
    addTags()
    moveToCurrentTag()
  })

  watch(visible, value => {
    if (value) {
      document.body.addEventListener('click', closeMenu)
    } else {
      document.body.removeEventListener('click', closeMenu)
    }
  })

  onMounted(() => {
    addTags()
    moveToCurrentTag()
  })
</script>

<style scoped>
  .tags-view-container {
    height: 34px;
    width: 100%;
    background-color: var(--el-bg-color);
    border-bottom: 1px solid var(--el-border-color-lighter);
    box-shadow:
      0 1px 3px 0 rgba(0, 0, 0, 0.12),
      0 0 3px 0 rgba(0, 0, 0, 0.04);
  }

  .tags-view-wrapper {
    height: 100%;
  }

  .tags-view-item {
    display: inline-block;
    position: relative;
    cursor: pointer;
    height: 26px;
    line-height: 26px;
    background-color: var(--el-bg-color);
    border: 1px solid var(--el-border-color-lighter);
    color: var(--el-text-color-regular);
    padding: 0 8px;
    font-size: 12px;
    margin-left: 5px;
    margin-top: 4px;
    border-radius: 4px;
    transition: all 0.2s ease;
    text-decoration: none;
  }

  .tags-view-item:hover {
    color: var(--primary-color);
    border-color: var(--primary-color);
  }

  .tags-view-item:first-of-type {
    margin-left: 15px;
  }

  .tags-view-item:last-of-type {
    margin-right: 15px;
  }

  .tags-view-item.active {
    background-color: var(--primary-color);
    color: #fff;
    border-color: var(--primary-color);
  }

  .tags-view-item.active::before {
    content: '';
    background: #fff;
    display: inline-block;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    position: relative;
    margin-right: 5px;
  }

  .tags-view-item-title {
    margin-right: 4px;
  }

  .tags-view-item-close {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 16px;
    height: 16px;
    border-radius: 50%;
    transition: all 0.2s ease;
    margin-left: 4px;
  }

  .tags-view-item-close:hover {
    background-color: rgba(255, 255, 255, 0.3);
    color: #fff;
  }

  .tags-view-item.active .tags-view-item-close:hover {
    background-color: rgba(255, 255, 255, 0.2);
  }

  .contextmenu {
    margin: 0;
    background: var(--el-bg-color);
    z-index: 3000;
    position: fixed;
    list-style-type: none;
    padding: 5px 0;
    border-radius: 4px;
    font-size: 12px;
    font-weight: 400;
    box-shadow: 2px 2px 3px 0 rgba(0, 0, 0, 0.3);
    border: 1px solid var(--el-border-color-lighter);
  }

  .contextmenu li {
    margin: 0;
    padding: 7px 16px;
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: 8px;
    transition: background-color 0.2s ease;
  }

  .contextmenu li:hover {
    background: var(--el-fill-color-light);
  }

  html.dark .contextmenu {
    box-shadow: 2px 2px 3px 0 rgba(0, 0, 0, 0.5);
  }

  html.dark .contextmenu li:hover {
    background: var(--el-fill-color-light);
  }
</style>
