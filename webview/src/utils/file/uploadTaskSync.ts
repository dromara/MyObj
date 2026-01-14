import { uploadTaskManager, type UploadTask } from './uploadTaskManager'
import { listUncompletedUploads } from '@/api/file'
import logger from '@/plugins/logger'

/**
 * 计算已上传大小（根据分片信息）
 * @param uploadedChunks 已上传分片数
 * @param totalChunks 总分片数
 * @param fileSize 文件总大小（字节）
 * @returns number 已上传大小（字节）
 */
function calculateUploadedSize(uploadedChunks: number, totalChunks: number, fileSize: number): number {
  if (totalChunks <= 0) {
    return 0
  }
  return Math.floor((uploadedChunks / totalChunks) * fileSize)
}

/**
 * 映射后端状态到前端状态
 * @param backendStatus 后端状态字符串
 * @returns UploadTask['status'] 前端状态
 */
function mapBackendStatusToFrontend(backendStatus: string): UploadTask['status'] {
  if (backendStatus === 'completed') {
    return 'completed'
  } else if (backendStatus === 'failed' || backendStatus === 'aborted') {
    return 'failed'
  } else if (backendStatus === 'uploading' || backendStatus === 'pending') {
    return 'paused'
  }
  return 'paused'
}

export interface BackendUploadTask {
  id: string
  file_name: string
  file_size: number
  chunk_size: number
  total_chunks: number
  uploaded_chunks: number
  progress: number
  status: string
  error_message?: string
  path_id: string
  create_time: string
  update_time: string
  expire_time: string
}

/**
 * 同步后端任务到前端任务管理器
 * @param backendTasks 后端任务列表
 * @returns 同步结果统计，包括创建数、更新数、跳过数
 */
export function syncBackendTasksToFrontend(backendTasks: BackendUploadTask[]): {
  created: number
  updated: number
  skipped: number
} {
  const frontendTasks = uploadTaskManager.getAllTasks()
  let created = 0
  let updated = 0
  let skipped = 0

  for (const backendTask of backendTasks) {
    if (uploadTaskManager.isPrecheckIdDeleted(backendTask.id)) {
      skipped++
      continue
    }

    const existingTask = frontendTasks.find(t => t.precheckId === backendTask.id)

    if (!existingTask) {
      const taskId = uploadTaskManager.createTask(backendTask.file_name, backendTask.file_size)

      if (taskId) {
        const uploadedSize = calculateUploadedSize(
          backendTask.uploaded_chunks,
          backendTask.total_chunks,
          backendTask.file_size
        )

        const frontendStatus = mapBackendStatusToFrontend(backendTask.status)

        uploadTaskManager.updateTask(taskId, {
          precheckId: backendTask.id,
          pathId: backendTask.path_id,
          progress: Math.floor(backendTask.progress),
          uploaded_size: uploadedSize,
          status: frontendStatus,
          error: backendTask.error_message
        })
        uploadTaskManager.saveTasksToStorage()
        created++
      }
    } else {
      const uploadedSize = calculateUploadedSize(
        backendTask.uploaded_chunks,
        backendTask.total_chunks,
        backendTask.file_size
      )

      let statusUpdate: Partial<UploadTask> | undefined
      if (backendTask.status === 'completed' && existingTask.status !== 'completed') {
        statusUpdate = { status: 'completed', progress: 100, uploaded_size: existingTask.file_size }
      } else if (
        (backendTask.status === 'failed' || backendTask.status === 'aborted') &&
        existingTask.status !== 'failed'
      ) {
        statusUpdate = { status: 'failed', error: backendTask.error_message }
      }

      uploadTaskManager.updateTask(existingTask.id, {
        progress: Math.floor(backendTask.progress),
        uploaded_size: uploadedSize,
        ...statusUpdate
      })
      updated++
    }
  }

  return { created, updated, skipped }
}

/**
 * 从后端加载并同步未完成的上传任务
 * @returns Promise<同步结果> 包含成功标志、创建数、更新数、跳过数和错误信息
 */
export async function loadAndSyncBackendTasks(): Promise<{
  success: boolean
  created: number
  updated: number
  skipped: number
  error?: string
}> {
  try {
    const response = await listUncompletedUploads()

    if (response.code === 200) {
      // 处理 data 为 null 或 undefined 的情况（视为空数组）
      let backendTasks: BackendUploadTask[] = []

      if (response.data !== null && response.data !== undefined) {
        if (Array.isArray(response.data)) {
          backendTasks = response.data as BackendUploadTask[]
        } else {
          logger.warn('后端返回数据格式错误: data 不是数组类型', response.data)
          return {
            success: false,
            created: 0,
            updated: 0,
            skipped: 0,
            error: '后端返回数据格式错误: data 不是数组类型'
          }
        }
      }

      const result = syncBackendTasksToFrontend(backendTasks)
      return {
        success: true,
        ...result
      }
    }

    logger.warn('后端返回数据格式错误: code 不是 200', response)
    return {
      success: false,
      created: 0,
      updated: 0,
      skipped: 0,
      error: `后端返回数据格式错误: code=${response.code}, message=${response.message || '未知错误'}`
    }
  } catch (error: any) {
    logger.warn('从后端加载未完成上传任务失败:', error)
    return {
      success: false,
      created: 0,
      updated: 0,
      skipped: 0,
      error: error.message || '加载失败'
    }
  }
}

/**
 * 查找后端任务信息（用于恢复上传）
 * @param precheckId 预检ID
 * @returns Promise<BackendUploadTask | null> 后端任务信息，如果不存在则返回null
 */
export async function findBackendTask(precheckId: string): Promise<BackendUploadTask | null> {
  try {
    const response = await listUncompletedUploads()

    if (response.code === 200) {
      // 处理 data 为 null 或 undefined 的情况（视为空数组）
      let backendTasks: BackendUploadTask[] = []

      if (response.data !== null && response.data !== undefined) {
        if (Array.isArray(response.data)) {
          backendTasks = response.data as BackendUploadTask[]
        } else {
          logger.warn('查找后端任务失败: data 不是数组类型', response.data)
          return null
        }
      }

      return backendTasks.find(t => t.id === precheckId) || null
    }

    return null
  } catch (error: any) {
    logger.warn('查找后端任务失败:', error)
    return null
  }
}
