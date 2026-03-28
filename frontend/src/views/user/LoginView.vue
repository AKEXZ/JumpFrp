<template>
  <div class="login-page">
    <div class="login-box">
      <div class="logo">
        <h1>⚡ JumpFrp</h1>
        <p>高速内网穿透服务</p>
      </div>
      <el-tabs v-model="activeTab">
        <el-tab-pane label="登录" name="login">
          <el-form :model="loginForm" @submit.prevent="handleLogin" label-position="top">
            <el-form-item label="邮箱 / 用户名">
              <el-input v-model="loginForm.email" placeholder="请输入邮箱或用户名" prefix-icon="Message" />
            </el-form-item>
            <el-form-item label="密码">
              <el-input v-model="loginForm.password" type="password" placeholder="请输入密码" prefix-icon="Lock" show-password />
            </el-form-item>
            <el-button type="primary" native-type="submit" :loading="loading" style="width:100%">登录</el-button>
          </el-form>
          <div class="links">
            <span>没有账号？</span>
            <router-link to="/register">立即注册</router-link>
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { userApi } from '../../api'
import { useAuthStore } from '../../stores/auth'

const router = useRouter()
const auth = useAuthStore()
const activeTab = ref('login')
const loading = ref(false)

const loginForm = ref({ email: '', password: '' })

async function handleLogin() {
  if (!loginForm.value.email || !loginForm.value.password) {
    ElMessage.warning('请输入用户名/邮箱和密码')
    return
  }
  loading.value = true
  try {
    const res: any = await userApi.login(loginForm.value)
    auth.setAuth(res.data.token, res.data.user)
    ElMessage.success('登录成功')
    if (auth.isAdmin) {
      router.push('/admin/dashboard')
    } else {
      router.push('/dashboard')
    }
  } catch (e: any) {
    // 错误已在拦截器中显示
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
}
.login-box {
  background: white;
  border-radius: 12px;
  padding: 40px;
  width: 400px;
  box-shadow: 0 20px 60px rgba(0,0,0,0.3);
}
.logo { text-align: center; margin-bottom: 24px; }
.logo h1 { font-size: 28px; color: #409eff; margin: 0; }
.logo p { color: #999; margin: 4px 0 0; }
.links { text-align: center; margin-top: 16px; color: #999; }
.links a { color: #409eff; margin-left: 4px; }
</style>
