import { renameFile } from '@/api/file'
import { renameDir, deleteFolder } from '@/api/folder'
import { useI18n } from '@/composables'
import type { FileItem, FolderItem } from '@/types'

export function useRename(
  selectedFileIds: Ref<string[]>,
  selectedFolderIds: Ref<number[]>,
  loadFileList: () => Promise<void>
) {
  const { proxy } = getCurrentInstance() as ComponentInternalInstance
  const { t } = useI18n()

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
      { required: true, message: t('files.newFileNameRequired'), trigger: 'blur' },
      { min: 1, max: 255, message: t('files.fileNameLength'), trigger: 'blur' },
      {
        pattern: /^[^\\/:*?"<>|]+$/,
        message: t('files.fileNameInvalidChars'),
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
      { required: true, message: t('files.newDirNameRequired'), trigger: 'blur' },
      { min: 1, max: 50, message: t('files.dirNameLength'), trigger: 'blur' },
      {
        pattern: /^[^\\/:*?"<>|]+$/,
        message: t('files.dirNameInvalidChars'),
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
        proxy?.$modal.msgWarning(t('files.sameFileName'))
        return
      }

      renamingFile.value = true
      try {
        const result = await renameFile({
          file_id: renameFileForm.file_id,
          new_file_name: renameFileForm.new_file_name
        })

        if (result.code === 200) {
          proxy?.$modal.msgSuccess(t('files.renameSuccess'))
          showRenameFileDialog.value = false
          selectedFileIds.value = []
          await loadFileList()
        } else {
          proxy?.$modal.msgError(result.message || t('files.renameFailed'))
        }
      } catch (error: any) {
        proxy?.$modal.msgError(error.message || t('files.renameFailed'))
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
        proxy?.$modal.msgWarning(t('files.sameDirName'))
        return
      }

      renamingDir.value = true
      try {
        const result = await renameDir({
          dir_id: renameDirForm.dir_id,
          new_dir_name: renameDirForm.new_dir_name
        })

        if (result.code === 200) {
          proxy?.$modal.msgSuccess(t('files.renameSuccess'))
          showRenameDirDialog.value = false
          selectedFolderIds.value = []
          await loadFileList()
        } else {
          proxy?.$modal.msgError(result.message || t('files.renameFailed'))
        }
      } catch (error: any) {
        proxy?.$modal.msgError(error.message || t('files.renameFailed'))
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

  const handleDeleteDir = async (folder: FolderItem) => {
    try {
      await proxy?.$modal.confirm(t('files.confirmDeleteDir', { dirName: folder.name.replace(/^\//, '') }))

      try {
        const result = await deleteFolder({
          dir_id: folder.id
        })

        if (result.code === 200) {
          proxy?.$modal.msgSuccess(t('files.dirDeleteSuccess'))
          selectedFolderIds.value = []
          await loadFileList()
        } else {
          proxy?.$modal.msgError(result.message || t('files.dirDeleteFailed'))
        }
      } catch (error: any) {
        proxy?.$modal.msgError(error.message || t('files.dirDeleteFailed'))
      }
    } catch (error) {
      // 用户取消删除
    }
  }

  const handleFolderAction = (command: string, folder: FolderItem): void => {
    if (command === 'rename') {
      handleRenameDir(folder)
    } else if (command === 'delete') {
      handleDeleteDir(folder)
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
