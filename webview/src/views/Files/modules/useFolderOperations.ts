import { createFolder } from '@/api/folder'
import { useI18n } from '@/composables/useI18n'

export function useFolderOperations(
  currentPath: Ref<string>,
  loadFileList: () => Promise<void>
) {
  const { proxy } = getCurrentInstance() as ComponentInternalInstance
  const { t } = useI18n()

  const showNewFolderDialog = ref(false)
  const creating = ref(false)
  const folderFormRef = ref<FormInstance>()
  const folderForm = reactive({
    dir_path: ''
  })

  const folderRules: FormRules = {
    dir_path: [
      { required: true, message: t('files.folderNameRequired'), trigger: 'blur' },
      { min: 1, max: 50, message: t('files.folderNameLength'), trigger: 'blur' },
      { 
        pattern: /^[^\\/:*?"<>|]+$/, 
        message: t('files.folderNameInvalidChars'), 
        trigger: 'blur' 
      }
    ]
  }

  const handleNewFolder = () => {
    showNewFolderDialog.value = true
    folderForm.dir_path = ''
  }

  const handleDialogClose = () => {
    folderFormRef.value?.resetFields()
  }

  const handleCreateFolder = async () => {
    if (!folderFormRef.value) return
    
    await folderFormRef.value.validate(async (valid: boolean) => {
      if (valid) {
        creating.value = true
        try {
          const res = await createFolder({
            parent_level: currentPath.value,
            dir_path: folderForm.dir_path
          })
          
          if (res.code === 200) {
            proxy?.$modal.msgSuccess(t('files.createFolderSuccess'))
            showNewFolderDialog.value = false
            folderForm.dir_path = ''
            loadFileList()
          } else {
            proxy?.$modal.msgError(res.message || t('files.createFolderFailed'))
          }
        } catch (error: any) {
          proxy?.$modal.msgError(error.message || t('files.createFolderFailed'))
        } finally {
          creating.value = false
        }
      }
    })
  }

  return {
    showNewFolderDialog,
    creating,
    folderFormRef,
    folderForm,
    folderRules,
    handleNewFolder,
    handleDialogClose,
    handleCreateFolder
  }
}

