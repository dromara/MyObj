import { get, post } from '@myobj/http'
import { API_ENDPOINTS } from '@myobj/shared'
import type { ApiResponse } from '@myobj/shared'

export type CloudAuthType = 'cookie' | 'refresh_token' | 'oauth2' | 'share_link'

export interface CredentialField {
  key: string
  label: string
  required: boolean
  secret?: boolean
  help?: string
}

export interface CloudProviderInfo {
  id: string
  name: string
  auth_type: CloudAuthType
  description?: string
  max_file_size?: number
  enabled: boolean
  requires_proxy?: boolean
  credential_fields?: CredentialField[]
}

export interface OAuthProviderInfo {
  id: string
  name: string
  scopes?: string[]
  enabled: boolean
  authorize_url?: string
}

export interface CloudProvidersResponse {
  providers: CloudProviderInfo[]
  oauth_providers: OAuthProviderInfo[]
}

export interface CloudFileInfo {
  fid: string
  file_name: string
  size: number
  is_dir: boolean
}

export interface CloudUserInfo {
  provider: string
  nickname: string
  total_size: number
  used_size: number
  binding_id?: string
  oauth_binding_id?: string
}

export interface CloudCredentialBinding {
  id: string
  provider: string
  account_name: string
  updated_at?: string
  created_at?: string
}

export interface LanzouParseResult {
  download_url: string
  file_name: string
  file_size: number
  file_size_text?: string
}

export interface CreateCloudDownloadRequest {
  provider: string
  cookie?: string
  binding_id?: string
  oauth_binding_id?: string
  file_id: string
  file_name?: string
  file_size?: number
  virtual_path?: string
  enable_encryption?: boolean
  file_password?: string
}

export interface ParseCloudShareRequest {
  provider: string
  share_url: string
  password?: string
  extra?: Record<string, string>
}

export interface CreateCloudShareDownloadRequest {
  provider: string
  share_url: string
  password?: string
  extra?: Record<string, string>
  virtual_path?: string
  enable_encryption?: boolean
  file_password?: string
}

export interface CreateLanzouDownloadRequest {
  share_url: string
  password?: string
  virtual_path?: string
  enable_encryption?: boolean
  file_password?: string
}

export const getCloudProviders = () => {
  return get<ApiResponse<CloudProvidersResponse>>(API_ENDPOINTS.DOWNLOAD.CLOUD_PROVIDERS)
}

export const validateCloudCredential = (data: {
  provider: string
  cookie?: string
  binding_id?: string
  oauth_binding_id?: string
  save_binding?: boolean
}) => {
  return post<ApiResponse<CloudUserInfo>>(API_ENDPOINTS.DOWNLOAD.CLOUD_VALIDATE, data)
}

export const listCloudFiles = (data: {
  provider: string
  cookie?: string
  binding_id?: string
  oauth_binding_id?: string
  pdir_fid?: string
  page: number
  page_size: number
}) => {
  return post<
    ApiResponse<{
      files: CloudFileInfo[]
      total: number
      page: number
      page_size: number
    }>
  >(API_ENDPOINTS.DOWNLOAD.CLOUD_FILES, data)
}

export const createCloudDownload = (data: CreateCloudDownloadRequest) => {
  return post<ApiResponse>(API_ENDPOINTS.DOWNLOAD.CLOUD_CREATE, data)
}

export const parseLanzouShare = (data: { share_url: string; password?: string }) => {
  return post<ApiResponse<LanzouParseResult>>(API_ENDPOINTS.DOWNLOAD.LANZOU_PARSE, data)
}

export const createLanzouDownload = (data: CreateLanzouDownloadRequest) => {
  return post<ApiResponse>(API_ENDPOINTS.DOWNLOAD.LANZOU_CREATE, data)
}

export const startCloudOAuth = (provider: string) => {
  return get<ApiResponse<{ authorize_url: string; state: string }>>(
    API_ENDPOINTS.DOWNLOAD.CLOUD_OAUTH_AUTHORIZE(provider)
  )
}

export const listCloudOAuthBindings = () => {
  return get<ApiResponse<Array<{ id: string; provider: string; account_name: string }>>>(
    API_ENDPOINTS.DOWNLOAD.CLOUD_OAUTH_BINDINGS
  )
}

export const deleteCloudOAuthBinding = (bindingId: string) => {
  return post<ApiResponse>(API_ENDPOINTS.DOWNLOAD.CLOUD_OAUTH_UNBIND, { binding_id: bindingId })
}

export const listCloudCredentialBindings = () => {
  return get<ApiResponse<CloudCredentialBinding[]>>(API_ENDPOINTS.DOWNLOAD.CLOUD_BINDINGS)
}

export const deleteCloudCredentialBinding = (bindingId: string) => {
  return post<ApiResponse>(API_ENDPOINTS.DOWNLOAD.CLOUD_BINDINGS_UNBIND, { binding_id: bindingId })
}

export const parseCloudShare = (data: ParseCloudShareRequest) => {
  return post<ApiResponse<LanzouParseResult>>(API_ENDPOINTS.DOWNLOAD.CLOUD_SHARE_PARSE, data)
}

export const createCloudShareDownload = (data: CreateCloudShareDownloadRequest) => {
  return post<ApiResponse>(API_ENDPOINTS.DOWNLOAD.CLOUD_SHARE_CREATE, data)
}
