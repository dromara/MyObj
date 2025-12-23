// HTTP请求工具
import { API_BASE_URL } from '@/config/api'

interface RequestConfig extends RequestInit {
  params?: Record<string, any>
}

// 请求拦截器 - 添加token
const requestInterceptor = (config: RequestConfig): RequestConfig => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers = {
      ...config.headers,
      'Authorization': `Bearer ${token}`
    }
  }
  return config
}

// 响应拦截器 - 处理错误
const responseInterceptor = async <T = any>(response: Response): Promise<T> => {
  let data: any
  try {
    data = await response.json()
  } catch (e) {
    data = {}
  }
  
  // 先检查 HTTP 状态码
  if (!response.ok) {
    // 处理 HTTP 错误
    if (response.status === 401) {
      // 检查是否是权限不足（通过错误消息判断）
      if (data.message && data.message.includes('用户无权限')) {
        // 权限不足，不跳转登录页
        throw new Error(data.message || '权限不足')
      }
      // token过期，跳转登录
      localStorage.removeItem('token')
      window.location.href = '/login'
      throw new Error(data.message || '登录已过期，请重新登录')
    }
    throw new Error(data.message || '请求失败')
  }
  
  // 检查业务状态码（后端统一返回格式：{code, message, data}）
  if (data.code && data.code !== 200) {
    // 业务错误
    if (data.code === 401) {
      // 区分"用户未登录"和"用户无权限"
      // 如果是权限不足，不跳转登录页，而是抛出错误让调用方处理
      if (data.message && data.message.includes('用户无权限')) {
        // 权限不足，不跳转登录页
        throw new Error(data.message || '权限不足')
      }
      // token无效或过期，跳转登录
      localStorage.removeItem('token')
      window.location.href = '/login'
      throw new Error(data.message || '登录已过期，请重新登录')
    }
    return data
  }
  
  return data
}

// 基础请求方法
const request = async <T = any>(url: string, options: RequestConfig = {}): Promise<T> => {
  const config: RequestConfig = {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
  }
  
  // 应用请求拦截器
  const interceptedConfig = requestInterceptor(config)
  
  try {
    const response = await fetch(API_BASE_URL + url, {
      ...interceptedConfig,
      signal: options.signal,
    })
    
    return await responseInterceptor<T>(response)
  } catch (error: any) {
    if (error.name === 'AbortError') {
      throw new Error('请求已取消')
    }
    throw error
  }
}

// GET请求
export const get = <T = any>(url: string, params: Record<string, any> = {}, options: RequestConfig = {}): Promise<T> => {
  const queryString = new URLSearchParams(params).toString()
  const fullUrl = queryString ? `${url}?${queryString}` : url
  
  return request<T>(fullUrl, {
    method: 'GET',
    ...options,
  })
}

// POST请求
export const post = <T = any>(url: string, data: any = {}, options: RequestConfig = {}): Promise<T> => {
  return request<T>(url, {
    method: 'POST',
    body: JSON.stringify(data),
    ...options,
  })
}

// PUT请求
export const put = <T = any>(url: string, data: any = {}, options: RequestConfig = {}): Promise<T> => {
  return request<T>(url, {
    method: 'PUT',
    body: JSON.stringify(data),
    ...options,
  })
}

// DELETE请求
export const del = <T = any>(url: string, options: RequestConfig = {}): Promise<T> => {
  return request<T>(url, {
    method: 'DELETE',
    ...options,
  })
}

// 文件上传
export const upload = <T = any>(
  url: string,
  file: File,
  outerParams: FormData,
  onProgress?: (percent: number, loaded?: number, total?: number) => void,
  options: { onCancel?: (cancel: () => void) => void } = {}
): Promise<T> => {
  const formData = outerParams;
  formData.append('file', file)
  
  const token = localStorage.getItem('token')
  
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest()
    
    // 上传进度
    if (onProgress) {
      xhr.upload.addEventListener('progress', (e) => {
        if (e.lengthComputable) {
          const percentComplete = (e.loaded / e.total) * 100
          onProgress(percentComplete, e.loaded, e.total)
        }
      })
    }
    
    // 请求完成
    xhr.addEventListener('load', () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        try {
          const response = JSON.parse(xhr.responseText)
          resolve(response)
        } catch (e) {
          resolve(xhr.responseText as any)
        }
      } else {
        try {
          const error = JSON.parse(xhr.responseText)
          reject(new Error(error.message || '上传失败'))
        } catch (e) {
          reject(new Error('上传失败'))
        }
      }
    })
    
    // 请求失败
    xhr.addEventListener('error', () => {
      reject(new Error('网络错误'))
    })
    
    // 请求中止
    xhr.addEventListener('abort', () => {
      reject(new Error('上传已取消'))
    })
    
    xhr.open('POST', API_BASE_URL + url)
    if (token) {
      xhr.setRequestHeader('Authorization', `Bearer ${token}`)
    }
    
    xhr.send(formData)
    
    // 返回取消方法
    if (options.onCancel) {
      options.onCancel(() => xhr.abort())
    }
  })
}

// 文件下载
export const download = async (url: string, filename: string): Promise<void> => {
  const token = localStorage.getItem('token')
  
  try {
    const response = await fetch(API_BASE_URL + url, {
      method: 'GET',
      headers: {
        'Authorization': token ? `Bearer ${token}` : '',
      },
    })
    
    if (!response.ok) {
      throw new Error('下载失败')
    }
    
    const blob = await response.blob()
    const downloadUrl = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = downloadUrl
    a.download = filename
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    window.URL.revokeObjectURL(downloadUrl)
  } catch (error: any) {
    throw new Error(error.message || '下载失败')
  }
}

export default {
  get,
  post,
  put,
  del,
  upload,
  download,
}
