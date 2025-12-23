import { ref, reactive, getCurrentInstance, ComponentInternalInstance, type Ref } from 'vue'
import { useRouter } from 'vue-router'
import { deleteFiles } from '@/api/file'
import { createLocalFileDownload, getDownloadTaskList, getLocalFileDownloadUrl } from '@/api/download'
import type { FileItem, FileListResponse } from '@/types'

export function useFileOperations(
  fileListData: Ref<FileListResponse>,
  selectedFileIds: Ref<string[]>,
  selectedFolderIds: Ref<number[]>,
  loadFileList: () => Promise<void>
) {
  const { proxy } = getCurrentInstance() as ComponentInternalInstance
  const router = useRouter()

  const previewVisible = ref(false)
  const previewFile = ref<FileItem | null>(null)
  const showShareDialog = ref(false)
  const shareForm = reactive({
    file_id: '',
    file_name: '',
    file_size: 0
  })
  const showDownloadPasswordDialog = ref(false)
  const downloadPasswordForm = reactive({
    file_id: '',
    file_name: '',
    file_password: ''
  })
  const downloadingFile = ref(false)

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
    if (file.is_enc) {
      downloadPasswordForm.file_id = file.file_id
      downloadPasswordForm.file_name = file.file_name
      downloadPasswordForm.file_password = ''
      showDownloadPasswordDialog.value = true
    } else {
      await executeDownload(file.file_id, '')
    }
  }

  const executeDownload = async (fileId: string, password: string) => {
    try {
      downloadingFile.value = true
      const res = await createLocalFileDownload({
        file_id: fileId,
        file_password: password
      })
      
      if (res.code === 200) {
        const taskId = res.data?.task_id
        if (!taskId) {
          proxy?.$modal.msgError('任务创建失败')
          downloadingFile.value = false
          return
        }
        
        proxy?.$modal.msgSuccess('准备下载中，请稍候...')
        showDownloadPasswordDialog.value = false
        
        let retryCount = 0
        const maxRetries = 30
        
        const checkTaskStatus = async () => {
          try {
            const taskRes = await getDownloadTaskList({ page: 1, pageSize: 100, state: -1 })
            if (taskRes.code === 200 && taskRes.data) {
              const task = taskRes.data.tasks?.find((t: any) => t.id === taskId)
              
              if (!task) {
                proxy?.$log.error('未找到任务:', taskId)
                retryCount++
                if (retryCount < maxRetries) {
                  setTimeout(checkTaskStatus, 1000)
                } else {
                  proxy?.$modal.msgError('未找到下载任务')
                  downloadingFile.value = false
                }
                return
              }
              
              proxy?.$log.debug('任务状态:', task.state, '任务信息:', task)
              
              if (task.state === 3) {
                router.push({
                  path: '/tasks',
                  query: { tab: 'download' }
                })
                
                const token = proxy?.$cache.local.get('token')
                const downloadUrl = getLocalFileDownloadUrl(taskId)
                
                proxy?.$log.debug('开始下载文件:', downloadUrl)
                
                try {
                  const response = await fetch(downloadUrl, {
                    method: 'GET',
                    headers: {
                      'Authorization': token ? `Bearer ${token}` : ''
                    }
                  })
                  
                  if (!response.ok) {
                    throw new Error('下载失败: ' + response.status)
                  }
                  
                  const blob = await response.blob()
                  proxy?.$log.debug('下载完成，文件大小:', blob.size)
                  
                  const url = window.URL.createObjectURL(blob)
                  const link = document.createElement('a')
                  link.href = url
                  link.download = task.file_name || 'download'
                  link.style.display = 'none'
                  document.body.appendChild(link)
                  link.click()
                  document.body.removeChild(link)
                  window.URL.revokeObjectURL(url)
                  
                  proxy?.$modal.msgSuccess('下载完成')
                } catch (error: any) {
                  proxy?.$log.error('下载文件失败:', error)
                  proxy?.$modal.msgError('下载失败: ' + (error.message || '未知错误'))
                }
                
                downloadingFile.value = false
                return
              } else if (task.state === 4) {
                proxy?.$log.error('任务失败:', task.error_msg)
                proxy?.$modal.msgError(task.error_msg || '下载准备失败')
                downloadingFile.value = false
                return
              }
              
              retryCount++
              if (retryCount < maxRetries) {
                setTimeout(checkTaskStatus, 1000)
              } else {
                proxy?.$modal.msgWarning('准备超时，请到任务中心查看')
                downloadingFile.value = false
              }
            } else {
              proxy?.$log.error('获取任务列表失败:', taskRes)
              retryCount++
              if (retryCount < maxRetries) {
                setTimeout(checkTaskStatus, 1000)
              } else {
                proxy?.$modal.msgError('获取任务状态失败')
                downloadingFile.value = false
              }
            }
          } catch (error: any) {
            proxy?.$log.error('查询任务状态异常:', error)
            retryCount++
            if (retryCount < maxRetries) {
              setTimeout(checkTaskStatus, 1000)
            } else {
              proxy?.$modal.msgError('查询任务状态失败')
              downloadingFile.value = false
            }
          }
        }
        
        setTimeout(checkTaskStatus, 1000)
      } else {
        proxy?.$modal.msgError(res.message || '创建下载任务失败')
        downloadingFile.value = false
      }
    } catch (error: any) {
      proxy?.$log.error('创建下载任务异常:', error)
      proxy?.$modal.msgError(error.message || '创建下载任务失败')
      downloadingFile.value = false
    }
  }

  const confirmDownloadPassword = async () => {
    if (!downloadPasswordForm.file_password) {
      proxy?.$modal.msgWarning('请输入文件密码')
      return
    }
    await executeDownload(downloadPasswordForm.file_id, downloadPasswordForm.file_password)
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
      proxy?.$modal.msg('批量下载功能开发中')
    }
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
    handleToolbarDownload,
    handleToolbarShare,
    handleToolbarDelete,
    handleFileAction
  }
}

