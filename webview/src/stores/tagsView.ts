import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { RouteLocationNormalized } from 'vue-router'
import { $t } from '@/i18n'
import { StoreId } from '@/enums/StoreId'

export const useTagsViewStore = defineStore(StoreId.TagsView, () => {
  const visitedViews = ref<RouteLocationNormalized[]>([])
  const cachedViews = ref<string[]>([])

  const getVisitedViews = computed(() => visitedViews.value)
  const getCachedViews = computed(() => cachedViews.value)

  /**
   * 添加访问的视图
   */
  const addVisitedView = (view: RouteLocationNormalized) => {
    if (visitedViews.value.some(v => v.path === view.path)) {
      return
    }

    // 获取国际化标题
    const title = view.meta?.i18nKey
      ? $t(view.meta.i18nKey as string)
      : (view.meta?.title as string) || view.name || 'no-name'

    visitedViews.value.push(
      Object.assign({}, view, {
        title
      })
    )
  }

  /**
   * 添加缓存的视图
   */
  const addCachedView = (view: RouteLocationNormalized) => {
    const viewName = view.name as string
    if (!viewName) return
    if (cachedViews.value.includes(viewName)) return
    if (!view.meta?.noCache) {
      cachedViews.value.push(viewName)
    }
  }

  /**
   * 添加视图（同时添加到访问列表和缓存列表）
   */
  const addView = (view: RouteLocationNormalized) => {
    addVisitedView(view)
    addCachedView(view)
  }

  /**
   * 删除访问的视图
   */
  const delVisitedView = (view: RouteLocationNormalized) => {
    for (const [i, v] of visitedViews.value.entries()) {
      if (v.path === view.path) {
        visitedViews.value.splice(i, 1)
        break
      }
    }
  }

  /**
   * 删除缓存的视图
   */
  const delCachedView = (view: RouteLocationNormalized) => {
    const viewName = view.name as string
    if (!viewName) return
    const index = cachedViews.value.indexOf(viewName)
    if (index > -1) {
      cachedViews.value.splice(index, 1)
    }
  }

  /**
   * 删除视图（同时从访问列表和缓存列表删除）
   */
  const delView = (view: RouteLocationNormalized) => {
    delVisitedView(view)
    if (!isDynamicRoute(view)) {
      delCachedView(view)
    }
  }

  /**
   * 删除其他视图（保留当前视图和固定视图）
   */
  const delOthersViews = (view: RouteLocationNormalized) => {
    visitedViews.value = visitedViews.value.filter(v => {
      return v.meta?.affix || v.path === view.path
    })

    const viewName = view.name as string
    if (viewName) {
      const index = cachedViews.value.indexOf(viewName)
      if (index > -1) {
        cachedViews.value = cachedViews.value.slice(index, index + 1)
      } else {
        cachedViews.value = []
      }
    } else {
      cachedViews.value = []
    }
  }

  /**
   * 删除所有视图（保留固定视图）
   */
  const delAllViews = () => {
    visitedViews.value = visitedViews.value.filter(tag => tag.meta?.affix)
    cachedViews.value = []
  }

  /**
   * 删除右侧标签
   */
  const delRightTags = (view: RouteLocationNormalized) => {
    const index = visitedViews.value.findIndex(v => v.path === view.path)
    if (index === -1) return

    visitedViews.value = visitedViews.value.filter((item, idx) => {
      if (idx <= index || (item.meta && item.meta.affix)) {
        return true
      }
      const i = cachedViews.value.indexOf(item.name as string)
      if (i > -1) {
        cachedViews.value.splice(i, 1)
      }
      return false
    })
  }

  /**
   * 删除左侧标签
   */
  const delLeftTags = (view: RouteLocationNormalized) => {
    const index = visitedViews.value.findIndex(v => v.path === view.path)
    if (index === -1) return

    visitedViews.value = visitedViews.value.filter((item, idx) => {
      if (idx >= index || (item.meta && item.meta.affix)) {
        return true
      }
      const i = cachedViews.value.indexOf(item.name as string)
      if (i > -1) {
        cachedViews.value.splice(i, 1)
      }
      return false
    })
  }

  /**
   * 更新访问的视图
   */
  const updateVisitedView = (view: RouteLocationNormalized) => {
    for (let v of visitedViews.value) {
      if (v.path === view.path) {
        const title = view.meta?.i18nKey
          ? $t(view.meta.i18nKey as string)
          : (view.meta?.title as string) || view.name || 'no-name'
        v = Object.assign(v, view, { title })
        break
      }
    }
  }

  /**
   * 检查是否为动态路由
   */
  const isDynamicRoute = (view: RouteLocationNormalized): boolean => {
    return view.matched.some(m => m.path.includes(':'))
  }

  return {
    visitedViews,
    cachedViews,
    getVisitedViews,
    getCachedViews,
    addVisitedView,
    addCachedView,
    addView,
    delVisitedView,
    delCachedView,
    delView,
    delOthersViews,
    delAllViews,
    delRightTags,
    delLeftTags,
    updateVisitedView
  }
})
