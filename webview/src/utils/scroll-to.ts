/**
 * 缓动函数：二次缓入缓出
 * @param t 当前时间
 * @param b 起始值
 * @param c 变化量
 * @param d 持续时间
 */
const easeInOutQuad = (t: number, b: number, c: number, d: number): number => {
  t /= d / 2
  if (t < 1) {
    return (c / 2) * t * t + b
  }
  t--
  return (-c / 2) * (t * (t - 2) - 1) + b
}

/**
 * requestAnimationFrame 兼容性处理
 */
const requestAnimFrame = (function () {
  return (
    window.requestAnimationFrame ||
    (window as any).webkitRequestAnimationFrame ||
    (window as any).mozRequestAnimationFrame ||
    function (callback: FrameRequestCallback) {
      window.setTimeout(callback, 1000 / 60)
    }
  )
})()

/**
 * 设置滚动位置（兼容多种浏览器）
 * @param amount 滚动位置
 */
const move = (amount: number): void => {
  document.documentElement.scrollTop = amount
  const bodyParent = document.body.parentNode as HTMLElement
  if (bodyParent) {
    bodyParent.scrollTop = amount
  }
  document.body.scrollTop = amount
}

/**
 * 获取当前滚动位置
 */
const position = (): number => {
  return (
    document.documentElement.scrollTop ||
    ((document.body.parentNode as HTMLElement)?.scrollTop || 0) ||
    document.body.scrollTop ||
    0
  )
}

/**
 * 平滑滚动到指定位置
 * @param to 目标滚动位置
 * @param duration 滚动持续时间（毫秒），默认 500
 * @param callback 滚动完成后的回调函数（可选）
 */
export const scrollTo = (to: number, duration: number = 500, callback?: () => void): void => {
  const start = position()
  const change = to - start
  const increment = 20
  let currentTime = 0

  const animateScroll = (): void => {
    // 增加时间
    currentTime += increment
    // 使用二次缓入缓出函数计算当前值
    const val = easeInOutQuad(currentTime, start, change, duration)
    // 移动滚动位置
    move(val)
    // 如果动画未完成，继续动画
    if (currentTime < duration) {
      requestAnimFrame(animateScroll)
    } else {
      // 动画完成，执行回调
      if (callback && typeof callback === 'function') {
        callback()
      }
    }
  }

  animateScroll()
}

