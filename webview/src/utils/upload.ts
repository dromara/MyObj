import SparkMD5 from 'spark-md5'
import { uploadPrecheck, uploadFile } from '@/api/file'
import { ElMessage } from 'element-plus'
import { UPLOAD_CONFIG } from '@/config/api'
import type { ApiResponse } from '@/types'

// 配置项接口
export interface UploadConfig {
  chunkSize: number          // 文件分片大小，单位：字节
  maxConcurrentChunks: number // 最大并发分片数量
  maxFileSize: number       // 最大文件大小，单位：字节
}

// 默认配置
export const DEFAULT_UPLOAD_CONFIG: UploadConfig = {
  chunkSize: UPLOAD_CONFIG.CHUNK_SIZE || 5 * 1024 * 1024, // 默认5MB
  maxConcurrentChunks: 1, // 默认并发上传3个分片
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

  constructor(maxConcurrent: number) {
    this.maxConcurrent = maxConcurrent
  }

  // 添加分片到队列
  public addChunk(chunkUploadFn: () => Promise<void>): void {
    this.chunkQueue.push(chunkUploadFn)
    this.processQueue()
  }

  // 处理队列
  private processQueue(): void {
    while (this.runningChunks < this.maxConcurrent && this.chunkQueue.length > 0) {
      const chunkUploadFn = this.chunkQueue.shift()
      if (chunkUploadFn) {
        this.runningChunks++
        chunkUploadFn().finally(() => {
          this.runningChunks--
          this.processQueue()
        })
      }
    }
  }

  // 等待所有分片上传完成
  public async waitForAll(): Promise<void> {
    while (this.runningChunks > 0 || this.chunkQueue.length > 0) {
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
    onError
  } = params

  // 合并配置
  const uploadConfig = { ...DEFAULT_UPLOAD_CONFIG, ...config }
  
  // 检查文件大小
  if (file.size > uploadConfig.maxFileSize) {
    const maxSizeMB = Math.round(uploadConfig.maxFileSize / (1024 * 1024))
    throw new Error(`文件 ${file.name} 大小超过限制（最大 ${maxSizeMB}MB）`)
  }
  
  try {
      ElMessage.info(`开始处理文件: ${file.name}`)

      // 1. 计算文件MD5
      const fileMD5 = await calculateFileMD5(file, uploadConfig.chunkSize, (progress) => {
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
    console.log('上传预检接口响应:', precheckResponse)

    // 秒传成功
    if (precheckResponse.code === 200) {
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
      const concurrentUploader = new ConcurrentUploader(uploadConfig.maxConcurrentChunks)
      const precheckId = precheckResponse.data?.precheck_id || precheckResponse.data

      if (!precheckId) {
        throw new Error('precheck_id获取失败')
      }

      // 4. 上传文件（根据文件大小选择分片或不分片上传）
      if (file.size <= uploadConfig.chunkSize) {
        // 小文件，单分片上传
        const uploadParams = {
          precheck_id: precheckId,
          file: file,
          chunk_index: 0,
          total_chunks: 1,
          chunk_md5: fileMD5,
          is_enc: false
        }

        const uploadResponse = await uploadFile(uploadParams)
        console.log('上传接口响应:', uploadResponse)

        if (uploadResponse.code === 200) {
          // 更新上传进度
          onProgress?.(100, file.name)
          ElMessage.success(`文件 ${file.name} 上传成功`)
          onSuccess?.(file.name)
        } else {
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
                // 计算分片MD5
                const chunkMD5 = await calculateChunkMD5(chunk)

                // 上传分片
                await uploadFile({
                  precheck_id: precheckId,
                  file: chunkFile,
                  chunk_index: chunkIndex,
                  total_chunks: totalChunks,
                  chunk_md5: chunkMD5,
                  is_enc: false // 默认不加密
                })

                // 标记分片上传完成
                uploadedChunks.add(chunkIndex)

                // 更新上传进度（MD5计算10% + 分片上传90%）
                const uploadProgress = Math.floor(10 + (uploadedChunks.size / totalChunks) * 90)
                onProgress?.(uploadProgress, file.name)
              } catch (error) {
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
        if (uploadedChunks.size !== totalChunks) {
          throw new Error('部分分片上传失败，请重试')
        }

        // 更新上传进度为100%
        onProgress?.(100, file.name)
        ElMessage.success(`文件 ${file.name} 上传成功`)
        onSuccess?.(file.name)
      }
    } else {
      throw new Error(precheckResponse.message)
    }
  } catch (error: any) {
    console.error(`处理文件 ${file.name} 时出错:`, error)
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
  // 遍历处理每个文件
  for (const file of files) {
    await uploadSingleFile({
      file,
      pathId,
      config,
      onProgress,
      onSuccess,
      onError
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
  multiple: boolean = true
): Promise<void> => {
  try {
    // 1. 选择文件
    const files = await openFileDialog(multiple)
    
    if (files.length === 0) {
      return
    }

    // 2. 显示上传提示
    ElMessage.info(`开始上传 ${files.length} 个文件`)

    // 3. 上传文件
    await uploadMultipleFiles(files, pathId, config, onProgress, onSuccess, onError)
  } catch (error: any) {
    console.error('处理文件上传时出错:', error)
    ElMessage.error(`处理文件上传时出错: ${error.message}`)
  }
}
