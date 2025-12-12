// 密码挑战工具函数
import { getChallenge } from '@/api/auth'
import { rsaEncrypt } from './crypto'

/**
 * 加密密码并获取挑战ID
 * @param password 明文密码
 * @returns 包含加密密码和挑战ID的对象
 */
export const encryptPasswordWithChallenge = async (password: string): Promise<{
  encryptedPassword: string
  challengeId: string
}> => {
  // 获取密码挑战秘钥
  const challengeRes = await getChallenge()
  
  if (!challengeRes.data || !challengeRes.data.publicKey || !challengeRes.data.id) {
    throw new Error('获取密码挑战失败')
  }
  
  // 使用公钥加密密码
  const encryptedPassword = rsaEncrypt(challengeRes.data.publicKey, password)
  
  return {
    encryptedPassword,
    challengeId: challengeRes.data.id
  }
}

/**
 * 加密两个密码（用于修改密码等场景）
 * @param oldPassword 旧密码
 * @param newPassword 新密码
 * @returns 包含加密后的旧密码、新密码和挑战ID的对象
 */
export const encryptTwoPasswordsWithChallenge = async (
  oldPassword: string, 
  newPassword: string
): Promise<{
  encryptedOldPassword: string
  encryptedNewPassword: string
  challengeId: string
}> => {
  // 获取密码挑战秘钥
  const challengeRes = await getChallenge()
  
  if (!challengeRes.data || !challengeRes.data.publicKey || !challengeRes.data.id) {
    throw new Error('获取密码挑战失败')
  }
  
  // 使用同一个公钥加密两个密码
  const encryptedOldPassword = rsaEncrypt(challengeRes.data.publicKey, oldPassword)
  const encryptedNewPassword = rsaEncrypt(challengeRes.data.publicKey, newPassword)
  
  return {
    encryptedOldPassword,
    encryptedNewPassword,
    challengeId: challengeRes.data.id
  }
}
