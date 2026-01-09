import cache from '@/plugins/cache'
import logger from '@/plugins/logger'
import { formatSpeed as formatSpeedUtil } from '@/utils/format'

export type UploadStage = 'reading' | 'calculating' | 'uploading' | 'completed' | 'failed'

export interface UploadTask {
  id: string
  file_name: string
  file_size: number
  uploaded_size: number
  progress: number
  status: 'pending' | 'uploading' | 'paused' | 'completed' | 'failed' | 'cancelled'
  stage?: UploadStage // 处理阶段：reading(读取文件) | calculating(计算MD5) | uploading(上传中) | completed(完成) | failed(失败)
  speed: string
  created_at: string
  error?: string
  lastUpdateTime?: number
  lastUploadedSize?: number
  speedHistory?: number[]
  lastSpeedUpdateTime?: number
  pathId?: string
  precheckId?: string
  chunkSignature?: string
  filesMd5?: string[]
  uploadedChunkMd5s?: string[]
}

class UploadTaskManager {
  private tasks: Map<string, UploadTask> = new Map()
  private listeners: Set<(tasks: UploadTask[]) => void> = new Set()
  private readonly STORAGE_KEY = 'upload_tasks'
  private readonly DELETED_TASKS_KEY = 'deleted_upload_tasks'
  private deletedPrecheckIds: Set<string> = new Set()
  
  constructor() {
    this.loadTasksFromStorage()
    this.loadDeletedPrecheckIds()
  }

  /**
   * 创建上传任务
   * @param fileName 文件名
   * @param fileSize 文件大小（字节）
   * @returns string 任务ID
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
      stage: 'reading', // 初始阶段：读取文件
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
   * 更新任务阶段
   * @param taskId 任务ID
   * @param stage 处理阶段
   */
  updateStage(taskId: string, stage: UploadStage) {
    const task = this.tasks.get(taskId)
    if (task) {
      task.stage = stage
      this.notifyListeners()
    }
  }

  /**
   * 更新任务进度（自动计算速度）
   * @param taskId 任务ID
   * @param progress 进度百分比（0-100）
   * @param uploadedSize 已上传大小（字节）
   */
  updateProgress(taskId: string, progress: number, uploadedSize: number) {
    const task = this.tasks.get(taskId)
    if (task && task.status !== 'paused' && task.status !== 'cancelled') {
      const now = Date.now()
      task.progress = progress
      task.uploaded_size = uploadedSize
      task.status = 'uploading'
      
      if (!task.speedHistory) {
        task.speedHistory = []
      }
      
      const shouldUpdateSpeed = !task.lastSpeedUpdateTime || (now - task.lastSpeedUpdateTime) >= 500
      
      if (task.lastUpdateTime && task.lastUploadedSize !== undefined && shouldUpdateSpeed) {
        const timeDiff = (now - task.lastUpdateTime) / 1000
        const sizeDiff = uploadedSize - task.lastUploadedSize
        
        if (timeDiff > 0 && sizeDiff >= 0) {
          const currentSpeedBytes = sizeDiff / timeDiff
          
          if (currentSpeedBytes >= 0) {
            task.speedHistory.push(currentSpeedBytes)
            if (task.speedHistory.length > 10) {
              task.speedHistory.shift()
            }
            
            const validSpeeds = task.speedHistory.filter(speed => speed >= 0)
            if (validSpeeds.length > 0) {
              const avgSpeed = validSpeeds.reduce((sum, speed) => sum + speed, 0) / validSpeeds.length
              task.speed = formatSpeedUtil(avgSpeed)
              task.lastSpeedUpdateTime = now
            }
          }
        }
      }
      
      if (task.lastUploadedSize !== undefined && uploadedSize < task.lastUploadedSize) {
        return
      }
      
      if (uploadedSize > task.file_size) {
        logger.warn(`上传大小超过文件总大小，已限制为文件大小: ${uploadedSize} > ${task.file_size}`)
        uploadedSize = task.file_size
      }
      
      task.lastUpdateTime = now
      task.lastUploadedSize = uploadedSize
      task.uploaded_size = uploadedSize
      this.notifyListeners()
    }
  }


  /**
   * 标记任务为完成状态
   * @param taskId 任务ID
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
   * 标记任务为失败状态
   * @param taskId 任务ID
   * @param error 错误信息
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
   * @param taskId 任务ID
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
   * 设置任务的取消函数（用于暂停时取消上传）
   * @param taskId 任务ID
   * @param cancelFn 取消函数
   */
  setCancelFunction(taskId: string, cancelFn: () => void) {
    const task = this.tasks.get(taskId)
    if (task) {
      ;(task as any).cancelFunction = cancelFn
    }
  }
  
  /**
   * 获取任务的取消函数
   * @param taskId 任务ID
   * @returns 取消函数，如果不存在则返回null
   */
  getCancelFunction(taskId: string): (() => void) | null {
    const task = this.tasks.get(taskId)
    return task ? ((task as any).cancelFunction || null) : null
  }
  
  /**
   * 取消所有正在进行的上传请求（用于暂停时）
   * @param taskId 任务ID
   */
  cancelAllUploads(taskId: string) {
    const task = this.tasks.get(taskId)
    if (task && (task as any).cancelFunctions) {
      const cancelFunctions = (task as any).cancelFunctions as (() => void)[]
      cancelFunctions.forEach(cancel => {
        try {
          if (typeof cancel === 'function') {
            cancel()
          }
        } catch (error) {
          // 静默处理
        }
      })
      cancelFunctions.length = 0
    }
  }

