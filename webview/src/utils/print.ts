/**
 * 打印工具函数
 * 支持图片、PDF、文本等文件的打印
 */

/**
 * 打印选项
 */
export interface PrintOptions {
  /** 打印标题 */
  title?: string
  /** 是否显示页眉页脚 */
  showHeaderFooter?: boolean
  /** 打印方向：portrait(纵向) | landscape(横向) */
  orientation?: 'portrait' | 'landscape'
  /** 页边距（单位：mm） */
  margin?: {
    top?: number
    right?: number
    bottom?: number
    left?: number
  }
  /** 打印背景图形 */
  printBackground?: boolean
}

/**
 * 打印图片
 * @param imageUrl 图片URL（可以是blob URL、data URL或普通URL）
 * @param options 打印选项
 */
export function printImage(imageUrl: string, options?: PrintOptions): Promise<void> {
  return new Promise((resolve, reject) => {
    try {
      const printWindow = window.open('', '_blank')
      if (!printWindow) {
        reject(new Error('无法打开打印窗口，请检查浏览器弹窗设置'))
        return
      }

      const title = options?.title || '图片打印'

      printWindow.document.write(`
        <!DOCTYPE html>
        <html>
        <head>
          <title>${title}</title>
          <meta charset="UTF-8">
          <style>
            * {
              box-sizing: border-box;
            }
            @media print {
              @page {
                size: ${options?.orientation === 'landscape' ? 'A4 landscape' : 'A4'};
                margin: ${options?.margin?.top || 10}mm ${options?.margin?.right || 10}mm ${options?.margin?.bottom || 10}mm ${options?.margin?.left || 10}mm;
              }
              body {
                margin: 0;
                padding: 0;
                display: flex;
                flex-direction: column;
                justify-content: center;
                align-items: center;
                min-height: 100vh;
                background: white;
              }
              .print-header {
                display: none;
              }
              .image-container {
                display: flex;
                justify-content: center;
                align-items: center;
                width: 100%;
                height: 100%;
              }
              img {
                width: auto;
                height: auto;
                max-width: 100%;
                max-height: 100%;
                object-fit: contain;
                background: white;
              }
            }
            @media screen {
              body {
                margin: 0;
                padding: 0;
                display: flex;
                flex-direction: column;
                justify-content: center;
                align-items: center;
                min-height: 100vh;
                background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
                font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Microsoft YaHei', Arial, sans-serif;
              }
              .print-header {
                width: 100%;
                padding: 20px;
                background: rgba(255, 255, 255, 0.95);
                backdrop-filter: blur(10px);
                box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
                text-align: center;
                border-bottom: 1px solid rgba(0, 0, 0, 0.05);
              }
              .print-header h2 {
                margin: 0;
                font-size: 18px;
                font-weight: 600;
                color: #333;
                letter-spacing: 0.5px;
              }
              .print-header p {
                margin: 8px 0 0 0;
                font-size: 13px;
                color: #666;
              }
              .image-container {
                flex: 1;
                display: flex;
                justify-content: center;
                align-items: center;
                width: 100%;
                padding: 30px;
                overflow: auto;
              }
              img {
                width: auto;
                height: auto;
                max-width: 100%;
                max-height: calc(100vh - 120px);
                object-fit: contain;
                background: white;
                border-radius: 8px;
                box-shadow: 0 8px 32px rgba(0, 0, 0, 0.15);
                transition: transform 0.3s ease;
              }
              img:hover {
                transform: scale(1.02);
              }
            }
          </style>
        </head>
        <body>
          <div class="print-header">
            <h2>${escapeHtml(title)}</h2>
            <p>准备打印中，请稍候...</p>
          </div>
          <div class="image-container">
            <img src="${imageUrl}" alt="${escapeHtml(title)}" onload="setTimeout(function() { window.print(); window.onafterprint = function() { window.close(); } }, 300);" onerror="alert('图片加载失败'); window.close();" />
          </div>
        </body>
        </html>
      `)
      printWindow.document.close()

      // 监听打印完成
      printWindow.addEventListener('afterprint', () => {
        printWindow.close()
        resolve()
      })

      // 如果打印窗口被关闭，也resolve
      const checkClosed = setInterval(() => {
        if (printWindow.closed) {
          clearInterval(checkClosed)
          resolve()
        }
      }, 100)
    } catch (error) {
      reject(error)
    }
  })
}

/**
 * 打印PDF
 * @param pdfUrl PDF URL
 * @param options 打印选项
 */
