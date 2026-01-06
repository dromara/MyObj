import { deleteFiles, setFilePublic } from '@/api/file'
import { createPackage, getPackageProgress, downloadPackage } from '@/api/package'
import { useFileDownload } from '@/composables/useFileDownload'
import { useUserStore } from '@/stores/user'
import type { FileItem, FileListResponse } from '@/types'

export function useFileOperations(
  fileListData: Ref<FileListResponse>,
  selectedFileIds: Ref<string[]>,
  selectedFolderIds: Ref<number[]>,
  loadFileList: () => Promise<void>
) {
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
      await proxy?.$modal.confirm(`确定要删除 "${file.file_name}" 吗？删除后将移动到回收站。`)
      try {
        const result = await deleteFiles({ file_ids: [file.file_id] })
        if (result.code === 200) {
          proxy?.$modal.msgSuccess('删除成功')
          selectedFileIds.value = []
          selectedFolderIds.value = []
          await loadFileList()
          // 删除成功后刷新用户信息，更新存储空间显示
          await userStore.fetchUserInfo()
        } else {
          proxy?.$modal.msgError(result.message || '删除失败')
        }
      } catch (error: any) {
        if (error !== 'cancel') {
          proxy?.$modal.msgError(error.message || '删除失败')
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
      proxy?.$modal.msgWarning('请先选择要下载的文件')
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
      proxy?.$modal.msgWarning('请先选择要下载的文件')
      return
    }

    try {
      // 创建打包任务
      proxy?.$modal.loading('正在创建压缩包，请稍候...')
      
      const res = await createPackage({
        file_ids: selectedFileIds.value,
        package_name: `files_${Date.now()}.zip`
      })
      
      if (res.code === 200 && res.data) {
        const packageId = res.data.package_id
        
        // 如果状态是 ready，直接下载
        if (res.data.status === 'ready') {
          const downloadUrl = downloadPackage(packageId)
          window.open(downloadUrl, '_blank')
          proxy?.$modal.closeLoading()
          proxy?.$modal.msgSuccess('开始下载')
          return
        }
        
        // 如果状态是 creating，轮询进度
        if (res.data.status === 'creating') {
          await pollPackageProgress(packageId)
        }
      } else {
        proxy?.$modal.closeLoading()
        if (res.code === 404 || res.message?.includes('404')) {
          proxy?.$modal.msg('打包下载功能开发中')
        } else {
          proxy?.$modal.msgError(res.message || '创建打包任务失败')
        }
      }
    } catch (error: any) {
      proxy?.$modal.closeLoading()
      if (error.response?.status === 404 || error.message?.includes('404')) {
        proxy?.$modal.msg('打包下载功能开发中')
      } else {
        proxy?.$modal.msgError(error.message || '打包下载失败')
      }
      proxy?.$log?.error(error)
    }
  }

  // 轮询打包进度
  const pollPackageProgress = async (packageId: string) => {
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
            window.open(downloadUrl, '_blank')
            proxy?.$modal.msgSuccess('压缩包已创建，开始下载')
            return
          }
          
          if (status === 'failed') {
            proxy?.$modal.closeLoading()
            proxy?.$modal.msgError(res.data.error_msg || '打包失败')
            return
          }
          
          // 更新进度提示
          if (progress < 100) {
            proxy?.$modal.loading(`正在打包... ${progress}%`)
          }
          
          // 继续轮询
          if (attempts < maxAttempts && status === 'creating') {
            attempts++
            setTimeout(poll, 5000) // 每5秒轮询一次
          } else if (attempts >= maxAttempts) {
            proxy?.$modal.closeLoading()
            proxy?.$modal.msgError('打包超时，请稍后重试')
          }
        } else {
          proxy?.$modal.closeLoading()
          if (res.code === 404 || res.message?.includes('404')) {
            proxy?.$modal.msg('打包下载功能开发中')
          } else {
            proxy?.$modal.msgError(res.message || '获取打包进度失败')
          }
        }
      } catch (error: any) {
        proxy?.$modal.closeLoading()
        if (error.response?.status === 404 || error.message?.includes('404')) {
          proxy?.$modal.msg('打包下载功能开发中')
        } else {
          proxy?.$modal.msgError('获取打包进度失败')
        }
        proxy?.$log?.error(error)
      }
    }
    
    poll()
  }

  const handleToolbarShare = () => {
    if (selectedFileIds.value.length === 0) {
      proxy?.$modal.msgWarning('请先选择要分享的文件')
      return
    }
    if (selectedFileIds.value.length > 1) {
      proxy?.$modal.msgWarning('一次只能分享一个文件')
      return
    }
    
    const fileId = selectedFileIds.value[0]
    const file = fileListData.value.files.find((f: any) => f.file_id === fileId)
    if (!file) {
      proxy?.$modal.msgError('文件不存在')
      return
    }
    
    handleShareFile(file)
  }

  const handleToolbarDelete = async () => {
    const totalCount = selectedFileIds.value.length + selectedFolderIds.value.length
    
    if (totalCount === 0) {
      proxy?.$modal.msgWarning('请先选择要删除的文件')
      return
    }
    
    try {
      await proxy?.$modal.confirm(`确定要删除 ${totalCount} 个文件吗？删除后将移动到回收站。`)
      try {
        if (selectedFileIds.value.length > 0) {
          const result = await deleteFiles({ file_ids: selectedFileIds.value })
          if (result.code === 200) {
            proxy?.$modal.msgSuccess(result.message || '删除成功')
            // 删除成功后刷新用户信息，更新存储空间显示
            await userStore.fetchUserInfo()
          } else {
            proxy?.$modal.msgError(result.message || '删除失败')
          }
        }
        
        if (selectedFolderIds.value.length > 0) {
          proxy?.$modal.msgWarning('文件夹删除功能待开发')
        }
        
        selectedFileIds.value = []
        selectedFolderIds.value = []
        await loadFileList()
      } catch (error: any) {
        if (error !== 'cancel') {
          proxy?.$modal.msgError(error.message || '删除失败')
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
        proxy?.$modal.msgError('加密文件不能设置为公开')
        return
      }

      const result = await setFilePublic({
        file_id: file.file_id,
        public: isPublic
      })

      if (result.code === 200) {
        proxy?.$modal.msgSuccess(isPublic ? '文件已公开' : '文件已取消公开')
        await loadFileList()
      } else {
        proxy?.$modal.msgError(result.message || '操作失败')
      }
    } catch (error: any) {
      proxy?.$log.error('设置文件公开状态失败:', error)
      proxy?.$modal.msgError(error.message || '操作失败')
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

