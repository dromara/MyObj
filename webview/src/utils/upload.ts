import SparkMD5 from 'spark-md5'
import { uploadPrecheck, uploadFile } from '@/api/file'
import { ElMessage } from 'element-plus'
import { UPLOAD_CONFIG } from '@/config/api'
import type { ApiResponse } from '@/types'
import logger from '@/plugins/logger'
import { uploadTaskManager } from './uploadTaskManager'

// 配置项接口
export interface UploadConfig {
  chunkSize: number          // 文件分片大小，单位：字节
  maxConcurrentChunks: number // 最大并发分片数量
  maxFileSize: number       // 最大文件大小，单位：字节
}

// 默认配置
export const DEFAULT_UPLOAD_CONFIG: UploadConfig = {
  chunkSize: UPLOAD_CONFIG.CHUNK_SIZE || 5 * 1024 * 1024, // 默认5MB
  maxConcurrentChunks: 3, // 默认并发上传3个分片
  maxFileSize: UPLOAD_CONFIG.MAX_FILE_SIZE || 10 * 1024 * 1024 * 1024 // 默认10GB
}

// 上传参数接口
export interface UploadParams {
  file: File
  pathId: string
  config?: Partial<UploadConfig>
  onProgress?: (progress: number, fileName: string) => void
  onSuccess?: (fileName: string) => void
  onError?: (error: Error, fileName: string) => void
  taskId?: string | null // 可选：已创建的任务ID
}

/**
 * 计算文件的MD5哈希值（使用spark-md5分片计算）
 * @param file 要计算MD5的文件对象
 * @param chunkSize 分片大小
 * @param onProgress 进度回调
 * @returns Promise<string> MD5哈希值（十六进制字符串）
 */
