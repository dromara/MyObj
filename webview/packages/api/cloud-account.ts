import { get, post, del } from '@myobj/http'

export const getCloudAccounts = () => get('/cloud/accounts')
export const getCloudAccountStatus = () => get('/cloud/accounts/status')
export const deleteCloudAccount = (provider: string) => del(`/cloud/accounts/${provider}`)

// OAuth 登录
export const getAliyunAuthUrl = () => get('/cloud/aliyun/auth')
export const getAliyunStatus = () => get('/cloud/aliyun/status')
export const aliyunLogout = () => post('/cloud/aliyun/logout')

export const getBaiduAuthUrl = () => get('/cloud/baidu/auth')
export const getBaiduStatus = () => get('/cloud/baidu/status')
export const baiduLogout = () => post('/cloud/baidu/logout')

export const getXunleiAuthUrl = () => get('/cloud/xunlei/auth')
export const xunleiLogin = (username: string, password: string) => post('/cloud/xunlei/login', { username, password })
export const getXunleiStatus = () => get('/cloud/xunlei/status')
export const xunleiLogout = () => post('/cloud/xunlei/logout')

export const getPikPakAuthUrl = () => get('/cloud/pikpak/auth')
export const getPikPakStatus = () => get('/cloud/pikpak/status')
export const pikpakLogout = () => post('/cloud/pikpak/logout')

// Cookie 登录
export const saveQuarkCookie = (cookie: string) => post('/cloud/quark/cookie/save', { cookie })
export const getQuarkStatus = () => get('/cloud/quark/cookie/status')
export const quarkLogout = () => post('/cloud/quark/cookie/delete')

export const saveUcCookie = (cookie: string) => post('/cloud/uc/cookie/save', { cookie })
export const getUcStatus = () => get('/cloud/uc/cookie/status')
export const ucLogout = () => post('/cloud/uc/cookie/delete')

export const save115Cookie = (cookie: string) => post('/cloud/115/cookie', { cookie })
export const get115Status = () => get('/cloud/115/status')
export const logout115 = () => post('/cloud/115/logout')

// 用户名密码登录
export const tianyiLogin = (username: string, password: string) => post('/cloud/tianyi/login', { username, password })
export const getTianyiStatus = () => get('/cloud/tianyi/status')
export const tianyiLogout = () => post('/cloud/tianyi/logout')
