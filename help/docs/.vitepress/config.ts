import { defineConfig } from 'vitepress'

const base = '/help/'

export default defineConfig({
  title: 'MyObj 帮助文档',
  description: 'MyObj 私有云存储系统使用指南',
  base: base,
  
  // Head 配置 - 设置 favicon
  // 注意：public 目录下的文件会被原样复制到输出目录根目录
  // 使用 base 路径确保在子路径部署时正确加载
  head: [
    // SVG favicon（现代浏览器支持，可缩放，优先使用）
    ['link', { rel: 'icon', type: 'image/svg+xml', href: `${base}favicon.svg` }],
    // 标准 favicon 回退（使用现有的 LOGO.png）
    ['link', { rel: 'icon', type: 'image/png', href: `${base}LOGO.png` }],
    // Apple Touch Icon（iOS 设备）
    ['link', { rel: 'apple-touch-icon', href: `${base}LOGO.png` }],
    // 网站清单
    ['meta', { name: 'theme-color', content: '#1e1e1e' }]
  ],
  
  // 主题配置
  themeConfig: {
    // Logo 配置（public 目录下的文件，使用 base 路径）
    logo: `${base}LOGO.png`,
    // 网站标题
    siteTitle: 'MyObj 帮助',
    
    // 导航栏
    nav: [
      { text: '首页', link: '/' },
      { text: '快速开始', link: '/guide/getting-started' },
      { text: '功能指南', link: '/guide/features' },
      { text: '协议支持', items: [
        { text: 'WebDAV', link: '/guide/webdav' },
        { text: 'S3 协议', link: '/guide/s3' }
      ]},
      { text: 'API 文档', link: '/api/overview' },
      { text: '常见问题', link: '/faq' }
    ],
    
    // 侧边栏
    sidebar: {
      '/guide/': [
        {
          text: '入门指南',
          items: [
            { text: '快速开始', link: '/guide/getting-started' },
            { text: '安装部署', link: '/guide/installation' },
            { text: '配置说明', link: '/guide/configuration' }
          ]
        },
        {
          text: '核心功能',
          items: [
            { text: '功能概览', link: '/guide/features' },
            { text: '完整功能清单', link: '/guide/features-complete' },
            { text: '文件管理', link: '/guide/file-management' },
            { text: '文件预览', link: '/guide/file-preview' },
            { text: '文件加密', link: '/guide/file-encryption' },
            { text: '文件分享', link: '/guide/file-sharing' },
            { text: '文件广场', link: '/guide/file-square' },
            { text: 'cli工具', link: '/guide/cli' }
          ]
        },
        {
          text: '下载功能',
          items: [
            { text: '离线下载', link: '/guide/offline-download' },
            { text: '任务中心', link: '/guide/tasks' }
          ]
        },
        {
          text: '协议支持',
          items: [
            { text: 'WebDAV 使用', link: '/guide/webdav' },
            { text: 'S3 协议使用', link: '/guide/s3' },
            { text: '在RuoYi-Plus 框架里集成', link: '/guide/integration-ruoyi' }
          ]
        },
        {
          text: '系统管理',
          items: [
            { text: '用户权限', link: '/guide/user-permissions' },
            { text: '系统管理', link: '/guide/system-admin' },
            { text: '回收站', link: '/guide/recycle-bin' }
          ]
        }
      ],
      '/api/': [
        {
          text: 'REST API',
          items: [
            { text: 'API 概览', link: '/api/overview' },
            { text: '认证授权', link: '/api/authentication' },
            { text: '文件操作', link: '/api/files' },
            { text: '用户管理', link: '/api/users' },
            { text: '分享功能', link: '/api/shares' }
          ]
        },
        {
          text: 'S3 API',
          items: [
            { text: 'S3 API 文档', link: '/api/s3' }
          ]
        }
      ]
    },
    
    // 社交链接
    socialLinks: [
      { icon: 'github', link: 'https://github.com/dromara/MyObj' },
      { icon: 'gitee', link: 'https://gitee.com/dromara/my-obj' }
    ],
    
    // 页脚
    footer: {
      message: '基于 Apache 2.0 许可证发布',
      copyright: 'Copyright © 2024 MyObj Team'
    },
    
    // 搜索
    search: {
      provider: 'local'
    },
    
    // 编辑链接
    editLink: {
      pattern: 'https://gitee.com/dromara/my-obj/edit/master/help/docs/:path',
      text: '在 Gitee 上编辑此页'
    },
    
    // 最后更新时间
    lastUpdated: {
      text: '最后更新于',
      formatOptions: {
        dateStyle: 'short',
        timeStyle: 'medium'
      }
    }
  }
})
