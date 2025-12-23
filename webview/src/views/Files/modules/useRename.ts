import { ref, reactive, getCurrentInstance, ComponentInternalInstance, type Ref } from 'vue'
import { renameFile } from '@/api/file'
import { renameDir } from '@/api/folder'
import type { FileItem, FolderItem } from '@/types'

export function useRename(
  selectedFileIds: Ref<string[]>,
  selectedFolderIds: Ref<number[]>,
  loadFileList: () => Promise<void>
) {
  const { proxy } = getCurrentInstance() as ComponentInternalInstance

  // 文件重命名
  const showRenameFileDialog = ref(false)
  const renamingFile = ref(false)
  const renameFileFormRef = ref<FormInstance>()
  const renameFileForm = reactive({
    file_id: '',
    old_file_name: '',
    new_file_name: ''
  })

  const renameFileRules: FormRules = {
    new_file_name: [
      { required: true, message: '请输入新文件名', trigger: 'blur' },
      { min: 1, max: 255, message: '文件名长度在1-255个字符', trigger: 'blur' },
      { 
        pattern: /^[^\\/:*?"<>|]+$/, 
        message: '文件名不能包含特殊字符 \\ / : * ? " < > |', 
        trigger: 'blur' 
      }
    ]
  }

  // 目录重命名
  const showRenameDirDialog = ref(false)
  const renamingDir = ref(false)
  const renameDirFormRef = ref<FormInstance>()
  const renameDirForm = reactive({
    dir_id: 0,
    old_dir_name: '',
    new_dir_name: ''
  })

  const renameDirRules: FormRules = {
    new_dir_name: [
      { required: true, message: '请输入新目录名', trigger: 'blur' },
      { min: 1, max: 50, message: '目录名长度在1-50个字符', trigger: 'blur' },
      { 
        pattern: /^[^\\/:*?"<>|]+$/, 
        message: '目录名不能包含特殊字符 \\ / : * ? " < > |', 
        trigger: 'blur' 
      }
    ]
  }

  const handleRenameFile = (file: FileItem) => {
    renameFileForm.file_id = file.file_id
    renameFileForm.old_file_name = file.file_name
    renameFileForm.new_file_name = file.file_name
    showRenameFileDialog.value = true
  }

  const handleConfirmRenameFile = async () => {
    if (!renameFileFormRef.value) return
    
    try {
      await renameFileFormRef.value.validate()
      
      if (renameFileForm.new_file_name === renameFileForm.old_file_name) {
        proxy?.$modal.msgWarning('新文件名与原文件名相同')
        return
      }
      
      renamingFile.value = true
      try {
        const result = await renameFile({
          file_id: renameFileForm.file_id,
          new_file_name: renameFileForm.new_file_name
        })
        
        if (result.code === 200) {
          proxy?.$modal.msgSuccess('重命名成功')
          showRenameFileDialog.value = false
          selectedFileIds.value = []
          await loadFileList()
        } else {
          proxy?.$modal.msgError(result.message || '重命名失败')
        }
      } catch (error: any) {
        proxy?.$modal.msgError(error.message || '重命名失败')
      } finally {
        renamingFile.value = false
      }
    } catch (error) {
      // 表单验证失败
    }
  }

  const handleRenameFileDialogClose = () => {
    renameFileFormRef.value?.resetFields()
    renameFileForm.file_id = ''
    renameFileForm.old_file_name = ''
    renameFileForm.new_file_name = ''
  }

  const handleRenameDir = (folder: FolderItem) => {
    renameDirForm.dir_id = folder.id
    renameDirForm.old_dir_name = folder.name.replace(/^\//, '')
    renameDirForm.new_dir_name = folder.name.replace(/^\//, '')
    showRenameDirDialog.value = true
  }

  const handleConfirmRenameDir = async () => {
    if (!renameDirFormRef.value) return
    
    try {
      await renameDirFormRef.value.validate()
      
      if (renameDirForm.new_dir_name === renameDirForm.old_dir_name) {
        proxy?.$modal.msgWarning('新目录名与原目录名相同')
        return
      }
      
      renamingDir.value = true
      try {
        const result = await renameDir({
          dir_id: renameDirForm.dir_id,
          new_dir_name: renameDirForm.new_dir_name
        })
        
        if (result.code === 200) {
          proxy?.$modal.msgSuccess('重命名成功')
          showRenameDirDialog.value = false
          selectedFolderIds.value = []
          await loadFileList()
        } else {
          proxy?.$modal.msgError(result.message || '重命名失败')
        }
      } catch (error: any) {
        proxy?.$modal.msgError(error.message || '重命名失败')
      } finally {
        renamingDir.value = false
      }
    } catch (error) {
      // 表单验证失败
    }
  }

  const handleRenameDirDialogClose = () => {
    renameDirFormRef.value?.resetFields()
    renameDirForm.dir_id = 0
    renameDirForm.old_dir_name = ''
    renameDirForm.new_dir_name = ''
  }

  const handleFileAction = (command: string, file: FileItem): void => {
    if (command === 'rename') {
      handleRenameFile(file)
    }
  }

  const handleFolderAction = (command: string, folder: FolderItem): void => {
    if (command === 'rename') {
      handleRenameDir(folder)
    }
  }

  return {
    showRenameFileDialog,
    renamingFile,
    renameFileFormRef,
    renameFileForm,
    renameFileRules,
    showRenameDirDialog,
    renamingDir,
    renameDirFormRef,
    renameDirForm,
    renameDirRules,
    handleRenameFile,
    handleConfirmRenameFile,
    handleRenameFileDialogClose,
    handleRenameDir,
    handleConfirmRenameDir,
    handleRenameDirDialogClose,
    handleFileAction,
    handleFolderAction
  }
}

