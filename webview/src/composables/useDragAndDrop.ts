import type { ComponentInternalInstance } from 'vue'

interface DragState {
  isDragging: boolean
  draggedElement: HTMLElement | null
  dropTarget: HTMLElement | null
}

/**
 * 拖拽功能 Composable
 */
export function useDragAndDrop() {
  const { proxy } = getCurrentInstance() as ComponentInternalInstance
  const dragState = ref<DragState>({
    isDragging: false,
    draggedElement: null,
    dropTarget: null
  })

  // 使元素可拖拽
  const makeDraggable = (element: HTMLElement, data: any) => {
    element.draggable = true
    element.style.cursor = 'grab'

    element.addEventListener('dragstart', (e) => {
      dragState.value.isDragging = true
      dragState.value.draggedElement = element
      
      // 设置拖拽数据
      if (e.dataTransfer) {
        e.dataTransfer.effectAllowed = 'move'
        e.dataTransfer.setData('application/json', JSON.stringify(data))
        // 设置拖拽预览图
        e.dataTransfer.setDragImage(element, 0, 0)
      }

      element.style.cursor = 'grabbing'
      element.style.opacity = '0.5'
    })

    element.addEventListener('dragend', () => {
      dragState.value.isDragging = false
      dragState.value.draggedElement = null
      dragState.value.dropTarget = null
      element.style.cursor = 'grab'
      element.style.opacity = '1'
    })
  }

  // 使元素可放置
  const makeDroppable = (
    element: HTMLElement,
    onDrop: (data: any, e: DragEvent) => void
  ) => {
    element.addEventListener('dragover', (e) => {
      e.preventDefault()
      if (e.dataTransfer) {
        e.dataTransfer.dropEffect = 'move'
      }
      dragState.value.dropTarget = element
      element.classList.add('drag-over')
    })

    element.addEventListener('dragleave', () => {
      dragState.value.dropTarget = null
      element.classList.remove('drag-over')
    })

    element.addEventListener('drop', (e) => {
      e.preventDefault()
      dragState.value.dropTarget = null
      element.classList.remove('drag-over')

      if (e.dataTransfer) {
        try {
          const data = JSON.parse(e.dataTransfer.getData('application/json'))
          onDrop(data, e)
        } catch (error) {
          proxy?.$log.error('解析拖拽数据失败:', error)
        }
      }
    })
  }

  return {
    dragState: readonly(dragState),
    makeDraggable,
    makeDroppable
  }
}
