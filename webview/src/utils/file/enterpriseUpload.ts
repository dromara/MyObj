import { enterpriseApi } from '@myobj/api'
import { upload } from '@myobj/http'
import { API_ENDPOINTS, UPLOAD_CONFIG } from '@myobj/shared'
import { calculateChunkMD5, calculateFileMD5, DEFAULT_UPLOAD_CONFIG } from './upload'

const { sharedUploadPrecheck } = enterpriseApi

export interface EnterpriseUploadParams {
  file: File
  enterpriseId: string
  pathId?: number
  onProgress?: (percent: number) => void
}

/**
 * 企业共享空间上传：MD5 预检 + 可选分片
 */
export async function uploadEnterpriseFile(params: EnterpriseUploadParams): Promise<{ code: number; message?: string }> {
  const { file, enterpriseId, pathId = 0, onProgress } = params
  const config = { ...DEFAULT_UPLOAD_CONFIG, chunkSize: UPLOAD_CONFIG.CHUNK_SIZE || DEFAULT_UPLOAD_CONFIG.chunkSize }

  onProgress?.(5)
  const fileMD5 = await calculateFileMD5(file, config.chunkSize, p => {
    onProgress?.(Math.min(25, Math.floor(p * 0.25)))
  })

  const precheckRes = await sharedUploadPrecheck({
    enterprise_id: enterpriseId,
    file_name: file.name,
    file_size: file.size,
    path_id: pathId || undefined,
    chunk_signature: fileMD5
  })

  if (precheckRes.code !== 200 && precheckRes.code !== 201) {
    return { code: precheckRes.code, message: precheckRes.message }
  }

  // 秒传成功（无 precheck_id）
  const precheckId = (precheckRes.data as { precheck_id?: string })?.precheck_id
  if (!precheckId) {
    onProgress?.(100)
    return { code: 200, message: precheckRes.message }
  }

  onProgress?.(30)
  const totalChunks = Math.ceil(file.size / config.chunkSize)
  const formBase: Record<string, string> = {
    enterprise_id: enterpriseId,
    precheck_id: precheckId
  }
  if (pathId) formBase.path_id = String(pathId)

  if (totalChunks <= 1) {
    const formData = new FormData()
    Object.entries(formBase).forEach(([k, v]) => formData.append(k, v))
    formData.append('chunk_index', '0')
    formData.append('total_chunks', '1')
    formData.append('chunk_md5', fileMD5)
    const result = await upload(API_ENDPOINTS.ENTERPRISE.SPACE.UPLOAD, file, formData, p => {
      onProgress?.(30 + Math.round(p * 0.7))
    })
    return { code: result.code, message: result.message }
  }

  for (let i = 0; i < totalChunks; i++) {
    const start = i * config.chunkSize
    const end = Math.min(start + config.chunkSize, file.size)
    const chunk = file.slice(start, end)
    const chunkMD5 = await calculateChunkMD5(chunk)
    const chunkFile = new File([chunk], file.name, { type: file.type })
    const formData = new FormData()
    Object.entries(formBase).forEach(([k, v]) => formData.append(k, v))
    formData.append('chunk_index', String(i))
    formData.append('total_chunks', String(totalChunks))
    formData.append('chunk_md5', chunkMD5)

    const result = await upload(API_ENDPOINTS.ENTERPRISE.SPACE.UPLOAD, chunkFile, formData, p => {
      const chunkProgress = ((i + p / 100) / totalChunks) * 70
      onProgress?.(30 + Math.round(chunkProgress))
    })
    if (result.code !== 200) {
      return { code: result.code, message: result.message }
    }
  }

  onProgress?.(100)
  return { code: 200 }
}
