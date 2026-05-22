/**
 * 平滑滚动到指定位置
 */
const easeInOutQuad = (t: number, b: number, c: number, d: number): number => {
  t /= d / 2
  if (t < 1) {
    return (c / 2) * t * t + b
  }
  t--
  return (-c / 2) * (t * (t - 2) - 1) + b
}

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

const move = (amount: number): void => {
  document.documentElement.scrollTop = amount
  const bodyParent = document.body.parentNode as HTMLElement
  if (bodyParent) {
    bodyParent.scrollTop = amount
  }
  document.body.scrollTop = amount
}

const position = (): number => {
  return (
    document.documentElement.scrollTop ||
    (document.body.parentNode as HTMLElement)?.scrollTop ||
    0 ||
    document.body.scrollTop ||
    0
  )
}

export const scrollTo = (to: number, duration: number = 500, callback?: () => void): void => {
  const start = position()
  const change = to - start
  const increment = 20
  let currentTime = 0

  const animateScroll = (): void => {
    currentTime += increment
    const val = easeInOutQuad(currentTime, start, change, duration)
    move(val)
    if (currentTime < duration) {
      requestAnimFrame(animateScroll)
    } else if (callback && typeof callback === 'function') {
      callback()
    }
  }

  animateScroll()
}
