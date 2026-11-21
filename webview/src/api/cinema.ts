import { get } from '@/utils/network/request'
import type { ApiResponse, CompactTag } from '@/types'

export interface CinemaDirectory {
  id: number
  name: string
  parent_id: number
  path: string
}

export interface CinemaVideo {
  file_id: string
  file_name: string
  file_size: number
  mime_type: string
  is_enc: boolean
  public: boolean
  has_thumbnail: boolean
  created_at: string
  directory: CinemaDirectory
  tags?: CompactTag[]
}

export interface CinemaSection {
  directory: CinemaDirectory
  videos: CinemaVideo[]
  total: number
  has_more: boolean
}

export interface CinemaHomeData {
  root: CinemaDirectory
  sections: CinemaSection[]
  total: number
  page: number
  page_size: number
  has_more: boolean
}

export interface CinemaVideoListData {
  root: CinemaDirectory
  directory: CinemaDirectory
  videos: CinemaVideo[]
  total: number
  page: number
  page_size: number
  has_more: boolean
}

export interface CinemaVideoDetailData {
  root: CinemaDirectory
  video: CinemaVideo
}

export interface CinemaRelatedData {
  videos: CinemaVideo[]
  total: number
  page: number
  page_size: number
  has_more: boolean
}

export type CinemaLatestData = CinemaRelatedData

export const getCinemaHome = (rootId: number, page = 1, pageSize = 20) =>
  get<ApiResponse<CinemaHomeData>>(`/cinema/${rootId}/home`, { page, page_size: pageSize })

export const getCinemaFolderVideos = (rootId: number, directoryId: number, page = 1, pageSize = 24) =>
  get<ApiResponse<CinemaVideoListData>>(`/cinema/${rootId}/folders/${directoryId}/videos`, {
    page,
    page_size: pageSize
  })

export const getCinemaLatest = (rootId: number, page = 1, pageSize = 24) =>
  get<ApiResponse<CinemaLatestData>>(`/cinema/${rootId}/latest`, { page, page_size: pageSize })

export const getCinemaVideo = (rootId: number, fileId: string) =>
  get<ApiResponse<CinemaVideoDetailData>>(`/cinema/${rootId}/videos/${fileId}`)

export const getRelatedCinemaVideos = (rootId: number, fileId: string, page = 1, pageSize = 20) =>
  get<ApiResponse<CinemaRelatedData>>(`/cinema/${rootId}/videos/${fileId}/related`, {
    page,
    page_size: pageSize
  })

export const getRandomCinemaVideo = (rootId: number) =>
  get<ApiResponse<CinemaVideoDetailData>>(`/cinema/${rootId}/videos/random`)
