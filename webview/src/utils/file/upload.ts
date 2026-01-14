import SparkMD5 from 'spark-md5'
import { uploadPrecheck, uploadFile } from '@/api/file'
import { UPLOAD_CONFIG } from '@/config/api'
import type { ApiResponse } from '@/types'
import logger from '@/plugins/logger'
import { uploadTaskManager } from './uploadTaskManager'
import i18n from '@/i18n'

export interface UploadConfig {
  chunkSize: number
  maxConcurrentChunks: number
  maxFileSize: number
  maxConcurrentFiles?: number
}

export const DEFAULT_UPLOAD_CONFIG: UploadConfig = {
  chunkSize: UPLOAD_CONFIG.CHUNK_SIZE || 5 * 1024 * 1024,
  maxConcurrentChunks: 3,
  maxFileSize: UPLOAD_CONFIG.MAX_FILE_SIZE || 10 * 1024 * 1024 * 1024,
  maxConcurrentFiles: 2
}

export interface UploadParams {
  file: File
  pathId: string
  config?: Partial<UploadConfig>
  onProgress?: (progress: number, fileName: string) => void
  onSuccess?: (fileName: string) => void
  onError?: (error: Error, fileName: string) => void
  taskId?: string | null
  is_enc?: boolean
  file_password?: string
}

/**
 * 计算文件的MD5哈希值（使用spark-md5分片计算）
 * @param file 要计算MD5的文件对象
 * @param chunkSize 分片大小（字节）
 * @param onProgress 进度回调函数，参数为0-100的进度值
 * @returns Promise<string> MD5哈希值（十六进制字符串）
 */
export const calculateFileMD5 = (
  file: File,
  chunkSize: number,
  onProgress?: (progress: number) => void
): Promise<string> => {
  return new Promise((resolve, reject) => {
    const spark = new SparkMD5.ArrayBuffer()
    const fileReader = new FileReader()
    let currentChunk = 0
    const chunks = Math.ceil(file.size / chunkSize)

    fileReader.onload = e => {
      if (e.target?.result) {
        spark.append(e.target.result as ArrayBuffer)
        currentChunk++

        if (onProgress) {
          onProgress(Math.floor((currentChunk / chunks) * 100))
        }

        if (currentChunk < chunks) {
          loadNextChunk()
        } else {
          resolve(spark.end())
        }
      } else {
        reject(new Error('Failed to read file chunk'))
      }
    }

    fileReader.onerror = () => {
      reject(new Error('Error reading file'))
    }

    const loadNextChunk = () => {
      const start = currentChunk * chunkSize
      const end = Math.min(start + chunkSize, file.size)
      const chunk = file.slice(start, end)
      fileReader.readAsArrayBuffer(chunk)
    }

    loadNextChunk()
  })
}

/**
 * 计算文件分片的MD5哈希值
 * @param chunk 分片数据（Blob对象）
 * @returns Promise<string> 分片MD5哈希值（十六进制字符串）
 */
export const calculateChunkMD5 = (chunk: Blob): Promise<string> => {
  return new Promise((resolve, reject) => {
    const spark = new SparkMD5.ArrayBuffer()
    const fileReader = new FileReader()

    fileReader.onload = e => {
      if (e.target?.result) {
        spark.append(e.target.result as ArrayBuffer)
        resolve(spark.end())
      } else {
        reject(new Error('Failed to read chunk'))
      }
    }

    fileReader.onerror = () => {
      reject(new Error('Error reading chunk'))
    }

    fileReader.readAsArrayBuffer(chunk)
  })
}

const activeUploadTasks = new Map<string, boolean>()
const runningUploadTasks = new Set<string>()

/**
 * 检查上传任务是否还在活动状态（文件对象还在内存中或上传函数仍在运行）
 * @param taskId 任务ID
 * @returns boolean 如果任务在活动状态返回true，否则返回false
 */
