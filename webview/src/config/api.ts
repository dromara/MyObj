// API配置文件
const isDevelopment = import.meta.env.MODE === 'development'

// 获取API基础URL
export const getBaseURL = (): string => {
  if (isDevelopment) {
    // 开发环境默认使用localhost:8080
    return 'http://localhost:8080'
  } else {
    // 生产环境从HTTP头读取配置
    const metaApiUrl = document.querySelector('meta[name="api-url"]')?.getAttribute('content')
    return metaApiUrl || window.location.origin
  }
}

// API版本
export const API_VERSION = '/api'

// 完整的API基础路径
export const API_BASE_URL = getBaseURL() + API_VERSION

// API端点
export const API_ENDPOINTS = {
  // 认证相关
  AUTH: {
    LOGIN: '/user/login',
    REGISTER: '/user/register',
    LOGOUT: '/user/logout',
    REFRESH: '/user/refresh',
    CHALLENGE: '/user/challenge',
  },
  
  // 用户相关
  USER: {
    INFO: '/user/info',
    UPDATE: '/user/update',
    CHANGE_PASSWORD: '/user/updatePassword',
    STORAGE: '/user/storage',
    SET_FILE_PASSWORD: '/user/setFilePassword',
    UPDATE_FILE_PASSWORD: '/user/updateFilePassword',
    SYS_INFO: '/user/sysInfo',
  },
  
  // 文件相关
  FILE: {
    LIST: '/file/list',
    UPLOAD: '/file/upload',
    DOWNLOAD: '/file/download',
    DELETE: '/file/delete',
    RENAME: '/file/rename',
    MOVE: '/file/move',
    COPY: '/file/copy',
    INFO: '/file/info',
    PREVIEW: '/file/preview',
    THUMBNAIL: '/file/thumbnail',
    PRECHECK: '/file/upload/precheck',
    // 分片上传
    CHUNK_UPLOAD: '/file/chunk/upload',
    CHUNK_MERGE: '/file/chunk/merge',
    CHUNK_CHECK: '/file/chunk/check',
    // 搜索
    SEARCH_USER: '/file/search/user',
    SEARCH_PUBLIC: '/file/search/public',
  },
  
  // 文件夹相关
  FOLDER: {
    CREATE: '/file/makeDir',
    LIST: '/folder/list',
    DELETE: '/folder/delete',
    RENAME: '/folder/rename',
  },
  
  // 分享相关
  SHARE: {
    CREATE: '/share/create',
    LIST: '/share/list',
    DELETE: '/share/delete',
    UPDATE_PASSWORD: '/share/updatePassword',
    ACCESS: '/share/access',
    INFO: '/share/info',
  },
  
  // 离线下载
  DOWNLOAD: {
    CREATE_OFFLINE: '/download/offline/create',
    CREATE_TORRENT: '/download/torrent/create',
    LIST: '/download/list',
    CANCEL: '/download/cancel',
    DELETE: '/download/delete',
    PAUSE: '/download/pause',
    RESUME: '/download/resume',
    LOCAL_CREATE: '/download/local/create',
    LOCAL_FILE: '/download/local/file',
  },
  
  // 上传/下载任务
  TASK: {
    UPLOAD_LIST: '/task/upload/list',
    DOWNLOAD_LIST: '/task/download/list',
    CANCEL: '/task/cancel',
  },
  
  // 回收站
  RECYCLED: {
    LIST: '/recycled/list',
    RESTORE: '/recycled/restore',
    DELETE: '/recycled/delete',
    EMPTY: '/recycled/empty',
  },
}

// 请求超时配置
export const TIMEOUT = {
  DEFAULT: 10000,      // 默认10秒
  UPLOAD: 300000,      // 上传5分钟
  DOWNLOAD: 300000,    // 下载5分钟
}

// 文件上传配置
export const UPLOAD_CONFIG = {
  CHUNK_SIZE: 1024 * 1024 * 5, // 5MB分片大小
  MAX_FILE_SIZE: 1024 * 1024 * 1024 * 10, // 10GB最大文件
  CONCURRENT_CHUNKS: 3, // 并发上传3个分片
}