  /**
   * 恢复任务（从暂停状态恢复为上传中）
   * @param taskId 任务ID
   * @returns 任务对象，如果任务不存在或状态不正确则返回null
   */
  resumeTask(taskId: string): UploadTask | null {
    const task = this.tasks.get(taskId)
    if (task && task.status === 'paused') {
      task.status = 'uploading'
      task.lastUpdateTime = Date.now()
      task.lastUploadedSize = task.uploaded_size
      this.notifyListeners()
      return task
    }
    return null
  }

  /**
   * 取消任务
   * @param taskId 任务ID
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
   * 更新任务属性
   * @param taskId 任务ID
   * @param updates 要更新的属性
   */
  updateTask(taskId: string, updates: Partial<UploadTask>) {
    const task = this.tasks.get(taskId)
    if (task) {
      Object.assign(task, updates)
      this.saveTasksToStorage()
      this.notifyListeners()
    }
  }

  /**
   * 删除任务
   * 如果任务有precheckId，会将其添加到已删除列表，防止同步时恢复
   * @param taskId 任务ID
   */
  deleteTask(taskId: string) {
    const task = this.tasks.get(taskId)
    if (task) {
      if (task.precheckId) {
        this.deletedPrecheckIds.add(task.precheckId)
        this.saveDeletedPrecheckIds()
      }
      this.tasks.delete(taskId)
      this.saveTasksToStorage()
      this.notifyListeners()
    }
  }

  /**
   * 检查precheckId是否在已删除列表中
   * @param precheckId 预检ID
   * @returns boolean 如果已删除返回true，否则返回false
   */
  isPrecheckIdDeleted(precheckId: string): boolean {
    return this.deletedPrecheckIds.has(precheckId)
  }

  private loadDeletedPrecheckIds() {
    try {
      const deletedIds: string[] | null = cache.local.getJSON(this.DELETED_TASKS_KEY)
      if (deletedIds && Array.isArray(deletedIds)) {
        this.deletedPrecheckIds = new Set(deletedIds)
      }
    } catch (error) {
      logger.error('加载已删除任务列表失败:', error)
    }
  }

  private saveDeletedPrecheckIds() {
    try {
      const deletedIds = Array.from(this.deletedPrecheckIds)
      cache.local.setJSON(this.DELETED_TASKS_KEY, deletedIds)
    } catch (error) {
      logger.error('保存已删除任务列表失败:', error)
    }
  }

  /**
   * 清除已删除任务列表（用于恢复被误删的任务）
   */
  clearDeletedPrecheckIds() {
    this.deletedPrecheckIds.clear()
    this.saveDeletedPrecheckIds()
  }

  /**
   * 从已删除列表中移除指定的precheckId（用于恢复单个任务）
   * @param precheckId 预检ID
   * @returns boolean 如果成功移除返回true，否则返回false
   */
  removeFromDeletedPrecheckIds(precheckId: string) {
    if (this.deletedPrecheckIds.has(precheckId)) {
      this.deletedPrecheckIds.delete(precheckId)
      this.saveDeletedPrecheckIds()
      return true
    }
    return false
  }

  /**
   * 获取所有任务（按创建时间倒序排列）
   * @returns UploadTask[] 任务列表
   */
  getAllTasks(): UploadTask[] {
    return Array.from(this.tasks.values()).sort((a, b) => 
      new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
    )
  }

  /**
   * 获取指定任务
   * @param taskId 任务ID
   * @returns UploadTask | undefined 任务对象，如果不存在则返回undefined
   */
  getTask(taskId: string): UploadTask | undefined {
    return this.tasks.get(taskId)
  }

  /**
   * 订阅任务变化
   * @param listener 监听器函数，当任务列表变化时会被调用
   * @returns 取消订阅函数
   */
  subscribe(listener: (tasks: UploadTask[]) => void) {
    this.listeners.add(listener)
    listener(this.getAllTasks())
    
    return () => {
      this.listeners.delete(listener)
    }
  }

  private notifyListeners() {
    const tasks = this.getAllTasks()
    this.saveTasksToStorage()
    this.listeners.forEach(listener => {
      try {
        listener(tasks)
      } catch (error) {
        logger.error('上传任务监听器错误:', error)
      }
    })
  }

  /**
   * 保存任务到localStorage（公开方法，供外部调用）
   */
  saveTasksToStorage() {
    try {
      const tasks = Array.from(this.tasks.values())
      cache.local.setJSON(this.STORAGE_KEY, tasks)
    } catch (error) {
      logger.error('保存上传任务到 localStorage 失败:', error)
    }
  }

  private loadTasksFromStorage() {
    try {
      const tasks: UploadTask[] | null = cache.local.getJSON(this.STORAGE_KEY)
      if (tasks && Array.isArray(tasks)) {
        tasks.forEach(task => {
          const validStatuses: UploadTask['status'][] = ['pending', 'uploading', 'paused', 'completed', 'failed', 'cancelled']
          if (!validStatuses.includes(task.status)) {
            task.status = 'failed'
            task.error = task.error || '任务状态异常'
          }
          
          if (task.status === 'uploading' || task.status === 'pending') {
            task.status = 'paused'
            task.speed = '0 KB/s'
          } else if (task.status === 'paused') {
            task.speed = '0 KB/s'
          }
          
          task.lastUpdateTime = undefined
          task.lastUploadedSize = undefined
          task.speedHistory = []
          task.lastSpeedUpdateTime = undefined
          
          if (task.status === 'completed') {
            task.progress = 100
            task.uploaded_size = task.file_size
            task.speed = '0 KB/s'
          }
          
          this.tasks.set(task.id, task)
        })
        
        this.saveTasksToStorage()
      }
    } catch (error) {
      logger.error('从 localStorage 加载上传任务失败:', error)
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