export function isUploadTaskActive(taskId: string): boolean {
  // 检查 activeUploadTasks（文件对象还在内存中）
  if (activeUploadTasks.has(taskId)) {
    return true
  }
  // 检查 uploadSingleFile 是否仍在运行
  if (runningUploadTasks.has(taskId)) {
    return true
  }
  return false
}

/**
 * 并发上传控制类
 * 用于管理文件分片的并发上传，支持暂停/继续功能
 */
class ConcurrentUploader {
  private maxConcurrent: number
  private runningChunks: number = 0
  private chunkQueue: (() => Promise<void>)[] = []
  private isPaused: boolean = false
  private taskId: string | null = null

  /**
   * 创建并发上传控制器
   * @param maxConcurrent 最大并发分片数量
   * @param taskId 可选的任务ID，用于检查任务状态
   */
  constructor(maxConcurrent: number, taskId?: string) {
    this.maxConcurrent = maxConcurrent
    this.taskId = taskId || null
  }

  /**
   * 暂停上传
   */
  public pause() {
    this.isPaused = true
  }

  /**
   * 继续上传
   */
  public resume() {
    this.isPaused = false
    this.processQueue()
  }

  /**
   * 检查是否暂停（会检查任务状态）
   * @returns boolean 如果暂停返回true，否则返回false
   */
  private checkPaused(): boolean {
    if (this.taskId) {
      const task = uploadTaskManager.getTask(this.taskId)
      if (task) {
        if (task.status === 'paused') {
          this.isPaused = true
        } else if (task.status === 'uploading' || task.status === 'pending') {
          this.isPaused = false
        }
      }
    }
    return this.isPaused
  }

  /**
   * 添加分片上传任务到队列
   * @param chunkUploadFn 分片上传函数
   */
  public addChunk(chunkUploadFn: () => Promise<void>): void {
    this.chunkQueue.push(chunkUploadFn)
    if (!this.checkPaused()) {
      this.processQueue()
    }
  }

  /**
   * 处理上传队列，按最大并发数执行分片上传
   */
  private processQueue(): void {
    if (this.checkPaused()) {
      return
    }

    while (this.runningChunks < this.maxConcurrent && this.chunkQueue.length > 0 && !this.checkPaused()) {
      const chunkUploadFn = this.chunkQueue.shift()
      if (chunkUploadFn) {
        this.runningChunks++
        chunkUploadFn().finally(() => {
          this.runningChunks--
          if (!this.checkPaused()) {
            this.processQueue()
          }
        })
      }
    }
  }

  /**
   * 等待所有分片上传完成
   * 如果任务被暂停，会等待恢复；如果任务被取消或失败，会直接返回
   * @returns Promise<void>
   */
  public async waitForAll(): Promise<void> {
    while (this.runningChunks > 0 || this.chunkQueue.length > 0) {
      if (this.taskId) {
        const task = uploadTaskManager.getTask(this.taskId)
        if (task) {
          // 如果任务被取消，立即返回
          if (task.status === 'cancelled') {
            return
          }
          // 如果任务被暂停，等待恢复
          if (task.status === 'paused') {
            while (task.status === 'paused') {
              await new Promise(resolve => setTimeout(resolve, 100))
              const currentTask = uploadTaskManager.getTask(this.taskId)
              if (!currentTask) break
              if (currentTask.status === 'cancelled') {
                return
              }
              if (currentTask.status !== 'paused') {
                break
              }
            }
            this.processQueue()
            continue
          }
          // 注意：不再检查 failed 状态，因为单个分片失败不应该阻止其他分片上传
          // 等待所有分片完成后再统一判断任务是否失败
        }
      }

      await new Promise(resolve => setTimeout(resolve, 100))
    }
  }
}

/**
 * 处理单个文件的上传
 * 包括MD5计算、预检、分片上传等完整流程，支持断点续传
 * @param params 上传参数
 * @param params.file 要上传的文件对象
 * @param params.pathId 上传路径ID
 * @param params.config 可选的上传配置
 * @param params.onProgress 进度回调函数
 * @param params.onSuccess 成功回调函数
 * @param params.onError 错误回调函数
 * @param params.taskId 可选的任务ID（用于恢复上传）
 * @returns Promise<ApiResponse<any> | void> 如果秒传成功返回响应，否则返回void
 */
