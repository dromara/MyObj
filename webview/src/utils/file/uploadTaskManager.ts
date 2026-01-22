import cache from '@/plugins/cache'
import logger from '@/plugins/logger'
import { formatSpeed as formatSpeedUtil } from '@/utils/format/format'

export interface UploadTask {
  id: string
  file_name: string
  file_size: number
  uploaded_size: number
  progress: number
  status: 'prechecking' | 'pending' | 'uploading' | 'paused' | 'completed' | 'failed' | 'cancelled'
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
  precheckProgress?: number // 预检进度（0-100）
  currentStep?: string // 当前步骤描述
  startTime?: number // 开始上传时间（时间戳）
  endTime?: number // 结束上传时间（时间戳）
  totalDuration?: number // 总耗时（毫秒）
  averageSpeed?: number // 平均速度（字节/秒）
  isInstantUpload?: boolean // 是否秒传
}

class UploadTaskManager {
  private tasks: Map<string, UploadTask> = new Map()
  private listeners: Set<(tasks: UploadTask[]) => void> = new Set()
  private readonly STORAGE_KEY_PREFIX = 'upload_tasks'
  private readonly DELETED_TASKS_KEY_PREFIX = 'deleted_upload_tasks'
  private deletedPrecheckIds: Set<string> = new Set()
  private currentUserId: string | null = null

  constructor() {
    // 延迟加载，等待用户ID确定后再加载
  }

  /**
   * 获取当前用户ID（从缓存中读取）
   */
  private getCurrentUserId(): string | null {
    try {
      // 从 userInfo 键读取（与 user store 保持一致）
      const userInfo = cache.local.getJSON('userInfo')
      if (!userInfo) {
        return null
      }
      return userInfo.id || userInfo.user_id || null
    } catch (error) {
      logger.error('获取用户ID失败:', error)
      return null
    }
  }

  /**
   * 获取存储键（包含用户ID）
   */
  private getStorageKey(): string {
    const userId = this.getCurrentUserId()
    if (!userId) {
      logger.error('无法获取用户ID，无法确定存储键')
      throw new Error('用户未登录，无法访问上传任务')
    }
    return `${this.STORAGE_KEY_PREFIX}_${userId}`
  }

  /**
   * 获取已删除任务列表的存储键（包含用户ID）
   */
  private getDeletedTasksKey(): string {
    const userId = this.getCurrentUserId()
    if (!userId) {
      logger.error('无法获取用户ID，无法确定存储键')
      throw new Error('用户未登录，无法访问上传任务')
    }
    return `${this.DELETED_TASKS_KEY_PREFIX}_${userId}`
  }

  /**
   * 初始化（在用户登录后调用）
   */
  init() {
    const userId = this.getCurrentUserId()
    // 如果用户ID发生变化，清空当前任务并加载新用户的任务
    if (userId !== this.currentUserId) {
      this.tasks.clear()
      this.deletedPrecheckIds.clear()
      this.currentUserId = userId
      this.loadTasksFromStorage()
      this.loadDeletedPrecheckIds()
      this.notifyListeners()
    }
  }

  /**
   * 清空当前用户的所有任务（在用户登出时调用）
   */
  clearCurrentUserTasks() {
    this.tasks.clear()
    this.deletedPrecheckIds.clear()
    this.currentUserId = null
    this.notifyListeners()
  }

