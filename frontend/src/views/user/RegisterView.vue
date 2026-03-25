<template>
  <div class="register-page">
    <div class="register-box">
      <div class="logo">
        <h1>⚡ JumpFrp</h1>
        <p>创建你的账号</p>
      </div>
      <el-form :model="form" label-position="top" @submit.prevent="handleRegister">
        <el-form-item label="用户名">
          <el-input v-model="form.username" placeholder="3-20位字母数字" prefix-icon="User" />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="form.email" placeholder="请输入邮箱" prefix-icon="Message" />
        </el-form-item>
        <el-form-item label="验证码">
          <div style="display:flex;gap:8px">
            <el-input v-model="form.code" placeholder="6位验证码" />
            <el-button :disabled="countdown > 0" @click="sendCode" style="white-space:nowrap">
              {{ countdown > 0 ? `${countdown}s` : '发送验证码' }}
            </el-button>
          </div>
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="form.password" type="password" placeholder="至少8位" prefix-icon="Lock" show-password />
        </el-form-item>
        <el-button type="primary" native-type="submit" :loading="loading" style="width:100%">注册</el-button>
      </el-form>
      <div class="links">
        <span>已有账号？</span>
        <router-link to="/login">立即登录</router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { userApi } from '../../api'

const router = useRouter()
const loading = ref(false)
const countdown = ref(0)

const form = ref({ username: '', email: '', password: '', code: '' })

async function sendCode() {
  if (!form.value.email) {
    ElMessage.warning('请先填写邮箱')
    return
  }
  await userApi.sendCode(form.value.email)
  ElMessage.success('验证码已发送')
  countdown.value = 60
  const timer = setInterval(() => {
    countdown.value--
    if (countdown.value <= 0) clearInterval(timer)
  }, 1000)
}

async function handleRegister() {
  loading.value = true
  try {
    await userApi.register(form.value)
    ElMessage.success('注册成功，请登录')
    router.push('/login')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.register-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
}
.register-box {
  background: white;
  border-radius: 12px;
  padding: 40px;
  width: 420px;
  box-shadow: 0 20px 60px rgba(0,0,0,0.3);
}
.logo { text-align: center; margin-bottom: 24px; }
.logo h1 { font-size: 28px; color: #409eff; margin: 0; }
.logo p { color: #999; margin: 4px 0 0; }
.links { text-align: center; margin-top: 16px; color: #999; }
.links a { color: #409eff; margin-left: 4px; }
</style>
