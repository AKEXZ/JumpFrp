import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '../api'

interface User {
  id: number
  username: string
  email: string
  vip_level: number
  vip_expire_at: string | null
  status: string
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string>(localStorage.getItem('token') || '')
  const user = ref<User | null>(JSON.parse(localStorage.getItem('user') || 'null'))

  const isAdmin = computed(() => user.value?.username === 'admin')
  const isLoggedIn = computed(() => !!token.value)

  function setAuth(t: string, u: User) {
    token.value = t
    user.value = u
    localStorage.setItem('token', t)
    localStorage.setItem('user', JSON.stringify(u))
    api.defaults.headers.common['Authorization'] = `Bearer ${t}`
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    delete api.defaults.headers.common['Authorization']
  }

  // 初始化时验证 token 有效性
  async function validateToken() {
    if (!token.value) return false
    try {
      const res: any = await api.get('/user/profile')
      if (res.code === 0) {
        user.value = res.data
        localStorage.setItem('user', JSON.stringify(res.data))
        return true
      }
    } catch (e: any) {
      // token 无效，清除登录状态
      console.warn('Token 无效，清除登录状态')
    }
    logout()
    return false
  }

  // 初始化时设置 token
  if (token.value) {
    api.defaults.headers.common['Authorization'] = `Bearer ${token.value}`
  }

  return { token, user, isAdmin, isLoggedIn, setAuth, logout, validateToken }
})
