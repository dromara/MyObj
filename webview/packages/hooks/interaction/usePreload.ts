/**
 * 资源预加载 Composable
 * 用于预加载关键资源，提升页面加载性能
 */

export interface PreloadOptions {
  /** 资源类型 */
  type?: 'script' | 'style' | 'image' | 'font' | 'fetch'
  /** 资源地址 */
  href: string
  /** 是否跨域 */
  crossorigin?: boolean
  /** 媒体查询（用于响应式资源） */
  media?: string
  /** 优先级 */
  priority?: 'high' | 'low' | 'auto'
}

/**
 * 预加载资源
 * @param options 预加载配置
 */
export function preloadResource(options: PreloadOptions): Promise<void> {
  return new Promise((resolve, reject) => {
    const { type = 'fetch', href, crossorigin, media, priority = 'auto' } = options

    if (typeof document === 'undefined') {
      resolve()
      return
    }

    const link = document.createElement('link')
    link.rel = 'preload'
    link.href = href
    link.as = type

    if (crossorigin) {
      link.crossOrigin = 'anonymous'
    }

    if (media) {
      link.media = media
    }

    if (priority === 'high') {
      link.setAttribute('fetchpriority', 'high')
    } else if (priority === 'low') {
      link.setAttribute('fetchpriority', 'low')
    }

    link.onload = () => resolve()
    link.onerror = () => reject(new Error(`Failed to preload resource: ${href}`))

    document.head.appendChild(link)
  })
}

/**
 * 预连接（DNS 预解析和 TCP 预连接）
 * @param url 目标 URL
 */
export function preconnect(url: string): void {
  if (typeof document === 'undefined') return

  const link = document.createElement('link')
  link.rel = 'preconnect'
  link.href = url
  link.crossOrigin = 'anonymous'
  document.head.appendChild(link)
}

/**
 * DNS 预解析
 * @param url 目标 URL
 */
export function dnsPrefetch(url: string): void {
  if (typeof document === 'undefined') return

  const link = document.createElement('link')
  link.rel = 'dns-prefetch'
  link.href = url
  document.head.appendChild(link)
}

/**
 * 预获取（Prefetch）资源
 * @param href 资源地址
 * @param type 资源类型
 */
export function prefetchResource(href: string, type: 'script' | 'style' | 'image' | 'font' | 'fetch' = 'fetch'): void {
  if (typeof document === 'undefined') return

  const link = document.createElement('link')
  link.rel = 'prefetch'
  link.href = href
  link.as = type
  document.head.appendChild(link)
}

/**
 * 预加载关键资源 Composable
 */
export function usePreload() {
  /**
   * 预加载关键脚本
   */
  const preloadScripts = (scripts: string[]) => {
    scripts.forEach(script => {
      preloadResource({
        type: 'script',
        href: script,
        priority: 'high'
      }).catch(err => {
        console.warn('Failed to preload script:', script, err)
      })
    })
  }

  /**
   * 预加载关键样式
   */
  const preloadStyles = (styles: string[]) => {
    styles.forEach(style => {
      preloadResource({
        type: 'style',
        href: style,
        priority: 'high'
      }).catch(err => {
        console.warn('Failed to preload style:', style, err)
      })
    })
  }

  /**
   * 预加载关键图片
   */
  const preloadImages = (images: string[]) => {
    images.forEach(image => {
      preloadResource({
        type: 'image',
        href: image,
        priority: 'high'
      }).catch(err => {
        console.warn('Failed to preload image:', image, err)
      })
    })
  }

  /**
   * 预连接 API 服务器
   */
  const preconnectAPI = (apiUrl: string) => {
    try {
      const url = new URL(apiUrl)
      preconnect(url.origin)
      dnsPrefetch(url.origin)
    } catch (err) {
      console.warn('Failed to preconnect API:', apiUrl, err)
    }
  }

  return {
    preloadResource,
    preconnect,
    dnsPrefetch,
    prefetchResource,
    preloadScripts,
    preloadStyles,
    preloadImages,
    preconnectAPI
  }
}