export function printPDF(pdfUrl: string, options?: PrintOptions): Promise<void> {
  return new Promise((resolve, reject) => {
    try {
      const printWindow = window.open('', '_blank')
      if (!printWindow) {
        reject(new Error('无法打开打印窗口，请检查浏览器弹窗设置'))
        return
      }

      const title = options?.title || 'PDF打印'
      
      // 在PDF URL后添加参数以隐藏工具栏和导航栏，并设置100%缩放
      // zoom=100 表示100%缩放，确保PDF在打印预览中完整显示
      let finalPdfUrl = pdfUrl
      const pdfParams = 'toolbar=0&navpanes=0&scrollbar=0&zoom=100'
      if (pdfUrl.includes('#')) {
        finalPdfUrl = pdfUrl.split('#')[0] + '#' + pdfParams
      } else {
        finalPdfUrl = pdfUrl + '#' + pdfParams
      }

      printWindow.document.write(`
        <!DOCTYPE html>
        <html>
        <head>
          <title>${title}</title>
          <meta charset="UTF-8">
          <style>
            * {
              margin: 0;
              padding: 0;
              box-sizing: border-box;
            }
            @media print {
              @page {
                size: ${options?.orientation === 'landscape' ? 'A4 landscape' : 'A4'};
                margin: ${options?.margin?.top || 0}mm ${options?.margin?.right || 0}mm ${options?.margin?.bottom || 0}mm ${options?.margin?.left || 0}mm;
              }
              html {
                width: 250%;
                height: 250%;
                margin: 0;
                padding: 0;
                transform: scale(0.4);
                transform-origin: top left;
                -webkit-print-color-adjust: exact;
                print-color-adjust: exact;
              }
              body {
                margin: 0 !important;
                padding: 0 !important;
                width: 100%;
                height: 100%;
                overflow: hidden;
              }
              iframe {
                position: fixed !important;
                top: 0 !important;
                left: 0 !important;
                width: 100% !important;
                height: 100% !important;
                border: none !important;
                margin: 0 !important;
                padding: 0 !important;
              }
            }
            @media screen {
              body {
                margin: 0;
                padding: 0;
                overflow: hidden;
                background: #f5f5f5;
              }
              iframe {
                width: 100%;
                height: 100vh;
                border: none;
                display: block;
              }
            }
          </style>
        </head>
        <body>
          <iframe id="pdfFrame" src="${finalPdfUrl}" style="width: 100%; height: 100vh; border: none;" onerror="alert('PDF加载失败'); window.close();"></iframe>
          <script>
            (function() {
              var iframe = document.getElementById('pdfFrame');
              var printTimeout;
              
              function tryPrint() {
                try {
                  // 尝试通过iframe访问PDF并设置缩放（可能受跨域限制）
                  try {
                    var iframeDoc = iframe.contentDocument || iframe.contentWindow.document;
                    // 如果PDF查看器支持，尝试设置缩放
                    if (iframeDoc && iframeDoc.querySelector) {
                      // 某些PDF查看器可能支持这种方式
                    }
                  } catch(e) {
                    // 跨域限制，无法访问iframe内容
                  }
                  
                  // 等待PDF完全加载后再打印
                  printTimeout = setTimeout(function() {
                    window.print();
                    window.onafterprint = function() {
                      window.close();
                    };
                  }, 1500);
                } catch(e) {
                  console.error('打印错误:', e);
                }
              }
              
              // 监听iframe加载完成
              iframe.addEventListener('load', function() {
                tryPrint();
              });
              
              // 备用方案：窗口加载完成后也尝试打印
              window.addEventListener('load', function() {
                if (!printTimeout) {
                  tryPrint();
                }
              });
            })();
          </script>
        </body>
        </html>
      `)
      printWindow.document.close()

      // 监听打印完成
      printWindow.addEventListener('afterprint', () => {
        printWindow.close()
        resolve()
      })

      // 如果打印窗口被关闭，也resolve
      const checkClosed = setInterval(() => {
        if (printWindow.closed) {
          clearInterval(checkClosed)
          resolve()
        }
      }, 100)
    } catch (error) {
      reject(error)
    }
  })
}

/**
 * 打印文本内容
 * @param content 文本内容
 * @param title 标题
 * @param options 打印选项
 */
