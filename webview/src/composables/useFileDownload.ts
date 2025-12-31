/**
 * 文件下载 Composable
 * 统一处理文件下载逻辑，支持加密文件
 */
import { createLocalFileDownload, getDownloadTaskList, getLocalFileDownloadUrl } from '@/api/download'
import type { FileItem } from '@/types'

export interface DownloadPasswordForm {
  file_id: string
  file_name: string
  file_password: string
}

export function useFileDownload(options?: {
  onTaskReady?: () => void // 任务准备完成时的回调（可选，用于跳转等）
}) {
  const { proxy } = getCurrentInstance() as ComponentInternalInstance

  const showDownloadPasswordDialog = ref(false)
  const downloadPasswordForm = reactive<DownloadPasswordForm>({
    file_id: '',
    file_name: '',
    file_password: ''
  })
  const downloadingFile = ref(false)

  /**
   * 处理文件下载
   * @param file 文件信息
   */
  const handleDownload = async (file: FileItem) => {
    // 如果是加密文件，需要输入密码
    if (file.is_enc) {
      downloadPasswordForm.file_id = file.file_id
      downloadPasswordForm.file_name = file.file_name
      downloadPasswordForm.file_password = ''
      showDownloadPasswordDialog.value = true
    } else {
      await executeDownload(file.file_id, '')
    }
  }

  /**
   * 执行下载（任务式下载）
   * @param fileId 文件ID
   * @param password 文件密码（加密文件必需）
   */
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
                // 任务完成，触发回调（可选）
                if (options?.onTaskReady) {
                  options.onTaskReady()
                }
                
                // 开始下载文件 - 直接使用浏览器下载
                const downloadUrl = getLocalFileDownloadUrl(taskId)
                
                proxy?.$log.debug('开始下载文件:', downloadUrl)
                
                try {
                  // 直接创建下载链接，让浏览器处理下载
                  const link = document.createElement('a')
                  link.href = downloadUrl
                  link.download = task.file_name || 'download'
                  link.style.display = 'none'
                  document.body.appendChild(link)
                  link.click()
                  document.body.removeChild(link)
                  
                  proxy?.$modal.msgSuccess('下载已开始')
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

  /**
   * 确认下载密码
   */
  const confirmDownloadPassword = async () => {
    if (!downloadPasswordForm.file_password) {
      proxy?.$modal.msgWarning('请输入文件密码')
      return
    }
    await executeDownload(downloadPasswordForm.file_id, downloadPasswordForm.file_password)
  }

  return {
    showDownloadPasswordDialog,
    downloadPasswordForm,
    downloadingFile,
    handleDownload,
    confirmDownloadPassword
  }
}

