export function useFileSelection() {
  const selectedFolderIds = ref<number[]>([])
  const selectedFileIds = ref<string[]>([])

  const selectedCount = computed(() => selectedFolderIds.value.length + selectedFileIds.value.length)

  const isSelectedFolder = (id: number) => selectedFolderIds.value.includes(id)
  const toggleSelectFolder = (id: number) => {
    const index = selectedFolderIds.value.indexOf(id)
    if (index > -1) {
      selectedFolderIds.value.splice(index, 1)
    } else {
      selectedFolderIds.value.push(id)
    }
  }

  const isSelectedFile = (id: string) => selectedFileIds.value.includes(id)
  const toggleSelectFile = (id: string) => {
    const index = selectedFileIds.value.indexOf(id)
    if (index > -1) {
      selectedFileIds.value.splice(index, 1)
    } else {
      selectedFileIds.value.push(id)
    }
  }

  const handleSelectionChange = (selection: Array<{ isFolder?: boolean; id?: number; file_id?: string }>) => {
    selectedFolderIds.value = selection.filter(s => s.isFolder).map(s => s.id!).filter(id => id !== undefined)
    selectedFileIds.value = selection.filter(s => !s.isFolder).map(s => s.file_id!).filter(id => id !== undefined)
  }

  const clearSelection = () => {
    selectedFolderIds.value = []
    selectedFileIds.value = []
  }

  return {
    selectedFolderIds,
    selectedFileIds,
    selectedCount,
    isSelectedFolder,
    toggleSelectFolder,
    isSelectedFile,
    toggleSelectFile,
    handleSelectionChange,
    clearSelection
  }
}

