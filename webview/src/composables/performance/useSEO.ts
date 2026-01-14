/**
 * SEO 管理 Composable
 * 用于动态管理页面的 meta 标签和 SEO 信息
 */

export interface SEOOptions {
  /** 页面标题 */
  title?: string
  /** 页面描述 */
  description?: string
  /** 关键词 */
  keywords?: string
  /** Open Graph 图片 */
  ogImage?: string
  /** Open Graph 类型 */
  ogType?: string
  /** 作者 */
  author?: string
  /** 语言 */
  lang?: string
  /** 是否禁止索引 */
  noindex?: boolean
  /** 是否禁止跟踪 */
  nofollow?: boolean
}

/**
 * SEO 管理 Composable
 * @param options SEO 配置
 */
export function useSEO(options: SEOOptions = {}) {
  const {
    title,
    description,
    keywords,
    ogImage,
    ogType = 'website',
    author,
    lang = 'zh-CN',
    noindex = false,
    nofollow = false
  } = options

  /**
   * 设置或更新 meta 标签
   */
  const setMetaTag = (name: string, content: string, attribute: 'name' | 'property' = 'name') => {
    if (typeof document === 'undefined') return

    let element = document.querySelector(`meta[${attribute}="${name}"]`) as HTMLMetaElement

    if (!element) {
      element = document.createElement('meta')
      element.setAttribute(attribute, name)
      document.head.appendChild(element)
    }

    element.setAttribute('content', content)
  }

  /**
   * 移除 meta 标签
   */
  const removeMetaTag = (name: string, attribute: 'name' | 'property' = 'name') => {
    if (typeof document === 'undefined') return

    const element = document.querySelector(`meta[${attribute}="${name}"]`)
    if (element) {
      element.remove()
    }
  }

  /**
   * 设置页面标题
   */
  const setTitle = (newTitle?: string) => {
    if (typeof document === 'undefined') return

    const finalTitle = newTitle || title || 'MyObj 网盘系统'
    document.title = finalTitle
    setMetaTag('title', finalTitle)
    setMetaTag('og:title', finalTitle, 'property')
  }

  /**
   * 设置页面描述
   */
  const setDescription = (newDescription?: string) => {
    const finalDescription = newDescription || description || 'MyObj 网盘系统 - 安全、高效的文件存储和管理平台'
    setMetaTag('description', finalDescription)
    setMetaTag('og:description', finalDescription, 'property')
  }

  /**
   * 设置关键词
   */
  const setKeywords = (newKeywords?: string) => {
    if (newKeywords || keywords) {
      setMetaTag('keywords', newKeywords || keywords || '网盘,文件存储,文件管理')
    }
  }

  /**
   * 设置 Open Graph 图片
   */
  const setOGImage = (newOgImage?: string) => {
    if (newOgImage || ogImage) {
      setMetaTag('og:image', newOgImage || ogImage || '', 'property')
    }
  }

  /**
   * 设置 Open Graph 类型
   */
  const setOGType = (newOgType?: string) => {
    setMetaTag('og:type', newOgType || ogType, 'property')
  }

  /**
   * 设置作者
   */
  const setAuthor = (newAuthor?: string) => {
    if (newAuthor || author) {
      setMetaTag('author', newAuthor || author || '')
    }
  }

  /**
   * 设置语言
   */
  const setLang = (newLang?: string) => {
    if (typeof document === 'undefined') return

    const finalLang = newLang || lang
    const htmlElement = document.documentElement
    if (htmlElement) {
      htmlElement.setAttribute('lang', finalLang)
    }
  }

  /**
   * 设置 robots meta
   */
  const setRobots = (noIndex?: boolean, noFollow?: boolean) => {
    const finalNoIndex = noIndex !== undefined ? noIndex : noindex
    const finalNoFollow = noFollow !== undefined ? noFollow : nofollow

    if (finalNoIndex || finalNoFollow) {
      const directives: string[] = []
      if (finalNoIndex) directives.push('noindex')
      if (finalNoFollow) directives.push('nofollow')
      setMetaTag('robots', directives.join(', '))
    } else {
      removeMetaTag('robots')
    }
  }

  /**
   * 应用所有 SEO 设置
   */
  const applySEO = (newOptions?: SEOOptions) => {
    const finalOptions = { ...options, ...newOptions }

    if (finalOptions.title) setTitle(finalOptions.title)
    if (finalOptions.description) setDescription(finalOptions.description)
    if (finalOptions.keywords) setKeywords(finalOptions.keywords)
    if (finalOptions.ogImage) setOGImage(finalOptions.ogImage)
    if (finalOptions.ogType) setOGType(finalOptions.ogType)
    if (finalOptions.author) setAuthor(finalOptions.author)
    if (finalOptions.lang) setLang(finalOptions.lang)
    setRobots(finalOptions.noindex, finalOptions.nofollow)
  }

  // 初始化时应用设置
  onMounted(() => {
    applySEO()
  })

  return {
    setTitle,
    setDescription,
    setKeywords,
    setOGImage,
    setOGType,
    setAuthor,
    setLang,
    setRobots,
    applySEO,
    setMetaTag,
    removeMetaTag
  }
}
