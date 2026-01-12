import { moveFile, getVirtualPathTree } from '@/api/file'
import { useI18n } from '@/composables/useI18n'

export function useMoveFile(
  currentPath: Ref<string>,
  selectedFileIds: Ref<string[]>,
  loadFileList: () => Promise<void>
) {
  const { proxy } = getCurrentInstance() as ComponentInternalInstance
  const { t } = useI18n()

  const showMoveDialog = ref(false)
  const moving = ref(false)
  const targetFolderId = ref<string>('')
  const folderTreeData = ref<any[]>([])
  const loadingTree = ref(false)

  const buildFolderTree = async () => {
    loadingTree.value = true
    try {
      const res = await getVirtualPathTree()
      
      if (res.code !== 200 || !res.data) {
        proxy?.$modal.msgError(t('files.getFolderTreeFailed'))
        return
      }
      
      const virtualPaths = res.data as Array<{
        id: number
        path: string
        parent_level: string
        is_dir: boolean
      }>
      
      const pathMap = new Map<string, any>()
      const rootNodes: any[] = []
      
      virtualPaths.forEach(vp => {
        const nodeId = String(vp.id)
        pathMap.set(nodeId, {
          value: nodeId,
          label: vp.path.replace(/^\//, '') || t('files.rootDir'),
          children: [],
          _raw: vp
        })
      })
      
      virtualPaths.forEach(vp => {
        const nodeId = String(vp.id)
        const node = pathMap.get(nodeId)
        
        if (!node) return
        
        if (vp.parent_level && vp.parent_level !== '' && vp.parent_level !== '0') {
          const parentNode = pathMap.get(vp.parent_level)
          if (parentNode) {
            parentNode.children.push(node)
          } else {
            rootNodes.push(node)
          }
        } else {
          rootNodes.push(node)
        }
      })
      
      const cleanEmptyChildren = (nodes: any[]) => {
        nodes.forEach(node => {
          if (node.children && node.children.length === 0) {
            delete node.children
          } else if (node.children) {
            cleanEmptyChildren(node.children)
          }
        })
      }
      cleanEmptyChildren(rootNodes)
      
      folderTreeData.value = rootNodes
    } catch (error: any) {
      proxy?.$modal.msgError(error.message || t('files.getFolderTreeFailed'))
    } finally {
      loadingTree.value = false
    }
  }

  const getFileName = (fileId: string, fileListData: any): string => {
    const file = fileListData.value?.files?.find((f: any) => f.file_id === fileId)
    return file?.file_name || ''
  }

  const handleMoveFile = async () => {
    if (selectedFileIds.value.length === 0) {
      proxy?.$modal.msgWarning(t('files.selectFilesFirst'))
      return
    }
    
    showMoveDialog.value = true
    targetFolderId.value = ''
    await buildFolderTree()
  }

  const handleConfirmMove = async () => {
    if (!targetFolderId.value) {
      proxy?.$modal.msgWarning(t('files.selectTargetDir'))
      return
    }
    
    if (targetFolderId.value === currentPath.value) {
      proxy?.$modal.msgWarning(t('files.sameDir'))
      return
    }
    
    moving.value = true
    try {
      for (const fileId of selectedFileIds.value) {
        const res = await moveFile({
          file_id: fileId,
          source_path: currentPath.value,
          target_path: targetFolderId.value
        })
        
        if (res.code !== 200) {
          proxy?.$modal.msgError(t('files.moveFileFailed') + `: ${res.message}`)
          return
        }
      }
      
      proxy?.$modal.msgSuccess(t('files.moveFilesSuccess', { count: selectedFileIds.value.length }))
      showMoveDialog.value = false
      selectedFileIds.value = []
      loadFileList()
    } catch (error: any) {
      proxy?.$modal.msgError(error.message || t('files.moveFileFailed'))
    } finally {
      moving.value = false
    }
  }

  return {
    showMoveDialog,
    moving,
    targetFolderId,
    folderTreeData,
    loadingTree,
    getFileName,
    handleMoveFile,
    handleConfirmMove
  }
}