export const calculateFileMD5 = (file: File, chunkSize: number, onProgress?: (progress: number) => void): Promise<string> => {
  return new Promise((resolve, reject) => {
    const spark = new SparkMD5.ArrayBuffer()
    const fileReader = new FileReader()
    let currentChunk = 0
    const chunks = Math.ceil(file.size / chunkSize)

    fileReader.onload = (e) => {
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
 * @param chunk 分片数据
 * @returns Promise<string> 分片MD5哈希值（十六进制字符串）
 */
export const calculateChunkMD5 = (chunk: Blob): Promise<string> => {
  return new Promise((resolve, reject) => {
    const spark = new SparkMD5.ArrayBuffer()
    const fileReader = new FileReader()

    fileReader.onload = (e) => {
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

// 并发上传控制类
class ConcurrentUploader {
  private maxConcurrent: number
  private runningChunks: number = 0
  private chunkQueue: (() => Promise<void>)[] = []
  private isPaused: boolean = false
  private taskId: string | null = null

  constructor(maxConcurrent: number, taskId?: string) {
    this.maxConcurrent = maxConcurrent
    this.taskId = taskId || null
  }

  // 暂停
  public pause() {
    this.isPaused = true
  }

  // 继续
  public resume() {
    this.isPaused = false
    this.processQueue()
  }

  // 检查是否暂停
  private checkPaused(): boolean {
    if (this.taskId) {
      const task = uploadTaskManager.getTask(this.taskId)
      if (task && task.status === 'paused') {
        this.isPaused = true
      } else if (task && task.status === 'uploading') {
        this.isPaused = false
      }
    }
    return this.isPaused
  }

  // 添加分片到队列
  public addChunk(chunkUploadFn: () => Promise<void>): void {
    this.chunkQueue.push(chunkUploadFn)
    if (!this.checkPaused()) {
      this.processQueue()
    }
  }

  // 处理队列
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

  // 等待所有分片上传完成
  public async waitForAll(): Promise<void> {
    while (this.runningChunks > 0 || this.chunkQueue.length > 0) {
      // 检查任务状态
      if (this.taskId) {
        const task = uploadTaskManager.getTask(this.taskId)
        if (task) {
          // 如果任务被取消或失败，直接退出
          if (task.status === 'cancelled' || task.status === 'failed') {
            return
          }
          // 如果任务暂停，等待恢复
          if (task.status === 'paused') {
            while (task.status === 'paused') {
              await new Promise(resolve => setTimeout(resolve, 100))
              const currentTask = uploadTaskManager.getTask(this.taskId)
              if (!currentTask) break
              // 如果恢复后状态变为取消或失败，直接退出
              if (currentTask.status === 'cancelled' || currentTask.status === 'failed') {
                return
              }
              // 如果状态不再是暂停，退出循环
              if (currentTask.status !== 'paused') {
                break
              }
            }
            // 恢复后继续处理队列
            this.processQueue()
            continue
          }
        }
      }
      
      // 正常等待
      await new Promise(resolve => setTimeout(resolve, 100))
    }
  }
}

/**
 * 处理单个文件的上传
 * @param params 上传参数
 */
export const uploadSingleFile = async (params: UploadParams): Promise<ApiResponse<any> | void> => {
  const {
    file,
    pathId,
    config = {},
    onProgress,
    onSuccess,
    onError,
    taskId: providedTaskId
  } = params

  // 合并配置
  const uploadConfig = { ...DEFAULT_UPLOAD_CONFIG, ...config }
  
  // 检查文件大小
  if (file.size > uploadConfig.maxFileSize) {
    const maxSizeMB = Math.round(uploadConfig.maxFileSize / (1024 * 1024))
    throw new Error(`文件 ${file.name} 大小超过限制（最大 ${maxSizeMB}MB）`)
  }
  
  // 使用提供的任务ID，如果没有则创建新任务
  let taskId: string | null = providedTaskId || null
  if (!taskId) {
    try {
      taskId = uploadTaskManager.createTask(file.name, file.size)
    } catch (err) {
      logger.error('创建上传任务失败:', err)
    }
  }
  
  try {
      ElMessage.info(`开始处理文件: ${file.name}`)

      // 1. 计算文件MD5
      const fileMD5 = await calculateFileMD5(file, uploadConfig.chunkSize, (md5Progress) => {
        // MD5计算占10%进度
        const progress = Math.floor(md5Progress * 0.1)
        if (taskId) {
          uploadTaskManager.updateProgress(taskId, progress, Math.floor(file.size * md5Progress * 0.1))
        }
        onProgress?.(progress, file.name)
      })

      // 2. 计算所有分片的MD5
      const totalChunks = Math.ceil(file.size / uploadConfig.chunkSize)
      const filesMD5: string[] = []
      
      for (let i = 0; i < totalChunks; i++) {
        const start = i * uploadConfig.chunkSize
        const end = Math.min(start + uploadConfig.chunkSize, file.size)
        const chunk = file.slice(start, end)
        const chunkMD5 = await calculateChunkMD5(chunk)
        filesMD5.push(chunkMD5)
      }

      // 3. 准备上传预检参数
      const precheckParams = {
        chunk_signature: fileMD5,
        file_name: file.name,
        file_size: file.size,
        files_md5: filesMD5,
        path_id: pathId
      }

      // 4. 调用上传预检接口
      const precheckResponse = await uploadPrecheck(precheckParams)
    logger.debug('上传预检接口响应:', precheckResponse)

    // 秒传成功
    if (precheckResponse.code === 200) {
      if (taskId) {
        uploadTaskManager.completeTask(taskId)
      }
      ElMessage.success(`文件 ${file.name} 秒传成功`)
      onSuccess?.(file.name)
      return precheckResponse
    }

    // 预检成功，准备上传
    if (precheckResponse.code === 201) {
      ElMessage.success(`文件 ${file.name} 预检成功`)

      // 计算分片数
      const totalChunks = Math.ceil(file.size / uploadConfig.chunkSize)
      const uploadedChunks = new Set<number>()
      const concurrentUploader = new ConcurrentUploader(uploadConfig.maxConcurrentChunks, taskId || undefined)
      const precheckId = precheckResponse.data?.precheck_id || precheckResponse.data

      if (!precheckId) {
        if (taskId) {
          uploadTaskManager.failTask(taskId, 'precheck_id获取失败')
        }
        throw new Error('precheck_id获取失败')
      }

      // 4. 上传文件（根据文件大小选择分片或不分片上传）
      if (file.size <= uploadConfig.chunkSize) {
        // 小文件，单分片上传
        if (taskId) {
          uploadTaskManager.updateProgress(taskId, 10, Math.floor(file.size * 0.1))
        }
        
        const uploadParams = {
          precheck_id: precheckId,
          file: file,
          chunk_index: 0,
          total_chunks: 1,
          chunk_md5: fileMD5,
          is_enc: false
        }

        const uploadResponse = await uploadFile(uploadParams, (_percent, loaded, _total) => {
          // 小文件上传进度，实时更新
          if (taskId && loaded !== undefined) {
            const totalProgress = Math.floor(10 + (loaded / file.size) * 90)
            uploadTaskManager.updateProgress(taskId, totalProgress, loaded)
          }
        })
        logger.debug('上传接口响应:', uploadResponse)

        if (uploadResponse.code === 200) {
          // 更新上传进度
          if (taskId) {
            uploadTaskManager.completeTask(taskId)
          }
          onProgress?.(100, file.name)
          ElMessage.success(`文件 ${file.name} 上传成功`)
          onSuccess?.(file.name)
        } else {
          if (taskId) {
            uploadTaskManager.failTask(taskId, uploadResponse.message || '上传失败')
          }
          throw new Error(uploadResponse.message)
        }
      } else {
        // 大文件，分片上传
        for (let i = 0; i < totalChunks; i++) {
          const start = i * uploadConfig.chunkSize
          const end = Math.min(start + uploadConfig.chunkSize, file.size)
          const chunk = file.slice(start, end)
          // 将Blob转换为File类型
          const chunkFile = new File([chunk], file.name, { type: file.type })

          // 使用立即执行函数捕获当前索引
          ;(async (chunkIndex) => {
            // 创建上传任务
            const uploadChunkTask = async () => {
              try {
                // 检查是否暂停
                if (taskId) {
                  const task = uploadTaskManager.getTask(taskId)
                  if (task && task.status === 'paused') {
                    // 等待恢复
                    while (task.status === 'paused') {
                      await new Promise(resolve => setTimeout(resolve, 100))
                      const currentTask = uploadTaskManager.getTask(taskId)
                      if (!currentTask || currentTask.status !== 'paused') break
                    }
                  }
                  // 如果任务被取消，停止上传
                  const currentTask = uploadTaskManager.getTask(taskId)
                  if (currentTask && (currentTask.status === 'cancelled' || currentTask.status === 'failed')) {
                    return
                  }
                }

                // 计算分片MD5
                const chunkMD5 = await calculateChunkMD5(chunk)

                // 再次检查暂停状态
                if (taskId) {
                  const task = uploadTaskManager.getTask(taskId)
                  if (task && task.status === 'paused') {
                    // 暂停时等待恢复
                    while (task.status === 'paused') {
                      await new Promise(resolve => setTimeout(resolve, 100))
                      const currentTask = uploadTaskManager.getTask(taskId)
                      if (!currentTask || currentTask.status !== 'paused') break
                    }
                    // 如果恢复后状态变为取消或失败，直接返回
                    const currentTask = uploadTaskManager.getTask(taskId)
                    if (currentTask && (currentTask.status === 'cancelled' || currentTask.status === 'failed')) {
                      return
                    }
                  }
                  // 如果任务被取消或失败，停止上传
                  const currentTask = uploadTaskManager.getTask(taskId)
                  if (currentTask && (currentTask.status === 'cancelled' || currentTask.status === 'failed')) {
                    return
                  }
                }

                // 上传分片（带进度回调，用于实时更新总进度和速度）
                await uploadFile({
                  precheck_id: precheckId,
                  file: chunkFile,
                  chunk_index: chunkIndex,
                  total_chunks: totalChunks,
                  chunk_md5: chunkMD5,
                  is_enc: false // 默认不加密
                }, (_percent, loaded, _total) => {
                  // 分片上传进度，用于实时更新总进度
                  if (taskId && loaded !== undefined) {
                    // 计算当前分片已上传大小
                    const chunkUploaded = loaded
                    // 计算总已上传大小（之前已完成的分片 + 当前分片已上传部分）
                    // 注意：这里只计算已完成的分片，不包括当前正在上传的分片
                    const previousChunksSize = Array.from(uploadedChunks).reduce((sum) => {
                      return sum + uploadConfig.chunkSize
                    }, 0)
                    // 当前分片的上传大小不能超过分片大小
                    const currentChunkSize = Math.min(chunkUploaded, uploadConfig.chunkSize)
                    const totalUploaded = previousChunksSize + currentChunkSize
                    // 确保不超过文件总大小
                    const clampedTotalUploaded = Math.min(totalUploaded, file.size)
                    const totalProgress = Math.floor(10 + (clampedTotalUploaded / file.size) * 90)
                    // 更新进度（会自动计算速度）
                    uploadTaskManager.updateProgress(taskId, totalProgress, clampedTotalUploaded)
                  }
                })

                // 标记分片上传完成
                uploadedChunks.add(chunkIndex)

                // 更新上传进度（MD5计算10% + 分片上传90%）
                const uploadProgress = Math.floor(10 + (uploadedChunks.size / totalChunks) * 90)
                const uploadedSize = Math.floor(file.size * uploadProgress / 100)
                if (taskId) {
                  // 自动计算速度
                  uploadTaskManager.updateProgress(taskId, uploadProgress, uploadedSize)
                }
                onProgress?.(uploadProgress, file.name)
              } catch (error) {
                // 如果是取消错误，不标记为失败
                if (error instanceof Error && error.message === '上传已取消') {
                  return
                }
                if (taskId) {
                  const task = uploadTaskManager.getTask(taskId)
                  if (task && task.status !== 'cancelled') {
                    uploadTaskManager.failTask(taskId, error instanceof Error ? error.message : '上传失败')
                  }
                }
                onError?.(error as Error, file.name)
                throw error
              }
            }

            // 将任务添加到并发队列
            concurrentUploader.addChunk(uploadChunkTask)
          })(i)
        }

        // 5. 等待所有分片上传完成
        await concurrentUploader.waitForAll()

        // 6. 检查是否所有分片都上传成功
        // 如果任务是暂停或取消状态，不检查分片完成情况
        if (taskId) {
          const task = uploadTaskManager.getTask(taskId)
          if (task && (task.status === 'paused' || task.status === 'cancelled')) {
            // 暂停或取消状态，不抛出错误，直接返回
            return
          }
        }
        
        if (uploadedChunks.size !== totalChunks) {
          if (taskId) {
            uploadTaskManager.failTask(taskId, '部分分片上传失败，请重试')
          }
          throw new Error('部分分片上传失败，请重试')
        }

        // 更新上传进度为100%
        if (taskId) {
          uploadTaskManager.completeTask(taskId)
        }
        onProgress?.(100, file.name)
        ElMessage.success(`文件 ${file.name} 上传成功`)
        onSuccess?.(file.name)
      }
    } else {
      // 预检失败
      if (taskId) {
        uploadTaskManager.failTask(taskId, precheckResponse.message || '预检失败')
      }
      throw new Error(precheckResponse.message)
    }
  } catch (error: any) {
    // 如果任务已创建但未处理，标记为失败
    if (taskId) {
      const task = uploadTaskManager.getTask(taskId)
      if (task && task.status !== 'completed' && task.status !== 'failed') {
        uploadTaskManager.failTask(taskId, error.message || '上传失败')
      }
    }
    logger.error(`处理文件 ${file.name} 时出错:`, error)
    ElMessage.error(`处理文件 ${file.name} 时出错: ${error.message}`)
    onError?.(error, file.name)
  }
}

/**
 * 处理多文件上传
 * @param files 文件列表
 * @param pathId 路径ID
 * @param config 配置
 * @param onProgress 进度回调
 * @param onSuccess 成功回调
 * @param onError 错误回调
 */
export const uploadMultipleFiles = async (
  files: File[],
  pathId: string,
  config?: Partial<UploadConfig>,
  onProgress?: (progress: number, fileName: string) => void,
  onSuccess?: (fileName: string) => void,
  onError?: (error: Error, fileName: string) => void
): Promise<void> => {
  // 先为所有文件创建任务，确保所有任务立即显示
  const fileTaskMap = new Map<File, string>()
  for (const file of files) {
    try {
      const taskId = uploadTaskManager.createTask(file.name, file.size)
      fileTaskMap.set(file, taskId)
    } catch (err) {
      logger.error(`为文件 ${file.name} 创建任务失败:`, err)
    }
  }

  // 遍历处理每个文件
  for (const file of files) {
    const taskId = fileTaskMap.get(file)
    await uploadSingleFile({
      file,
      pathId,
      config,
      onProgress,
      onSuccess,
      onError,
      taskId // 传递已创建的任务ID
    })
  }
}

/**
 * 打开文件选择对话框
 * @param multiple 是否允许选择多个文件
 * @returns Promise<File[]> 选择的文件列表
 */
export const openFileDialog = (multiple: boolean = false): Promise<File[]> => {
  return new Promise((resolve) => {
    const input = document.createElement('input')
    input.type = 'file'
    input.multiple = multiple

    input.onchange = (e) => {
      const target = e.target as HTMLInputElement
      const files = Array.from(target.files || [])
      resolve(files)
    }

    input.click()
  })
}

/**
 * 处理文件上传（从选择文件到上传完成的完整流程）
 * @param params 上传参数
 */
export const handleFileUpload = async (
  pathId: string,
  config?: Partial<UploadConfig>,
  onProgress?: (progress: number, fileName: string) => void,
  onSuccess?: (fileName: string) => void,
  onError?: (error: Error, fileName: string) => void,
  multiple: boolean = true,
  onFilesSelected?: () => void
): Promise<void> => {
  try {
    // 1. 选择文件
    const files = await openFileDialog(multiple)
    
    if (files.length === 0) {
      return
    }

    // 2. 文件选择完成后，调用回调（用于跳转等操作）
    if (onFilesSelected) {
      onFilesSelected()
      // 等待一下让页面跳转完成，再开始上传
      await new Promise(resolve => setTimeout(resolve, 100))
    }

    // 3. 显示上传提示
    ElMessage.info(`开始上传 ${files.length} 个文件`)

    // 4. 上传文件
    await uploadMultipleFiles(files, pathId, config, onProgress, onSuccess, onError)
  } catch (error: any) {
    logger.error('处理文件上传时出错:', error)
    ElMessage.error(`处理文件上传时出错: ${error.message}`)
  }
}
