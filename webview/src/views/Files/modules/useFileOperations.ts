import { deleteFiles, setFilePublic } from '@/api/file'
import { createPackage, getPackageProgress, downloadPackage } from '@/api/package'
import { useFileDownload } from '@/composables/useFileDownload'
import { useUserStore } from '@/stores/user'
import { useI18n } from '@/composables/useI18n'
import cache from '@/plugins/cache'
import type { FileItem, FileListResponse } from '@/types'

export function useFileOperations(
  fileListData: Ref<FileListResponse>,
  selectedFileIds: Ref<string[]>,
  selectedFolderIds: Ref<number[]>,
  loadFileList: () => Promise<void>
) {
  const { t } = useI18n()
  const { proxy } = getCurrentInstance() as ComponentInternalInstance
  const router = useRouter()
  const userStore = useUserStore()

  const previewVisible = ref(false)
  const previewFile = ref<FileItem | null>(null)
  const showShareDialog = ref(false)
  const shareForm = reactive({
    file_id: '',
    file_name: '',
    file_size: 0
  })
  
  // 使用统一的文件下载 composable（Files 页面需要跳转到任务中心）
  const {
    showDownloadPasswordDialog,
    downloadPasswordForm,
    downloadingFile,
    handleDownload: handleFileDownload,
    confirmDownloadPassword
  } = useFileDownload({
    onTaskReady: () => {
      // 任务准备完成时，跳转到任务中心
      router.push({
        path: '/tasks',
        query: { tab: 'download' }
      })
    }
  })

  const getFileSize = (fileId: string): number => {
    const file = fileListData.value.files.find((f: any) => f.file_id === fileId)
    return file?.file_size || 0
  }

  const handleShareSuccess = () => {
    selectedFileIds.value = []
  }

  const handleFilePreview = (file: FileItem) => {
    previewFile.value = file
    previewVisible.value = true
  }

  const handleShareFile = (file: FileItem) => {
    shareForm.file_id = file.file_id
    shareForm.file_name = file.file_name
    shareForm.file_size = file.file_size
    showShareDialog.value = true
  }

  const handleDownloadFile = async (file: FileItem) => {
    await handleFileDownload(file)
  }

  const handleDeleteFile = async (file: FileItem) => {
    try {
      await proxy?.$modal.confirm(t('files.confirmDeleteFile', { fileName: file.file_name }))
      try {
        const result = await deleteFiles({ file_ids: [file.file_id] })
        if (result.code === 200) {
          proxy?.$modal.msgSuccess(t('files.deleteSuccess'))
          selectedFileIds.value = []
          selectedFolderIds.value = []
          await loadFileList()
          // 删除成功后刷新用户信息，更新存储空间显示
          await userStore.fetchUserInfo()
        } else {
          proxy?.$modal.msgError(result.message || t('files.deleteFailed'))
        }
      } catch (error: any) {
        if (error !== 'cancel') {
          proxy?.$modal.msgError(error.message || t('files.deleteFailed'))
        }
      }
    } catch (error: any) {
      if (error !== 'cancel') {
        // 用户取消操作
      }
    }
  }

  const handleToolbarDownload = async () => {
    if (selectedFileIds.value.length === 0) {
      proxy?.$modal.msgWarning(t('files.selectFilesFirst'))
      return
    }
    
    if (selectedFileIds.value.length === 1) {
      const fileId = selectedFileIds.value[0]
      const file = fileListData.value.files.find((f: any) => f.file_id === fileId)
      if (file) {
        await handleDownloadFile(file)
      }
    } else {
      // 多文件打包下载
      await handlePackageDownload()
    }
  }

  // 打包下载
  const handlePackageDownload = async () => {
    if (selectedFileIds.value.length === 0) {
      proxy?.$modal.msgWarning(t('files.selectFilesFirst'))
      return
    }

    const packageName = `files_${Date.now()}.zip`

    try {
      // 创建打包任务
      proxy?.$modal.loading(t('files.creatingPackage'))

      const res = await createPackage({
        file_ids: selectedFileIds.value,
        package_name: packageName
      })
      
      if (res.code === 200 && res.data) {
        const packageId = res.data.package_id
        
        // 如果状态是 ready，直接下载
        if (res.data.status === 'ready') {
          const downloadUrl = downloadPackage(packageId)
          // 使用 fetch 先检查响应，确保是文件而不是 JSON 错误
          try {
            // 获取 token 用于 Authorization header
            const token = cache.local.get('token')
            const response = await fetch(downloadUrl, {
              method: 'GET',
              headers: {
                'Authorization': token ? `Bearer ${token}` : '',
              },
              credentials: 'include', // 同时携带 Cookie 作为备用
            })
            
            // 检查响应类型，如果是 JSON 错误则显示错误信息
            const contentType = response.headers.get('content-type') || ''
            if (contentType.includes('application/json') || !response.ok) {
              const errorData = await response.json()
              proxy?.$modal.closeLoading()
              proxy?.$modal.msgError(errorData.message || t('files.downloadFailed'))
              return
            }
            
            // 是文件，创建 blob 并下载
            const blob = await response.blob()
            const blobUrl = window.URL.createObjectURL(blob)
            const a = document.createElement('a')
            a.href = blobUrl
            a.download = res.data.package_name || packageName
            a.style.display = 'none'
            document.body.appendChild(a)
            a.click()
            document.body.removeChild(a)
            window.URL.revokeObjectURL(blobUrl)
            proxy?.$modal.closeLoading()
            proxy?.$modal.msgSuccess(t('files.downloadStart'))
          } catch (error: any) {
            proxy?.$modal.closeLoading()
            proxy?.$modal.msgError(error.message || t('files.downloadFailed'))
          }
          return
        }
        
        // 如果状态是 creating，轮询进度
        if (res.data.status === 'creating') {
          await pollPackageProgress(packageId, packageName)
        }
      } else {
        proxy?.$modal.closeLoading()
        if (res.code === 404 || res.message?.includes('404')) {
          proxy?.$modal.msg(t('files.packageFeaturePending'))
        } else {
          proxy?.$modal.msgError(res.message || t('files.createPackageFailed'))
        }
      }
    } catch (error: any) {
      proxy?.$modal.closeLoading()
      if (error.response?.status === 404 || error.message?.includes('404')) {
        proxy?.$modal.msg(t('files.packageFeaturePending'))
      } else {
        proxy?.$modal.msgError(error.message || t('files.packageDownloadFailed'))
      }
      proxy?.$log?.error(error)
    }
  }

  // 轮询打包进度
  const pollPackageProgress = async (packageId: string, packageName: string) => {
    const maxAttempts = 60 // 最多轮询60次（5分钟）
    let attempts = 0
    
    const poll = async () => {
      try {
        const res = await getPackageProgress(packageId)
        
        if (res.code === 200 && res.data) {
          const { status, progress } = res.data
          
          if (status === 'ready') {
            proxy?.$modal.closeLoading()
            const downloadUrl = downloadPackage(packageId)
            // 使用 fetch 先检查响应，确保是文件而不是 JSON 错误
            try {
              // 获取 token 用于 Authorization header
              const token = cache.local.get('token')
              const response = await fetch(downloadUrl, {
                method: 'GET',
                headers: {
                  'Authorization': token ? `Bearer ${token}` : '',
                },
                credentials: 'include', // 同时携带 Cookie 作为备用
              })
              
              // 检查响应类型，如果是 JSON 错误则显示错误信息
              const contentType = response.headers.get('content-type') || ''
              if (contentType.includes('application/json') || !response.ok) {
                const errorData = await response.json()
                proxy?.$modal.msgError(errorData.message || t('files.downloadFailed'))
                return
              }
              
              // 是文件，创建 blob 并下载
              const blob = await response.blob()
              const blobUrl = window.URL.createObjectURL(blob)
              const a = document.createElement('a')
              a.href = blobUrl
              a.download = packageName
              a.style.display = 'none'
              document.body.appendChild(a)
              a.click()
              document.body.removeChild(a)
              window.URL.revokeObjectURL(blobUrl)
              proxy?.$modal.msgSuccess(t('files.packageReady'))
            } catch (error: any) {
              proxy?.$modal.msgError(error.message || t('files.downloadFailed'))
            }
            return
          }
          
          if (status === 'failed') {
            proxy?.$modal.closeLoading()
            proxy?.$modal.msgError(res.data.error_msg || t('files.packageFailed'))
            return
          }
          
          // 更新进度提示
          if (progress < 100) {
            proxy?.$modal.loading(t('files.packaging', { progress }))
          }
          
          // 继续轮询
          if (attempts < maxAttempts && status === 'creating') {
            attempts++
            setTimeout(poll, 5000) // 每5秒轮询一次
          } else if (attempts >= maxAttempts) {
            proxy?.$modal.closeLoading()
            proxy?.$modal.msgError(t('files.packageTimeout'))
          }
        } else {
          proxy?.$modal.closeLoading()
          if (res.code === 404 || res.message?.includes('404')) {
            proxy?.$modal.msg(t('files.packageFeaturePending'))
          } else {
            proxy?.$modal.msgError(res.message || t('files.getPackageProgressFailed'))
          }
        }
      } catch (error: any) {
        proxy?.$modal.closeLoading()
        if (error.response?.status === 404 || error.message?.includes('404')) {
          proxy?.$modal.msg(t('files.packageFeaturePending'))
        } else {
          proxy?.$modal.msgError(t('files.getPackageProgressFailed'))
        }
        proxy?.$log?.error(error)
      }
    }
    
    poll()
  }

  const handleToolbarShare = () => {
    if (selectedFileIds.value.length === 0) {
      proxy?.$modal.msgWarning(t('files.selectShareFilesFirst'))
      return
    }
    if (selectedFileIds.value.length > 1) {
      proxy?.$modal.msgWarning(t('files.onlyOneShare'))
      return
    }
    
    const fileId = selectedFileIds.value[0]
    const file = fileListData.value.files.find((f: any) => f.file_id === fileId)
    if (!file) {
      proxy?.$modal.msgError(t('files.fileNotExists'))
      return
    }
    
    handleShareFile(file)
  }

  const handleToolbarDelete = async () => {
    const totalCount = selectedFileIds.value.length + selectedFolderIds.value.length
    
    if (totalCount === 0) {
      proxy?.$modal.msgWarning(t('files.selectDeleteFilesFirst'))
      return
    }
    
    try {
      await proxy?.$modal.confirm(t('files.confirmDeleteFiles', { count: totalCount }))
      try {
        if (selectedFileIds.value.length > 0) {
          const result = await deleteFiles({ file_ids: selectedFileIds.value })
          if (result.code === 200) {
            proxy?.$modal.msgSuccess(result.message || t('files.deleteSuccess'))
            // 删除成功后刷新用户信息，更新存储空间显示
            await userStore.fetchUserInfo()
          } else {
            proxy?.$modal.msgError(result.message || t('files.deleteFailed'))
          }
        }
        
        if (selectedFolderIds.value.length > 0) {
          proxy?.$modal.msgWarning(t('files.folderDeletePending'))
        }
        
        selectedFileIds.value = []
        selectedFolderIds.value = []
        await loadFileList()
      } catch (error: any) {
        if (error !== 'cancel') {
          proxy?.$modal.msgError(error.message || t('files.deleteFailed'))
        }
      }
    } catch (error: any) {
      if (error !== 'cancel') {
        // 用户取消操作
      }
    }
  }

  const handleSetFilePublic = async (file: FileItem, isPublic: boolean) => {
    try {
      // 如果要设置为公开，检查文件是否加密
      if (isPublic && file.is_enc) {
        proxy?.$modal.msgError(t('files.encryptedFileNotPublic'))
        return
      }

      const result = await setFilePublic({
        file_id: file.file_id,
        public: isPublic
      })

      if (result.code === 200) {
        proxy?.$modal.msgSuccess(isPublic ? t('files.filePublic') : t('files.filePrivate'))
        await loadFileList()
      } else {
        proxy?.$modal.msgError(result.message || t('files.operationFailed'))
      }
    } catch (error: any) {
      proxy?.$log.error('设置文件公开状态失败:', error)
      proxy?.$modal.msgError(error.message || t('files.operationFailed'))
    }
  }

  const handleFileAction = (command: string, file: FileItem): void => {
    switch (command) {
      case 'download':
        handleDownloadFile(file)
        break
      case 'share':
        handleShareFile(file)
        break
      case 'delete':
        handleDeleteFile(file)
        break
      case 'setPublic':
        handleSetFilePublic(file, true)
        break
      case 'setPrivate':
        handleSetFilePublic(file, false)
        break
    }
  }

  return {
    previewVisible,
    previewFile,
    showShareDialog,
    shareForm,
    showDownloadPasswordDialog,
    downloadPasswordForm,
    downloadingFile,
    getFileSize,
    handleShareSuccess,
    handleFilePreview,
    handleShareFile,
    handleDownloadFile,
    confirmDownloadPassword,
    handleDeleteFile,
    handleSetFilePublic,
    handleToolbarDownload,
    handleToolbarShare,
    handleToolbarDelete,
    handleFileAction
  }
}

