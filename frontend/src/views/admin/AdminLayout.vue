<template>
  <el-container class="admin-layout">
    <el-aside width="220px">
      <div class="sidebar-logo">⚡ JumpFrp</div>
      <el-menu :default-active="$route.path" router background-color="#1a1a2e" text-color="#ccc" active-text-color="#409eff">
        <el-menu-item index="/admin/dashboard">
          <el-icon><DataAnalysis /></el-icon>
          <span>仪表盘</span>
        </el-menu-item>
        <el-menu-item index="/admin/users">
          <el-icon><User /></el-icon>
          <span>用户管理</span>
        </el-menu-item>
        <el-menu-item index="/admin/nodes">
          <el-icon><Monitor /></el-icon>
          <span>节点管理</span>
        </el-menu-item>
        <el-menu-item index="/admin/tunnels">
          <el-icon><Connection /></el-icon>
          <span>隧道管理</span>
        </el-menu-item>
        <el-menu-item index="/admin/subdomains">
          <el-icon><Link /></el-icon>
          <span>域名管理</span>
        </el-menu-item>
        <el-menu-item index="/admin/orders">
          <el-icon><CreditCard /></el-icon>
          <span>VIP订单</span>
        </el-menu-item>
        <el-menu-item index="/admin/settings">
          <el-icon><Setting /></el-icon>
          <span>系统设置</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header>
        <div class="header-right">
          <span>管理员</span>
          <el-button text @click="logout">退出</el-button>
        </div>
      </el-header>
      <el-main>
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from '../../stores/auth'

const router = useRouter()
const auth = useAuthStore()

function logout() {
  auth.logout()
  router.push('/login')
}
</script>

<style scoped>
.admin-layout { height: 100vh; }
.sidebar-logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  font-weight: bold;
  color: #409eff;
  background: #1a1a2e;
  border-bottom: 1px solid #2a2a4e;
}
.el-aside { background: #1a1a2e; }
.el-header {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  border-bottom: 1px solid #eee;
  background: white;
}
.header-right { display: flex; align-items: center; gap: 12px; }
</style>
