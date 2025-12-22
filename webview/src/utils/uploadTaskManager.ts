/**
 * 上传任务管理器
 * 用于管理前端的上传任务状态
 */

import cache from '@/plugins/cache'

export interface UploadTask {
  id: string
  file_name: string
  file_size: number
  uploaded_size: number
  progress: number
  status: 'pending' | 'uploading' | 'paused' | 'completed' | 'failed' | 'cancelled'
  speed: string
  created_at: string
  error?: string
  // 速度计算相关
  lastUpdateTime?: number
  lastUploadedSize?: number
  // 速度平滑处理（移动平均）
  speedHistory?: number[] // 存储最近的速度值（字节/秒）
  lastSpeedUpdateTime?: number // 上次更新速度的时间
}

class UploadTaskManager {
  private tasks: Map<string, UploadTask> = new Map()
  private listeners: Set<(tasks: UploadTask[]) => void> = new Set()
  private readonly STORAGE_KEY = 'upload_tasks'
  
  constructor() {
    // 初始化时从 localStorage 恢复任务
    this.loadTasksFromStorage()
  }

  /**
   * 创建上传任务
   */
  createTask(fileName: string, fileSize: number): string {
    const taskId = `upload_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
    const now = Date.now()
    const task: UploadTask = {
      id: taskId,
      file_name: fileName,
      file_size: fileSize,
      uploaded_size: 0,
      progress: 0,
      status: 'pending',
      speed: '0 KB/s',
      created_at: new Date().toISOString(),
      lastUpdateTime: now,
      lastUploadedSize: 0,
      speedHistory: [],
      lastSpeedUpdateTime: undefined
    }
    
    this.tasks.set(taskId, task)
    this.notifyListeners()
    return taskId
  }

  /**
   * 更新任务进度（自动计算速度）
   */
  updateProgress(taskId: string, progress: number, uploadedSize: number) {
    const task = this.tasks.get(taskId)
    if (task && task.status !== 'paused' && task.status !== 'cancelled') {
      const now = Date.now()
      task.progress = progress
      task.uploaded_size = uploadedSize
      task.status = 'uploading'
      
      // 初始化速度历史记录
      if (!task.speedHistory) {
        task.speedHistory = []
      }
      
      // 计算速度（每秒传输的字节数）
      // 限制更新频率：至少间隔 500ms 才更新一次速度
      const shouldUpdateSpeed = !task.lastSpeedUpdateTime || (now - task.lastSpeedUpdateTime) >= 500
      
      if (task.lastUpdateTime && task.lastUploadedSize !== undefined && shouldUpdateSpeed) {
        const timeDiff = (now - task.lastUpdateTime) / 1000 // 秒
        const sizeDiff = uploadedSize - task.lastUploadedSize // 字节
        
        // 确保上传大小只能递增，不能回退（避免负数速度）
        if (timeDiff > 0 && sizeDiff >= 0) {
          const currentSpeedBytes = sizeDiff / timeDiff // 字节/秒
          
          // 只记录正数速度值（过滤异常值）
          if (currentSpeedBytes >= 0) {
            // 使用移动平均平滑速度（保留最近10个值）
            task.speedHistory.push(currentSpeedBytes)
            if (task.speedHistory.length > 10) {
              task.speedHistory.shift() // 移除最旧的值
            }
            
            // 计算平均值（只计算正数速度）
            const validSpeeds = task.speedHistory.filter(speed => speed >= 0)
            if (validSpeeds.length > 0) {
              const avgSpeed = validSpeeds.reduce((sum, speed) => sum + speed, 0) / validSpeeds.length
              task.speed = this.formatSpeed(avgSpeed)
              task.lastSpeedUpdateTime = now
            }
          }
        }
      }
      
      // 确保 uploadedSize 只能递增，不能回退
      if (task.lastUploadedSize !== undefined && uploadedSize < task.lastUploadedSize) {
        // 如果新值小于旧值，说明可能是并发更新导致的，保持旧值
        return
      }
      
      task.lastUpdateTime = now
      task.lastUploadedSize = uploadedSize
      this.notifyListeners()
    }
  }

  /**
   * 格式化速度（支持 B/s, KB/s, MB/s, GB/s）
   */
  private formatSpeed(bytesPerSecond: number): string {
    if (bytesPerSecond < 1024) {
      return `${Math.round(bytesPerSecond)} B/s`
    } else if (bytesPerSecond < 1024 * 1024) {
      return `${(bytesPerSecond / 1024).toFixed(2)} KB/s`
    } else if (bytesPerSecond < 1024 * 1024 * 1024) {
      return `${(bytesPerSecond / (1024 * 1024)).toFixed(2)} MB/s`
    } else {
      return `${(bytesPerSecond / (1024 * 1024 * 1024)).toFixed(2)} GB/s`
    }
  }

  /**
   * 完成任务
   */
  completeTask(taskId: string) {
    const task = this.tasks.get(taskId)
    if (task) {
      task.status = 'completed'
      task.progress = 100
      task.uploaded_size = task.file_size
      this.notifyListeners()
    }
  }

  /**
   * 任务失败
   */
  failTask(taskId: string, error: string) {
    const task = this.tasks.get(taskId)
    if (task) {
      task.status = 'failed'
      task.error = error
      this.notifyListeners()
    }
  }

  /**
   * 暂停任务
   */
  pauseTask(taskId: string) {
    const task = this.tasks.get(taskId)
    if (task && (task.status === 'uploading' || task.status === 'pending')) {
      task.status = 'paused'
      task.speed = '0 KB/s'
      this.notifyListeners()
    }
  }

  /**
   * 继续任务
   */
  resumeTask(taskId: string) {
    const task = this.tasks.get(taskId)
    if (task && task.status === 'paused') {
      task.status = 'uploading'
      task.lastUpdateTime = Date.now()
      task.lastUploadedSize = task.uploaded_size
      this.notifyListeners()
    }
  }

  /**
   * 取消任务
   */
  cancelTask(taskId: string) {
    const task = this.tasks.get(taskId)
    if (task) {
      task.status = 'cancelled'
      task.speed = '0 KB/s'
      this.notifyListeners()
    }
  }

  /**
   * 删除任务
   */
  deleteTask(taskId: string) {
    this.tasks.delete(taskId)
    this.notifyListeners()
  }

  /**
   * 获取所有任务
   */
  getAllTasks(): UploadTask[] {
    return Array.from(this.tasks.values()).sort((a, b) => 
      new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
    )
  }

  /**
   * 获取任务
   */
  getTask(taskId: string): UploadTask | undefined {
    return this.tasks.get(taskId)
  }

  /**
   * 订阅任务变化
   */
  subscribe(listener: (tasks: UploadTask[]) => void) {
    this.listeners.add(listener)
    // 立即通知一次
    listener(this.getAllTasks())
    
    // 返回取消订阅函数
    return () => {
      this.listeners.delete(listener)
    }
  }

  /**
   * 通知所有监听者
   */
  private notifyListeners() {
    const tasks = this.getAllTasks()
    // 同步保存到 localStorage
    this.saveTasksToStorage()
    this.listeners.forEach(listener => {
      try {
        listener(tasks)
      } catch (error) {
        console.error('上传任务监听器错误:', error)
      }
    })
  }

  /**
   * 保存任务到 localStorage
   */
  private saveTasksToStorage() {
    try {
      const tasks = Array.from(this.tasks.values())
      // 保存所有任务（包括正在上传的，刷新后会标记为失败）
      cache.local.setJSON(this.STORAGE_KEY, tasks)
    } catch (error) {
      console.error('保存上传任务到 localStorage 失败:', error)
    }
  }

  /**
   * 从 localStorage 加载任务
   */
  private loadTasksFromStorage() {
    try {
      const tasks: UploadTask[] | null = cache.local.getJSON(this.STORAGE_KEY)
      if (tasks && Array.isArray(tasks)) {
        tasks.forEach(task => {
          // 验证任务状态是否有效
          const validStatuses: UploadTask['status'][] = ['pending', 'uploading', 'paused', 'completed', 'failed', 'cancelled']
          if (!validStatuses.includes(task.status)) {
            // 如果状态无效，默认为失败
            task.status = 'failed'
            task.error = task.error || '任务状态异常'
          }
          
          // 如果任务还在上传中或暂停，标记为失败（因为刷新后无法继续）
          // 注意：已完成、已失败、已取消的任务保持不变
          if (task.status === 'uploading' || task.status === 'pending' || task.status === 'paused') {
            task.status = 'failed'
            task.error = task.error || '页面刷新后无法恢复上传'
            task.speed = '0 KB/s'
          }
          
          // 清理临时数据（所有任务都需要清理）
          task.lastUpdateTime = undefined
          task.lastUploadedSize = undefined
          task.speedHistory = []
          task.lastSpeedUpdateTime = undefined
          
          // 确保已完成的任务进度为100%
          if (task.status === 'completed') {
            task.progress = 100
            task.uploaded_size = task.file_size
            task.speed = '0 KB/s'
          }
          
          this.tasks.set(task.id, task)
        })
        
        // 加载完成后保存一次，确保状态更新被持久化
        this.saveTasksToStorage()
      }
    } catch (error) {
      console.error('从 localStorage 加载上传任务失败:', error)
    }
  }

  /**
   * 清理已完成的任务（可选，保留最近的任务）
   */
  cleanup(keepCount: number = 50) {
    const allTasks = this.getAllTasks()
    if (allTasks.length > keepCount) {
      const tasksToDelete = allTasks.slice(keepCount)
      tasksToDelete.forEach(task => {
        if (task.status === 'completed' || task.status === 'failed' || task.status === 'cancelled') {
          this.tasks.delete(task.id)
        }
      })
      this.notifyListeners()
    }
  }
}

// 导出单例
export const uploadTaskManager = new UploadTaskManager()

