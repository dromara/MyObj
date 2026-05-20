import axios, { type AxiosRequestConfig, type AxiosResponse, type InternalAxiosRequestConfig } from 'axios'
import { API_BASE_URL } from '@myobj/shared'
import { cache } from '@myobj/shared'

// 通用响应结构
interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

// 创建 axios 实例
const service = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器 — 自动注入 Token
service.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = cache.local.get('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器 — 统一错误处理
service.interceptors.response.use(
  (response: AxiosResponse<ApiResponse>) => {
    const data = response.data

    // 响应体为空（如 204 No Content 或空响应）
    if (data == null) {
      return { code: response.status, message: '', data: null } as any
    }

    // 业务状态码判断：401/403 拦截，其余放行由业务层处理
    if (data.code === 401) {
      cache.local.remove('token')
      window.location.href = '/login'
      return Promise.reject(new Error(data.message || '登录已过期，请重新登录'))
    }
    if (data.code === 403) {
      return Promise.reject(new Error(data.message || '权限不足'))
    }

    return data as any
  },
  (error) => {
    if (error.response) {
      const status = error.response.status
      const data = error.response.data

      if (status === 401) {
        cache.local.remove('token')
        window.location.href = '/login'
        return Promise.reject(new Error(data?.message || '登录已过期，请重新登录'))
      }
      if (status === 403) {
        return Promise.reject(new Error(data?.message || '权限不足'))
      }
      return Promise.reject(new Error(data?.message || '请求失败'))
    }

    if (error.code === 'ECONNABORTED') {
      return Promise.reject(new Error('请求超时'))
    }

    if (axios.isCancel(error)) {
      return Promise.reject(new Error('请求已取消'))
    }

    return Promise.reject(new Error('网络错误'))
  }
)

// GET 请求
export const get = <T = any>(
  url: string,
  params: Record<string, any> = {},
  options: AxiosRequestConfig = {}
): Promise<T> => {
  return service.get(url, { params, ...options }) as any
}

// POST 请求
export const post = <T = any>(
  url: string,
  data: any = {},
  options: AxiosRequestConfig = {}
): Promise<T> => {
  return service.post(url, data, options) as any
}

// PUT 请求
export const put = <T = any>(
  url: string,
  data: any = {},
  options: AxiosRequestConfig = {}
): Promise<T> => {
  return service.put(url, data, options) as any
}

// DELETE 请求
export const del = <T = any>(
  url: string,
  options: AxiosRequestConfig = {}
): Promise<T> => {
  return service.delete(url, options) as any
}

// 文件上传（支持进度回调和取消）
export const upload = <T = any>(
  url: string,
  file: File,
  formData: FormData,
  onProgress?: (percent: number, loaded?: number, total?: number) => void,
  options: { onCancel?: (cancel: () => void) => void } = {}
): Promise<T> => {
  formData.append('file', file)

  const source = axios.CancelToken.source()

  if (options.onCancel) {
    options.onCancel(() => source.cancel('上传已取消'))
  }

  return service.post(url, formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    cancelToken: source.token,
    timeout: 0, // 上传不限时
    onUploadProgress: (e) => {
      if (onProgress && e.total) {
        const percent = (e.loaded / e.total) * 100
        onProgress(percent, e.loaded, e.total)
      }
    }
  }) as any
}

// 文件下载
export const download = async (url: string, filename: string): Promise<void> => {
  const response = await service.get(url, {
    responseType: 'blob',
    timeout: 0
  })

  const blob = new Blob([response as any])
  const downloadUrl = window.URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = downloadUrl
  a.download = filename
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  window.URL.revokeObjectURL(downloadUrl)
}

export default {
  get,
  post,
  put,
  del,
  upload,
  download
}
