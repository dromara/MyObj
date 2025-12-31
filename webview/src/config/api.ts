// API配置文件
// 使用环境变量配置的代理路径（如 /dev-api 或 /prod-api）
// 开发环境：/dev-api -> 代理到后端 /api
// 生产环境：/prod-api -> 代理到后端 /api
export const API_BASE_URL = import.meta.env.VITE_APP_BASE_API || '/dev-api'

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
    UPDATE: '/user/updateUser',
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
    PROGRESS: '/file/upload/progress', // 上传进度查询
    UNCOMPLETED: '/file/upload/uncompleted', // 查询未完成的上传任务列表
    EXPIRED: '/file/upload/expired', // 查询过期的上传任务列表
    DELETE_UPLOAD_TASK: '/file/upload/delete', // 删除上传任务
    RENEW_TASK: '/file/upload/renew', // 延期过期任务（恢复任务）
    CLEAN_EXPIRED: '/file/upload/clean-expired', // 清理过期的上传任务
    // 分片上传
    CHUNK_UPLOAD: '/file/chunk/upload',
    CHUNK_MERGE: '/file/chunk/merge',
    CHUNK_CHECK: '/file/chunk/check',
    // 搜索
    SEARCH_USER: '/file/search/user',
    SEARCH_PUBLIC: '/file/search/public',
    // 公开文件列表
    PUBLIC_LIST: '/file/public/list',
    // 设置文件公开状态
    SET_PUBLIC: '/file/setPublic',
  },
  
  // 文件夹相关
  FOLDER: {
    CREATE: '/file/makeDir',
    LIST: '/folder/list',
    DELETE: '/file/deleteDir', // 目录删除接口
    RENAME: '/file/renameDir', // 目录重命名接口
  },
  
  // 分享相关
  SHARE: {
    CREATE: '/share/create',
    LIST: '/share/list',
    DELETE: '/share/delete',
    UPDATE_PASSWORD: '/share/updatePassword',
    INFO: '/share/info',
    DOWNLOAD: '/share/download',
  },
  
  // 离线下载
  DOWNLOAD: {
    CREATE_OFFLINE: '/download/offline/create',
    TORRENT_PARSE: '/download/torrent/parse',
    TORRENT_START: '/download/torrent/start',
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
  
  // 视频播放
  VIDEO: {
    PRECHECK: '/video/play/precheck',
    STREAM: '/video/stream',
  },
  
  // 文件预览/下载
  PREVIEW: {
    PREVIEW: '/download/preview',
  },
  
  // 管理相关
  ADMIN: {
    USER: {
      LIST: '/admin/user/list',
      CREATE: '/admin/user/create',
      UPDATE: '/admin/user/update',
      DELETE: '/admin/user/delete',
      TOGGLE_STATE: '/admin/user/toggle-state',
    },
    GROUP: {
      LIST: '/admin/group/list',
      CREATE: '/admin/group/create',
      UPDATE: '/admin/group/update',
      DELETE: '/admin/group/delete',
    },
    POWER: {
      LIST: '/admin/power/list',
      CREATE: '/admin/power/create',
      UPDATE: '/admin/power/update',
      DELETE: '/admin/power/delete',
      BATCH_DELETE: '/admin/power/batch-delete',
      ASSIGN: '/admin/power/assign',
      GROUP_POWERS: '/admin/power/group-powers',
    },
    DISK: {
      LIST: '/admin/disk/list',
      CREATE: '/admin/disk/create',
      UPDATE: '/admin/disk/update',
      DELETE: '/admin/disk/delete',
      SCAN: '/admin/disk/scan',
    },
    SYSTEM: {
      CONFIG: '/admin/system/config',
      UPDATE_CONFIG: '/admin/system/update-config',
    },
  },
  
  // 打包下载
  PACKAGE: {
    CREATE: '/file/package/create',
    DOWNLOAD: '/file/package/download',
    PROGRESS: '/file/package/progress',
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
