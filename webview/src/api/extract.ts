import { get, post } from '@/utils/network/request'
import { filterParams } from '@/utils/common/params'
import { API_ENDPOINTS } from '@/config/api'
import type { ApiResponse } from '@/types'

export interface ExtractCheckRequest {
  file_id: string
  target_path_id: string
  file_password?: string
}

export interface ExtractCheckResponse {
  has_conflict: boolean
  conflict_files: string[]
  total_files: number
}

export interface CreateExtractRequest {
  file_id: string
  target_path_id: string
  file_password?: string
  conflict_resolution?: string
}

export interface CreateExtractResponse {
  task_id: string
  archive_name: string
  archive_type: string
  total_files: number
  total_size: number
  status: string
}

export interface ExtractProgressResponse {
  task_id: string
  status: string
  progress: number
  current_file: string
  current_index: number
  total_files: number
  completed: number
  failed: number
  skipped: number
  error_msg: string
}

export const checkExtractConflict = (data: ExtractCheckRequest) => {
  return post<ApiResponse<ExtractCheckResponse>>(API_ENDPOINTS.EXTRACT.CHECK, data)
}

export const createExtract = (data: CreateExtractRequest) => {
  return post<ApiResponse<CreateExtractResponse>>(API_ENDPOINTS.EXTRACT.CREATE, data)
}

export const getExtractProgress = (taskId: string) => {
  return get<ApiResponse<ExtractProgressResponse>>(
    API_ENDPOINTS.EXTRACT.PROGRESS,
    filterParams({ task_id: taskId })
  )
}
