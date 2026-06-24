// RSA加密工具函数 - 使用 Web Crypto API 实现 RSA-OAEP
import logger from '../../plugins/logger'

/**
 * 将PEM格式公钥转换为CryptoKey
 */
async function importPublicKey(pem: string): Promise<CryptoKey> {
  // 提取PEM内容
  const pemHeader = '-----BEGIN PUBLIC KEY-----'
  const pemFooter = '-----END PUBLIC KEY-----'
  const pemContents = pem
    .replace(pemHeader, '')
    .replace(pemFooter, '')
    .replace(/\s/g, '')

  // Base64解码
  const binaryDer = Uint8Array.from(atob(pemContents), c => c.charCodeAt(0))

  // 导入密钥
  return await crypto.subtle.importKey(
    'spki',
    binaryDer,
    {
      name: 'RSA-OAEP',
      hash: 'SHA-256'
    },
    false,
    ['encrypt']
  )
}

/**
 * 将PEM格式私钥转换为CryptoKey
 */
async function importPrivateKey(pem: string): Promise<CryptoKey> {
  // 提取PEM内容
  const pemHeader = '-----BEGIN PRIVATE KEY-----'
  const pemFooter = '-----END PRIVATE KEY-----'
  const pemContents = pem
    .replace(pemHeader, '')
    .replace(pemFooter, '')
    .replace(/\s/g, '')

  // Base64解码
  const binaryDer = Uint8Array.from(atob(pemContents), c => c.charCodeAt(0))

  // 导入密钥
  return await crypto.subtle.importKey(
    'pkcs8',
    binaryDer,
    {
      name: 'RSA-OAEP',
      hash: 'SHA-256'
    },
    false,
    ['decrypt']
  )
}

/**
 * 使用RSA公钥加密数据（OAEP填充，SHA-256）
 * @param publicKey PEM格式的公钥
 * @param data 要加密的数据
 * @returns Base64编码的加密数据
 */
export const rsaEncrypt = async (publicKey: string, data: string): Promise<string> => {
  try {
    const key = await importPublicKey(publicKey)
    const encoded = new TextEncoder().encode(data)

    const encrypted = await crypto.subtle.encrypt(
      { name: 'RSA-OAEP' },
      key,
      encoded
    )

    // 转换为Base64
    return btoa(String.fromCharCode(...new Uint8Array(encrypted)))
  } catch (error) {
    logger.error('RSA加密错误:', error)
    throw new Error('RSA加密失败')
  }
}

/**
 * 使用RSA私钥解密数据（OAEP填充，SHA-256）
 * @param privateKey PEM格式的私钥
 * @param encryptedData Base64编码的加密数据
 * @returns 解密后的原始数据
 */
export const rsaDecrypt = async (privateKey: string, encryptedData: string): Promise<string> => {
  try {
    const key = await importPrivateKey(privateKey)

    // Base64解码
    const binaryDer = Uint8Array.from(atob(encryptedData), c => c.charCodeAt(0))

    const decrypted = await crypto.subtle.decrypt(
      { name: 'RSA-OAEP' },
      key,
      binaryDer
    )

    return new TextDecoder().decode(decrypted)
  } catch (error) {
    logger.error('RSA解密错误:', error)
    throw new Error('RSA解密失败')
  }
}

/**
 * 同步版本的RSA加密（兼容旧代码，使用JSEncrypt）
 * 注意：此版本使用PKCS1v1.5填充，与后端OAEP不兼容
 * @deprecated 请使用异步版本 rsaEncrypt
 */
export const rsaEncryptSync = (publicKey: string, data: string): string => {
  // 动态导入JSEncrypt（如果需要同步版本）
  throw new Error('请使用异步版本 rsaEncrypt')
}