export function printText(
  content: string,
  title?: string,
  options?: PrintOptions
): Promise<void> {
  return new Promise((resolve, reject) => {
    try {
      const printWindow = window.open('', '_blank')
      if (!printWindow) {
        reject(new Error('无法打开打印窗口，请检查浏览器弹窗设置'))
        return
      }

      const printTitle = title || options?.title || '文本打印'

      printWindow.document.write(`
        <!DOCTYPE html>
        <html>
        <head>
          <title>${printTitle}</title>
          <meta charset="UTF-8">
          <style>
            * {
              box-sizing: border-box;
            }
            @media print {
              @page {
                size: ${options?.orientation === 'landscape' ? 'A4 landscape' : 'A4'};
                margin: ${options?.margin?.top || 15}mm ${options?.margin?.right || 15}mm ${options?.margin?.bottom || 15}mm ${options?.margin?.left || 15}mm;
              }
              body {
                margin: 0;
                padding: 0;
                font-family: 'Microsoft YaHei', 'SimSun', Arial, sans-serif;
                font-size: 12pt;
                line-height: 1.8;
                color: #000;
                background: white;
              }
              .print-header {
                display: none;
              }
              .content-wrapper {
                margin: 0;
                padding: 0;
              }
              pre {
                white-space: pre-wrap;
                word-wrap: break-word;
                font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
                font-size: 10pt;
                line-height: 1.6;
                margin: 0;
                padding: 0;
                background: white;
                color: #000;
              }
            }
            @media screen {
              body {
                margin: 0;
                padding: 0;
                font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Microsoft YaHei', Arial, sans-serif;
                background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
                min-height: 100vh;
                display: flex;
                flex-direction: column;
              }
              .print-header {
                width: 100%;
                padding: 24px 30px;
                background: rgba(255, 255, 255, 0.95);
                backdrop-filter: blur(10px);
                box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
                border-bottom: 1px solid rgba(0, 0, 0, 0.05);
              }
              .print-header h2 {
                margin: 0 0 8px 0;
                font-size: 20px;
                font-weight: 600;
                color: #333;
                letter-spacing: 0.5px;
              }
              .print-header p {
                margin: 0;
                font-size: 13px;
                color: #666;
              }
              .content-wrapper {
                flex: 1;
                padding: 30px;
                overflow: auto;
                display: flex;
                justify-content: center;
                align-items: flex-start;
              }
              pre {
                white-space: pre-wrap;
                word-wrap: break-word;
                font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
                font-size: 13px;
                line-height: 1.7;
                background: white;
                padding: 24px;
                border-radius: 8px;
                box-shadow: 0 8px 32px rgba(0, 0, 0, 0.15);
                margin: 0;
                max-width: 100%;
                width: 100%;
                color: #333;
                border: 1px solid rgba(0, 0, 0, 0.05);
              }
            }
          </style>
        </head>
        <body>
          <div class="print-header">
            <h2>${escapeHtml(printTitle)}</h2>
            <p>准备打印中，请稍候...</p>
          </div>
          <div class="content-wrapper">
            <pre>${escapeHtml(content)}</pre>
          </div>
          <script>
            window.onload = function() {
              setTimeout(function() {
                window.print();
                window.onafterprint = function() {
                  window.close();
                };
              }, 300);
            };
          </script>
        </body>
        </html>
      `)
      printWindow.document.close()

      // 监听打印完成
      printWindow.addEventListener('afterprint', () => {
        printWindow.close()
        resolve()
      })

      // 如果打印窗口被关闭，也resolve
      const checkClosed = setInterval(() => {
        if (printWindow.closed) {
          clearInterval(checkClosed)
          resolve()
        }
      }, 100)
    } catch (error) {
      reject(error)
    }
  })
}

/**
 * 转义HTML特殊字符
 */
function escapeHtml(text: string): string {
  const div = document.createElement('div')
  div.textContent = text
  return div.innerHTML
}

/**
 * 检查文件类型是否支持打印
 * @param mimeType MIME类型
 * @returns 是否支持打印
 */
export function isPrintableType(mimeType: string): boolean {
  if (!mimeType) return false
  
  const mime = mimeType.toLowerCase()
  
  // 图片类型
  if (mime.startsWith('image/')) {
    return true
  }
  
  // PDF
  if (mime === 'application/pdf') {
    return true
  }
  
  // 文本类型
  if (mime.startsWith('text/')) {
    return true
  }
  
  // 代码文件（通过扩展名判断，这里只判断MIME）
  if (mime === 'application/json' || mime === 'application/xml') {
    return true
  }
  
  // Office文档类型（Excel、Word、PowerPoint）
  if (isOfficeDocument(mime)) {
    return true
  }
  
  return false
}