export const uploadSingleFile = async (params: UploadParams): Promise<ApiResponse<any> | void> => {
  const {
    file,
    pathId,
    config = {},
    onProgress,
    onSuccess,
    onError,
    taskId: providedTaskId,
    is_enc = false,
    file_password = ''
  } = params

  const uploadConfig = { ...DEFAULT_UPLOAD_CONFIG, ...config }

  if (file.size > uploadConfig.maxFileSize) {
    const maxSizeMB = Math.round(uploadConfig.maxFileSize / (1024 * 1024))
    throw new Error(
      i18n.global.t('upload.fileSizeExceeded', { fileName: file.name, maxSizeMB }) ||
        `文件 ${file.name} 大小超过限制（最大 ${maxSizeMB}MB）`
    )
  }

  let taskId: string | null = providedTaskId || null

  // 如果已有 taskId，将任务添加到运行中任务集合（用于跟踪 uploadSingleFile 是否仍在运行）
  if (taskId) {
    runningUploadTasks.add(taskId)
  }

  try {
    const fileMD5 = await calculateFileMD5(file, uploadConfig.chunkSize, md5Progress => {
      const progress = Math.floor(md5Progress * 0.1)
      if (taskId) {
        uploadTaskManager.updateProgress(taskId, progress, Math.floor(file.size * md5Progress * 0.1))
      }
      onProgress?.(progress, file.name)
    })

    const totalChunks = Math.ceil(file.size / uploadConfig.chunkSize)
    const filesMD5: string[] = []

    for (let i = 0; i < totalChunks; i++) {
      const start = i * uploadConfig.chunkSize
      const end = Math.min(start + uploadConfig.chunkSize, file.size)
      const chunk = file.slice(start, end)
      const chunkMD5 = await calculateChunkMD5(chunk)
      filesMD5.push(chunkMD5)
    }

    const precheckParams = {
      chunk_signature: fileMD5,
      file_name: file.name,
      file_size: file.size,
      files_md5: filesMD5,
      path_id: pathId
    }

    const precheckResponse = await uploadPrecheck(precheckParams)

    if (precheckResponse.code === 200) {
      onSuccess?.(file.name)
      return precheckResponse
    }

    if (precheckResponse.code !== 201) {
      const errorMsg =
        precheckResponse.message ||
        i18n.global.t('upload.precheckFailed', { fileName: file.name, errorMsg: '' }) ||
        '预检失败'
      ElMessage.error(
        i18n.global.t('upload.precheckFailed', { fileName: file.name, errorMsg }) ||
          `文件 ${file.name} 预检失败: ${errorMsg}`
      )
      throw new Error(errorMsg)
    }

    if (!taskId) {
      try {
        taskId = uploadTaskManager.createTask(file.name, file.size)
        if (taskId) {
          const task = uploadTaskManager.getTask(taskId)
          if (task) {
            task.pathId = pathId
            uploadTaskManager.saveTasksToStorage()
          }
        }
      } catch (err) {
        logger.error('创建上传任务失败:', err)
        throw new Error(i18n.global.t('upload.createTaskFailed') || '创建上传任务失败')
      }
    }

    if (taskId) {
      activeUploadTasks.set(taskId, true)
    }

    let precheckId: string
    let uploadedChunkMd5s: string[] = []

    if (typeof precheckResponse.data === 'string') {
      precheckId = precheckResponse.data
    } else {
      const data = precheckResponse.data as any
      precheckId = data.precheck_id || data
      if (data.md5 && Array.isArray(data.md5)) {
        uploadedChunkMd5s = data.md5
      }
    }

    if (!precheckId) {
      if (taskId) {
        uploadTaskManager.deleteTask(taskId)
      }
      throw new Error('precheck_id获取失败')
    }

    if (taskId) {
      const task = uploadTaskManager.getTask(taskId)
      if (task) {
        task.precheckId = precheckId
        task.chunkSignature = fileMD5
        task.filesMd5 = filesMD5
        task.pathId = pathId
        uploadTaskManager.saveTasksToStorage()
      }
    }

    const uploadedChunks = new Set<number>()
    const concurrentUploader = new ConcurrentUploader(uploadConfig.maxConcurrentChunks, taskId || undefined)

    const cancelFunctions: (() => void)[] = []

    if (taskId) {
      const task = uploadTaskManager.getTask(taskId)
      if (task) {
        ;(task as any).cancelFunctions = cancelFunctions
      }
    }

    if (uploadedChunkMd5s.length > 0 && filesMD5.length > 0) {
      filesMD5.forEach((chunkMd5, index) => {
        if (uploadedChunkMd5s.includes(chunkMd5)) {
          uploadedChunks.add(index)
        }
      })
    }

    if (file.size <= uploadConfig.chunkSize) {
      if (taskId) {
        uploadTaskManager.updateProgress(taskId, 10, Math.floor(file.size * 0.1))
      }

      const uploadParams = {
        precheck_id: precheckId,
        file: file,
        chunk_index: 0,
        total_chunks: 1,
        chunk_md5: fileMD5,
        is_enc: is_enc,
        file_password: file_password
      }

      const uploadResponse = await uploadFile(uploadParams, (_percent, loaded, _total) => {
        if (taskId && loaded !== undefined) {
          const totalProgress = Math.floor(10 + (loaded / file.size) * 90)
          uploadTaskManager.updateProgress(taskId, totalProgress, loaded)
        }
      })

      if (uploadResponse.code === 200) {
        if (taskId) {
          uploadTaskManager.completeTask(taskId)
        }
        onProgress?.(100, file.name)
        onSuccess?.(file.name)
      } else {
        if (taskId) {
          uploadTaskManager.failTask(
            taskId,
            uploadResponse.message || i18n.global.t('upload.uploadFailed') || '上传失败'
          )
        }
        throw new Error(uploadResponse.message)
      }
    } else {
      for (let i = 0; i < totalChunks; i++) {
        const start = i * uploadConfig.chunkSize
        const end = Math.min(start + uploadConfig.chunkSize, file.size)
        const chunk = file.slice(start, end)
        const chunkFile = new File([chunk], file.name, { type: file.type })

        ;(async chunkIndex => {
          const uploadChunkTask = async () => {
            try {
              // 等待任务从暂停状态恢复，或检查是否已取消/失败
              const waitForResumeOrCheckCancelled = async (): Promise<boolean> => {
                if (!taskId) return true

                const task = uploadTaskManager.getTask(taskId)
                if (!task) return false

                // 如果任务已取消或失败，直接返回 false
                if (task.status === 'cancelled' || task.status === 'failed') {
                  return false
                }

                // 如果任务已暂停，等待恢复
                if (task.status === 'paused') {
                  while (task.status === 'paused') {
                    await new Promise(resolve => setTimeout(resolve, 100))
                    const currentTask = uploadTaskManager.getTask(taskId)
                    if (!currentTask) return false
                    if (currentTask.status === 'cancelled' || currentTask.status === 'failed') {
                      return false
                    }
                    if (currentTask.status !== 'paused') {
                      break
                    }
                  }
                }

                return true
              }

              // 在计算MD5前检查任务状态
              if (!(await waitForResumeOrCheckCancelled())) {
                return
              }

              const chunkMD5 = await calculateChunkMD5(chunk)

              // 在计算MD5后再次检查任务状态
              if (!(await waitForResumeOrCheckCancelled())) {
                return
              }

              let cancelUpload: (() => void) | null = null
              await uploadFile(
                {
                  precheck_id: precheckId,
                  file: chunkFile,
                  chunk_index: chunkIndex,
                  total_chunks: totalChunks,
                  chunk_md5: chunkMD5,
                  is_enc: is_enc,
                  file_password: file_password
                },
                (_percent, loaded, _total) => {
                  if (taskId) {
                    const task = uploadTaskManager.getTask(taskId)
                    if (task && task.status === 'paused' && cancelUpload) {
                      cancelUpload()
                      return
                    }
                  }
                  if (taskId && loaded !== undefined) {
                    const chunkUploaded = loaded
                    const previousChunksSize = Array.from(uploadedChunks).reduce((sum, chunkIdx) => {
                      if (chunkIdx === totalChunks - 1) {
                        const lastChunkSize = file.size - (totalChunks - 1) * uploadConfig.chunkSize
                        return sum + lastChunkSize
                      } else {
                        return sum + uploadConfig.chunkSize
                      }
                    }, 0)
                    const currentChunkSize = Math.min(chunkUploaded, uploadConfig.chunkSize)
                    const totalUploaded = previousChunksSize + currentChunkSize
                    const clampedTotalUploaded = Math.min(totalUploaded, file.size)
                    const totalProgress = Math.floor(10 + (clampedTotalUploaded / file.size) * 90)
                    uploadTaskManager.updateProgress(taskId, totalProgress, clampedTotalUploaded)
                  }
                },
                {
                  onCancel: cancel => {
                    cancelUpload = cancel
                    if (taskId && typeof cancel === 'function') {
                      cancelFunctions.push(cancel)
                    }
                  }
                }
              ).catch(error => {
                if (error.message === '上传已取消' || error.message === '请求已取消') {
                  return
                }
                throw error
              })

              uploadedChunks.add(chunkIndex)

              const uploadProgress = Math.floor(10 + (uploadedChunks.size / totalChunks) * 90)
              const uploadedSize = Math.floor((file.size * uploadProgress) / 100)
              if (taskId) {
                uploadTaskManager.updateProgress(taskId, uploadProgress, uploadedSize)
              }
              onProgress?.(uploadProgress, file.name)
            } catch (error) {
              if (error instanceof Error && error.message === '上传已取消') {
                return
              }
              // 不要在这里将整个任务标记为失败，因为可能还有其他分片在上传
              // 只记录错误，等待所有分片完成后再统一判断
              logger.error(`分片 ${chunkIndex} 上传失败:`, error)
              onError?.(error as Error, file.name)
              // 不抛出错误，让其他分片继续上传
              // 错误会在 waitForAll 后通过 uploadedChunks.size 检查发现
            }
          }

          concurrentUploader.addChunk(uploadChunkTask)
        })(i)
      }

      await concurrentUploader.waitForAll()

      if (taskId) {
        const task = uploadTaskManager.getTask(taskId)
        if (task && (task.status === 'paused' || task.status === 'cancelled')) {
          return
        }
      }

      if (uploadedChunks.size !== totalChunks) {
        const failedChunks = totalChunks - uploadedChunks.size
        const errorMessage = `部分分片上传失败（${failedChunks}/${totalChunks} 个分片失败），请重试`
        if (taskId) {
          uploadTaskManager.failTask(taskId, errorMessage)
        }
        throw new Error(errorMessage)
      }

      if (taskId) {
        uploadTaskManager.completeTask(taskId)
      }
      onProgress?.(100, file.name)
      onSuccess?.(file.name)
    }
  } catch (error: any) {
    if (taskId) {
      const task = uploadTaskManager.getTask(taskId)
      if (task && task.status !== 'completed' && task.status !== 'failed') {
        uploadTaskManager.failTask(taskId, error.message || i18n.global.t('upload.uploadFailed') || '上传失败')
      }
    }
    logger.error(`处理文件 ${file.name} 时出错:`, error)
    ElMessage.error(
      i18n.global.t('upload.processFileError', { fileName: file.name, error: error.message }) ||
        `处理文件 ${file.name} 时出错: ${error.message}`
    )
    onError?.(error, file.name)
  } finally {
    if (taskId) {
      // 从运行中任务集合中移除（uploadSingleFile 执行完成）
      runningUploadTasks.delete(taskId)

      const task = uploadTaskManager.getTask(taskId)
      if (task && (task.status === 'completed' || task.status === 'failed' || task.status === 'cancelled')) {
        activeUploadTasks.delete(taskId)
      }
    }
  }
}

