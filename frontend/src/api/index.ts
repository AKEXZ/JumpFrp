import axios from 'axios'
import { ElMessage } from 'element-plus'

const isDev = import.meta.env.DEV
export const api = axios.create({
  baseURL: isDev ? (import.meta.env.VITE_API_URL || 'http://localhost:8080/api') : '/api',
  timeout: 10000,
})

// 响应拦截器
api.interceptors.response.use(
  (res) => {
    if (res.data.code !== 0) {
      ElMessage.error(res.data.msg || '请求失败')
      return Promise.reject(res.data)
    }
    return res.data
  },
  (err) => {
    if (err.response?.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      window.location.href = '/login'
    }
    ElMessage.error(err.response?.data?.msg || '网络错误')
    return Promise.reject(err)
  }
)

// 用户 API
export const userApi = {
  sendCode: (email: string) => api.post('/user/auth/send-code', { email }),
  register: (data: any) => api.post('/user/auth/register', data),
  login: (data: any) => api.post('/user/auth/login', data),
  profile: () => api.get('/user/profile'),
  updatePassword: (data: any) => api.put('/user/password', data),
  listNodes: () => api.get('/user/nodes'),
  listTunnels: () => api.get('/user/tunnels'),
  createTunnel: (data: any) => api.post('/user/tunnels', data),
  deleteTunnel: (id: number) => api.delete(`/user/tunnels/${id}`),
  getFrpcConfig: (id: number) =>
    api.get(`/user/tunnels/${id}/frpc-config`, {
      responseType: 'text',
      transformResponse: [(data) => data],
    }).then((res: any) => res.data),
  // 域名
  listSubdomains: () => api.get('/user/subdomains'),
  createSubdomain: (data: any) => api.post('/user/subdomains', data),
  deleteSubdomain: (id: number) => api.delete(`/user/subdomains/${id}`),
  getVIPPlans: () => api.get('/user/vip/plans'),
  getVIPInfo: () => api.get('/user/vip/info'),
  getVIPOrders: () => api.get('/user/vip/orders'),
}

// 管理员 API
export const adminApi = {
  dashboard: () => api.get('/admin/dashboard'),
  listUsers: (params?: any) => api.get('/admin/users', { params }),
  createUser: (data: any) => api.post('/admin/users', data),
  setVIP: (id: number, data: any) => api.put(`/admin/users/${id}/vip`, data),
  banUser: (id: number, ban: boolean) => api.put(`/admin/users/${id}/ban`, { ban }),
  resetUserPassword: (id: number, password: string) => api.put(`/admin/users/${id}/password`, { password }),
  deleteUser: (id: number) => api.delete(`/admin/users/${id}`),
  listNodes: () => api.get('/admin/nodes'),
  createNode: (data: any) => api.post('/admin/nodes', data),
  updateNode: (id: number, data: any) => api.put(`/admin/nodes/${id}`, data),
  deleteNode: (id: number) => api.delete(`/admin/nodes/${id}`),
  getInstallCmd: (id: number) => api.get(`/admin/nodes/${id}/install-cmd`),
  listTunnels: () => api.get('/admin/tunnels'),
  deleteTunnel: (id: number) => api.delete(`/admin/tunnels/${id}`),
  listOrders: (status?: string) => api.get('/admin/orders', { params: { status } }),
  grantVIP: (data: any) => api.post('/admin/vip/grant', data),
  getSettings: () => api.get('/admin/settings'),
  saveSmtp: (data: any) => api.post('/admin/settings/smtp', data),
  saveSite: (data: any) => api.post('/admin/settings/site', data),
  testSmtp: (email: string) => api.post('/admin/settings/smtp/test', { email }),
  // 域名管理
  listSubdomains: (status?: string) => api.get('/admin/subdomains', { params: { status } }),
  createSubdomain: (data: any) => api.post('/admin/subdomains', data),
  approveSubdomain: (id: number, approve: boolean) => api.put(`/admin/subdomains/${id}/approve`, { approve }),
  deleteSubdomain: (id: number) => api.delete(`/admin/subdomains/${id}`),
}
