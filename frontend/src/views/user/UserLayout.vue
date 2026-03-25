<template>
  <el-container class="user-layout">
    <el-header>
      <div class="nav-brand">⚡ JumpFrp</div>
      <div class="nav-links">
        <router-link to="/dashboard">控制台</router-link>
        <router-link to="/tunnels">我的隧道</router-link>
        <router-link to="/vip">VIP 中心</router-link>
      </div>
      <div class="nav-user">
        <el-tag :type="['info','','warning','danger'][user?.vip_level || 0]" size="small">
          {{ ['Free','Basic','Pro','Ultimate'][user?.vip_level || 0] }}
        </el-tag>
        <span style="margin:0 8px">{{ user?.username }}</span>
        <el-button text size="small" @click="logout">退出</el-button>
      </div>
    </el-header>
    <el-main>
      <router-view />
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../../stores/auth'

const auth = useAuthStore()
const router = useRouter()
const user = computed(() => auth.user)

function logout() {
  auth.logout()
  router.push('/login')
}
</script>

<style scoped>
.user-layout { min-height: 100vh; background: #f5f7fa; }
.el-header {
  display: flex; align-items: center; background: white;
  border-bottom: 1px solid #eee; padding: 0 24px; gap: 24px;
}
.nav-brand { font-size: 20px; font-weight: bold; color: #409eff; white-space: nowrap; }
.nav-links { display: flex; gap: 20px; flex: 1; }
.nav-links a { color: #555; text-decoration: none; font-size: 14px; }
.nav-links a.router-link-active { color: #409eff; font-weight: bold; }
.nav-user { display: flex; align-items: center; white-space: nowrap; }
.el-main { max-width: 1200px; margin: 0 auto; padding: 24px; width: 100%; }
</style>