  /**
   * 创建上传任务
   * @param fileName 文件名
   * @param fileSize 文件大小（字节）
   * @param initialStatus 初始状态，默认为 'pending'，可以是 'prechecking'（预检中）
   * @returns string 任务ID
   */
  createTask(fileName: string, fileSize: number, initialStatus: UploadTask['status'] = 'pending'): string {
    const taskId = `upload_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
    const now = Date.now()
    const task: UploadTask = {
      id: taskId,
      file_name: fileName,
      file_size: fileSize,
      uploaded_size: 0,
      progress: 0,
      status: initialStatus,
      speed: '0 KB/s',
      created_at: new Date().toISOString(),
      lastUpdateTime: now,
      lastUploadedSize: 0,
      speedHistory: [],
      lastSpeedUpdateTime: undefined,
      precheckProgress: initialStatus === 'prechecking' ? 0 : undefined,
      currentStep: initialStatus === 'prechecking' ? '正在初始化...' : undefined,
      startTime: now, // 记录开始时间
      isInstantUpload: false
    }

    this.tasks.set(taskId, task)
    this.notifyListeners()
    this.saveTasksToStorage()
    return taskId
  }

  /**
   * 更新任务进度（自动计算速度）
   * @param taskId 任务ID
   * @param progress 进度百分比（0-100）
   * @param uploadedSize 已上传大小（字节）
   */
  updateProgress(taskId: string, progress: number, uploadedSize: number) {
    const task = this.tasks.get(taskId)
    if (!task || task.status === 'paused' || task.status === 'cancelled') {
      return
    }

    const now = Date.now()

    // 1. 确保 uploadedSize 在有效范围内（0 到 file_size）
    uploadedSize = Math.max(0, Math.min(uploadedSize, task.file_size))

    // 2. 防止 uploadedSize 倒退（除非是 prechecking 状态）
    if (task.status !== 'prechecking' && task.lastUploadedSize !== undefined && uploadedSize < task.lastUploadedSize) {
      logger.debug(`任务 ${taskId} uploadedSize 倒退，跳过更新: ${task.lastUploadedSize} -> ${uploadedSize}`)
      return
    }

    // 3. 根据实际 uploadedSize 重新计算进度，确保进度准确
    const calculatedProgress = task.file_size > 0 ? Math.floor((uploadedSize / task.file_size) * 100) : 0
    progress = Math.max(0, Math.min(100, calculatedProgress))

    // 4. 防止进度倒退（除非是 prechecking 状态）
    if (task.status !== 'prechecking' && task.progress !== undefined && progress < task.progress) {
      logger.debug(`任务 ${taskId} progress 倒退，跳过更新: ${task.progress}% -> ${progress}%`)
      return
    }

    // 5. 如果 uploadedSize 超过 file_size，记录警告并限制
    if (uploadedSize > task.file_size) {
      logger.warn(`上传大小超过文件总大小，已限制为文件大小: ${uploadedSize} > ${task.file_size}`)
      uploadedSize = task.file_size
      progress = 100
    }

    // 6. 更新任务状态
    task.progress = progress
    task.uploaded_size = uploadedSize
    task.status = 'uploading'

    // 7. 计算上传速度
    if (!task.speedHistory) {
      task.speedHistory = []
    }

    const shouldUpdateSpeed = !task.lastSpeedUpdateTime || now - task.lastSpeedUpdateTime >= 500

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

    // 8. 更新最后更新时间
    task.lastUpdateTime = now
    task.lastUploadedSize = uploadedSize

    // 9. 通知监听器
    this.notifyListeners()
  }

  /**
   * 标记任务为完成状态
   * @param taskId 任务ID
   * @param isInstantUpload 是否秒传，默认为 false
   */
  completeTask(taskId: string, isInstantUpload: boolean = false) {
    const task = this.tasks.get(taskId)
    if (task) {
      const now = Date.now()
      task.status = 'completed'
      task.progress = 100
      task.uploaded_size = task.file_size
      task.isInstantUpload = isInstantUpload
      
      // 计算总耗时和平均速度
      if (task.startTime) {
        task.endTime = now
        task.totalDuration = now - task.startTime
        
        // 计算平均速度（字节/秒）
        if (task.totalDuration > 0 && task.file_size > 0) {
          task.averageSpeed = (task.file_size / task.totalDuration) * 1000 // 转换为字节/秒
        }
      }
      
      this.saveTasksToStorage()
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
    return task ? (task as any).cancelFunction || null : null
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
      // 支持取消预检中的任务
      if (task.status === 'prechecking' || task.status === 'pending' || task.status === 'uploading' || task.status === 'paused') {
        task.status = 'cancelled'
        task.speed = '0 KB/s'
        task.currentStep = undefined
        this.saveTasksToStorage()
        this.notifyListeners()
      }
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
      // 防止进度倒退：如果更新中包含 progress，且新进度小于当前进度，则不更新 progress
      // 但如果任务状态是 prechecking，则允许更新（因为预检进度可能不同）
      if (updates.progress !== undefined && task.status !== 'prechecking') {
        const currentProgress = task.progress || 0
        if (updates.progress < currentProgress) {
          // 进度倒退，不更新 progress，但更新其他属性
          const { progress, ...otherUpdates } = updates
          Object.assign(task, otherUpdates)
          this.saveTasksToStorage()
          this.notifyListeners()
          return
        }
      }
      
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
      const deletedIds: string[] | null = cache.local.getJSON(this.getDeletedTasksKey())
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
      cache.local.setJSON(this.getDeletedTasksKey(), deletedIds)
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
    return Array.from(this.tasks.values()).sort(
      (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
    )
  }

  /**
   * 清空所有任务
   * @param filterStatus 可选，只清空指定状态的任务。如果不传，清空所有任务
   */
  clearAllTasks(filterStatus?: UploadTask['status'][]): void {
    if (filterStatus && filterStatus.length > 0) {
      // 只清空指定状态的任务
      const tasksToDelete: string[] = []
      this.tasks.forEach((task, taskId) => {
        if (filterStatus.includes(task.status)) {
          if (task.precheckId) {
            this.deletedPrecheckIds.add(task.precheckId)
          }
          tasksToDelete.push(taskId)
        }
      })
      tasksToDelete.forEach(taskId => {
        this.tasks.delete(taskId)
      })
      if (tasksToDelete.length > 0) {
        this.saveDeletedPrecheckIds()
        this.saveTasksToStorage()
        this.notifyListeners()
      }
    } else {
      // 清空所有任务
      this.tasks.forEach(task => {
        if (task.precheckId) {
          this.deletedPrecheckIds.add(task.precheckId)
        }
      })
      this.tasks.clear()
      this.saveDeletedPrecheckIds()
      this.saveTasksToStorage()
      this.notifyListeners()
    }
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
      cache.local.setJSON(this.getStorageKey(), tasks)
    } catch (error) {
      logger.error('保存上传任务到 localStorage 失败:', error)
    }
  }

  private loadTasksFromStorage() {
    try {
      const tasks: UploadTask[] | null = cache.local.getJSON(this.getStorageKey())
      if (tasks && Array.isArray(tasks)) {
        tasks.forEach(task => {
          const validStatuses: UploadTask['status'][] = [
            'prechecking',
            'pending',
            'uploading',
            'paused',
            'completed',
            'failed',
            'cancelled'
          ]
          if (!validStatuses.includes(task.status)) {
            task.status = 'failed'
            task.error = task.error || '任务状态异常'
          }

          // 页面刷新后，预检中的任务转为暂停状态
          if (task.status === 'prechecking') {
            task.status = 'paused'
            task.speed = '0 KB/s'
            task.currentStep = undefined
          } else if (task.status === 'uploading' || task.status === 'pending') {
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
