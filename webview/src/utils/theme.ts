/**
 * Toggle html class
 * 
 * @param className Class name to toggle
 * @returns Object with add and remove functions
 */
export function toggleHtmlClass(className: string) {
  function add() {
    document.documentElement.classList.add(className)
  }

  function remove() {
    document.documentElement.classList.remove(className)
  }

  return {
    add,
    remove
  }
}

/**
 * Toggle CSS dark mode
 * 
 * @param darkMode Is dark mode
 */
export function toggleCssDarkMode(darkMode = false) {
  const { add, remove } = toggleHtmlClass('dark')

  if (darkMode) {
    add()
  } else {
    remove()
  }
}
