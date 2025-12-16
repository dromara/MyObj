// RSA加密工具函数
import JSEncrypt from 'jsencrypt'
import md5 from 'js-md5'

/**
 * 使用RSA公钥加密数据（PKCS1填充）
 * @param publicKey PEM格式的公钥
 * @param data 要加密的数据
 * @returns Base64编码的加密数据
 */
export const rsaEncrypt = (publicKey: string, data: string): string => {
  try {
    const encrypt = new JSEncrypt()
    encrypt.setPublicKey(publicKey)
    const encrypted = encrypt.encrypt(data)
    
    if (!encrypted) {
      throw new Error('RSA加密失败')
    }
    
    return encrypted
  } catch (error) {
    console.error('RSA加密错误:', error)
    throw new Error('RSA加密失败')
  }
}

/**
 * 使用RSA私钥解密数据
 * @param privateKey PEM格式的私钥
 * @param encryptedData Base64编码的加密数据
 * @returns 解密后的原始数据
 */
export const rsaDecrypt = (privateKey: string, encryptedData: string): string => {
  try {
    const decrypt = new JSEncrypt()
    decrypt.setPrivateKey(privateKey)
    const decrypted = decrypt.decrypt(encryptedData)
    
    if (!decrypted) {
      throw new Error('RSA解密失败')
    }
    
    return decrypted
  } catch (error) {
    console.error('RSA解密错误:', error)
    throw new Error('RSA解密失败')
  }
}

/**
 * 计算文件的MD5哈希值
 * @param file 要计算MD5的文件对象
 * @returns Promise<string> MD5哈希值（十六进制字符串）
 */
export const calculateFileMD5 = (file: File): Promise<string> => {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = (e) => {
      const arrayBuffer = e.target?.result as ArrayBuffer
      const hash = md5(new Uint8Array(arrayBuffer))
      resolve(hash)
    }
    reader.onerror = reject
    reader.readAsArrayBuffer(file)
  })
}