/**
 * 处理多文件上传（并行执行）
 * 默认最多同时上传2个文件，最大并行数为5
 * @param files 文件列表
 * @param pathId 上传路径ID
 * @param config 可选的上传配置
 * @param onProgress 进度回调函数
 * @param onSuccess 成功回调函数
 * @param onError 错误回调函数
 * @returns Promise<void>
 */
export const uploadMultipleFiles = async (
  files: File[],
  pathId: string,
  config?: Partial<UploadConfig>,
  onProgress?: (progress: number, fileName: string) => void,
  onSuccess?: (fileName: string) => void,
  onError?: (error: Error, fileName: string) => void,
  is_enc?: boolean,
  file_password?: string
): Promise<void> => {
  const MAX_CONCURRENT_FILES = 5
  const DEFAULT_CONCURRENT_FILES = 2
  const maxConcurrent = Math.min(config?.maxConcurrentFiles ?? DEFAULT_CONCURRENT_FILES, MAX_CONCURRENT_FILES)

  const uploadPromises: Promise<void>[] = []
  const runningUploads = new Set<Promise<void>>()

  for (const file of files) {
    const uploadPromise = (async () => {
      try {
        await uploadSingleFile({
          file,
          pathId,
          config,
          onProgress,
          onSuccess,
          onError,
          is_enc,
          file_password
        })
      } catch (error) {
        logger.error(`文件 ${file.name} 上传失败:`, error)
      }
    })()

    runningUploads.add(uploadPromise)
    uploadPromises.push(uploadPromise)

    uploadPromise.finally(() => {
      runningUploads.delete(uploadPromise)
    })

    if (runningUploads.size >= maxConcurrent) {
      await Promise.race(runningUploads)
    }
  }

  await Promise.all(uploadPromises)
}

