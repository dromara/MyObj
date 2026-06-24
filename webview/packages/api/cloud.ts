import { get, post } from '@myobj/http'

// 云盘提供者类型
export type CloudProvider = 'aliyun' | 'baidu' | 'xunlei' | 'quark' | '115' | 'tianyi' | 'uc' | 'pikpak' | 'caiyun' | 'wopan'

// 云盘提供者信息
export const providerInfo: Record<CloudProvider, { name: string; icon: string; color: string }> = {
  aliyun: { name: '阿里云盘', icon: '🅰️', color: '#409EFF' },
  baidu: { name: '百度网盘', icon: '🅱️', color: '#306CFF' },
  xunlei: { name: '迅雷网盘', icon: '⚡', color: '#00BE06' },
  quark: { name: '夸克网盘', icon: '🔮', color: '#6A5ACD' },
  '115': { name: '115网盘', icon: '💾', color: '#2196F3' },
  tianyi: { name: '天翼云盘', icon: '☁️', color: '#FF6B00' },
  uc: { name: 'UC网盘', icon: '🌐', color: '#FF6B00' },
  pikpak: { name: 'PikPak', icon: '📦', color: '#7C3AED' },
  caiyun: { name: '和彩云', icon: '🌤️', color: '#00BCD4' },
  wopan: { name: '联通云盘', icon: '🔗', color: '#E91E63' }
}

// 自动检测云盘类型
export const detectProvider = (url: string): CloudProvider | null => {
  const lowerUrl = url.toLowerCase()
  if (lowerUrl.includes('aliyundrive.com') || lowerUrl.includes('alipan.com')) return 'aliyun'
  if (lowerUrl.includes('pan.baidu.com') || lowerUrl.includes('yun.baidu.com')) return 'baidu'
  if (lowerUrl.includes('pan.xunlei.com') || lowerUrl.includes('xunlei.com')) return 'xunlei'
  if (lowerUrl.includes('pan.quark.cn') || lowerUrl.includes('quark.cn')) return 'quark'
  if (lowerUrl.includes('115.com') || lowerUrl.includes('115cdn.com')) return '115'
  if (lowerUrl.includes('cloud.189.cn') || lowerUrl.includes('tianyi.com')) return 'tianyi'
  if (lowerUrl.includes('drive.uc.cn') || lowerUrl.includes('pan.uc.cn')) return 'uc'
  if (lowerUrl.includes('mypikpak.com') || lowerUrl.includes('pikpak.com')) return 'pikpak'
  if (lowerUrl.includes('caiyun.139.com') || lowerUrl.includes('caiyun.com')) return 'caiyun'
  if (lowerUrl.includes('pan.wo.cn') || lowerUrl.includes('wopan.cn')) return 'wopan'
  return null
}

// 获取云盘提供者名称
export const getProviderName = (provider: string): string => {
  return providerInfo[provider as CloudProvider]?.name || provider
}

// 格式化文件大小
export const formatSize = (bytes: number): string => {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

// 获取支持的云盘列表
export const getSupportedProviders = () => get('/cloud/providers')

// 解析分享链接
export const parseShareLink = (data: {
  provider?: string
  share_url: string
  share_pwd?: string
  target_path?: string
}) => post('/cloud/parse', data)

// 获取任务状态
export const getTaskStatus = (taskId: number) => get(`/cloud/task/${taskId}`)

// 获取任务列表
export const getTaskList = (page = 1, pageSize = 20) =>
  get('/cloud/tasks', { page, page_size: pageSize })

// 保存分享文件到本地（转存）
export const saveShareFiles = (data: {
  provider: string
  share_id: string
  save_type: 'single' | 'multiple' | 'all' | 'directory'
  file_ids?: string[]
  dir_name?: string
  target_path?: string
}) => post('/cloud/save', data)

// 获取文件分类统计
export const getCategoryStats = () => get('/file/categories/stats')

// 获取文件分类列表
export const getCategories = () => get('/file/categories')

// 获取缩略图URL
export const getThumbnailUrl = (fileId: string) => `/api/file/thumbnail/${fileId}`

// 手动生成缩略图
export const generateThumbnail = (fileId: string) =>
  post(`/file/thumbnail/generate/${fileId}`)