/**
 * 检查是否为Office文档
 * @param mimeType MIME类型
 * @returns 是否为Office文档
 */
export function isOfficeDocument(mimeType: string): boolean {
  if (!mimeType) return false
  
  const mime = mimeType.toLowerCase()
  
  // Excel
  if (mime.includes('spreadsheetml') || 
      mime === 'application/vnd.ms-excel' ||
      mime === 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet') {
    return true
  }
  
  // Word
  if (mime.includes('wordprocessingml') ||
      mime === 'application/msword' ||
      mime === 'application/vnd.openxmlformats-officedocument.wordprocessingml.document') {
    return true
  }
  
  // PowerPoint
  if (mime.includes('presentationml') ||
      mime === 'application/vnd.ms-powerpoint' ||
      mime === 'application/vnd.openxmlformats-officedocument.presentationml.presentation') {
    return true
  }
  
  return false
}

/**
 * 打印Office文档（通过iframe尝试打开，如果失败则提示下载）
 * @param fileUrl 文件URL
 * @param fileName 文件名
 * @param options 打印选项
 */
export function printOfficeDocument(
  fileUrl: string,
  fileName: string,
  options?: PrintOptions
): Promise<void> {
  return new Promise((resolve, reject) => {
    try {
      const printWindow = window.open('', '_blank')
      if (!printWindow) {
        reject(new Error('无法打开打印窗口，请检查浏览器弹窗设置。建议先下载文件，然后用相应的Office软件打开并打印。'))
        return
      }

      const title = options?.title || fileName

      printWindow.document.write(`
        <!DOCTYPE html>
        <html>
        <head>
          <title>${title}</title>
          <meta charset="UTF-8">
          <style>
            @media print {
              @page {
                size: ${options?.orientation === 'landscape' ? 'A4 landscape' : 'A4'};
                margin: ${options?.margin?.top || 10}mm ${options?.margin?.right || 10}mm ${options?.margin?.bottom || 10}mm ${options?.margin?.left || 10}mm;
              }
              body {
                margin: 0;
                padding: 0;
              }
              iframe {
                width: 100%;
                height: 100vh;
                border: none;
              }
              .info {
                display: none;
              }
            }
            @media screen {
              body {
                margin: 0;
                padding: 20px;
                background: #f5f5f5;
                font-family: 'Microsoft YaHei', Arial, sans-serif;
              }
              .info {
                background: white;
                padding: 20px;
                border-radius: 8px;
                margin-bottom: 20px;
                box-shadow: 0 2px 8px rgba(0,0,0,0.1);
              }
              .info h3 {
                margin: 0 0 10px 0;
                color: #333;
                font-size: 16px;
              }
              .info p {
                margin: 8px 0;
                color: #666;
                font-size: 14px;
                line-height: 1.6;
              }
              .info a {
                color: #409eff;
                text-decoration: none;
                font-weight: 500;
              }
              .info a:hover {
                text-decoration: underline;
              }
              iframe {
                width: 100%;
                height: calc(100vh - 200px);
                border: 1px solid #ddd;
                border-radius: 4px;
                background: white;
              }
            }
          </style>
        </head>
        <body>
          <div class="info">
            <h3>文档打印提示</h3>
            <p>正在尝试在浏览器中打开文档...</p>
            <p>如果文档无法正常显示，请点击下方链接下载文件，然后用相应的Office软件（如Excel、Word、PowerPoint）打开并打印。</p>
            <p><a href="${fileUrl}" download="${fileName}">📥 下载文件</a></p>
          </div>
          <iframe src="${fileUrl}" onload="setTimeout(function() { try { window.print(); window.onafterprint = function() { window.close(); } } catch(e) { alert('无法直接打印此文档类型。\\n\\n请点击上方"下载文件"链接，下载后用相应的Office软件打开并打印。'); window.close(); } }, 1500);" onerror="alert('无法加载文档。\\n\\n请点击上方"下载文件"链接，下载后用相应的Office软件打开并打印。'); window.close();"></iframe>
        </body>
        </html>
      `)
      printWindow.document.close()

      // 监听打印完成
      printWindow.addEventListener('afterprint', () => {
        printWindow.close()
        resolve()
      })

      // 如果打印窗口被关闭，也resolve
      const checkClosed = setInterval(() => {
        if (printWindow.closed) {
          clearInterval(checkClosed)
          resolve()
        }
      }, 100)
    } catch (error) {
      reject(error)
    }
  })
}