/**
 * 打开文件选择对话框
 * @param multiple 是否允许选择多个文件，默认为false
 * @returns Promise<File[]> 选择的文件列表，如果用户取消则返回空数组
 */
export const openFileDialog = (multiple: boolean = false): Promise<File[]> => {
  return new Promise(resolve => {
    const input = document.createElement('input')
    input.type = 'file'
    input.multiple = multiple

    input.onchange = e => {
      const target = e.target as HTMLInputElement
      const files = Array.from(target.files || [])
      resolve(files)
    }

    input.click()
  })
}

/**
 * 处理文件上传（从选择文件到上传完成的完整流程）
 * @param pathId 上传路径ID
 * @param config 可选的上传配置
 * @param onProgress 进度回调函数
 * @param onSuccess 成功回调函数
 * @param onError 错误回调函数
 * @param multiple 是否允许选择多个文件，默认为true
 * @param onFilesSelected 文件选择完成后的回调函数（可用于页面跳转等操作）
 * @returns Promise<void>
 */
export const handleFileUpload = async (
  pathId: string,
  config?: Partial<UploadConfig>,
  onProgress?: (progress: number, fileName: string) => void,
  onSuccess?: (fileName: string) => void,
  onError?: (error: Error, fileName: string) => void,
  multiple: boolean = true,
  onFilesSelected?: () => void,
  encryptConfig?: { is_enc: boolean; file_password: string }
): Promise<void> => {
  try {
    const files = await openFileDialog(multiple)

    if (files.length === 0) {
      return
    }

    if (onFilesSelected) {
      onFilesSelected()
      await new Promise(resolve => setTimeout(resolve, 100))
    }

    await uploadMultipleFiles(
      files,
      pathId,
      config,
      onProgress,
      onSuccess,
      onError,
      encryptConfig?.is_enc,
      encryptConfig?.file_password
    )
  } catch (error: any) {
    logger.error('处理文件上传时出错:', error)
    ElMessage.error(
      i18n.global.t('upload.processUploadError', { error: error.message }) || `处理文件上传时出错: ${error.message}`
    )
  }
}
