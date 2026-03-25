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

  // 初始化时设置 token
  if (token.value) {
    api.defaults.headers.common['Authorization'] = `Bearer ${token.value}`
  }

  return { token, user, isAdmin, isLoggedIn, setAuth, logout }
})
