import { createFolder } from '@/api/folder'

export function useFolderOperations(
  currentPath: Ref<string>,
  loadFileList: () => Promise<void>
) {
  const { proxy } = getCurrentInstance() as ComponentInternalInstance

  const showNewFolderDialog = ref(false)
  const creating = ref(false)
  const folderFormRef = ref<FormInstance>()
  const folderForm = reactive({
    dir_path: ''
  })

  const folderRules: FormRules = {
    dir_path: [
      { required: true, message: '请输入文件夹名称', trigger: 'blur' },
      { min: 1, max: 50, message: '文件夹名称长度在1-50个字符', trigger: 'blur' },
      { 
        pattern: /^[^\\/:*?"<>|]+$/, 
        message: '文件夹名称不能包含特殊字符 \\ / : * ? " < > |', 
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
            proxy?.$modal.msgSuccess('文件夹创建成功')
            showNewFolderDialog.value = false
            folderForm.dir_path = ''
            loadFileList()
          } else {
            proxy?.$modal.msgError(res.message || '创建文件夹失败')
          }
        } catch (error: any) {
          proxy?.$modal.msgError(error.message || '创建文件夹失败')
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

