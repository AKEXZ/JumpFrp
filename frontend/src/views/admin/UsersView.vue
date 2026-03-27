<template>
  <div>
    <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:16px">
      <h2>用户管理</h2>
      <div style="display:flex;gap:8px">
        <el-input v-model="keyword" placeholder="搜索用户名/邮箱" style="width:220px" clearable @change="loadUsers" />
        <el-button type="primary" @click="openCreate">+ 手动添加用户</el-button>
      </div>
    </div>

    <el-table :data="users" border stripe>
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="username" label="用户名" width="130" />
      <el-table-column prop="email" label="邮箱" min-width="160" />
      <el-table-column prop="vip_level" label="VIP等级" width="110">
        <template #default="{ row }">
          <el-tag :type="(['info','','warning','danger'] as any)[row.vip_level]">
            {{ ['Free','Basic','Pro','Ultimate'][row.vip_level] }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="vip_expire_at" label="VIP到期" width="120">
        <template #default="{ row }">
          <span style="font-size:12px">{{ row.vip_expire_at ? row.vip_expire_at.slice(0,10) : '-' }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="email_verified" label="邮箱验证" width="90">
        <template #default="{ row }">
          <el-tag :type="row.email_verified ? 'success' : 'info'" size="small">
            {{ row.email_verified ? '已验证' : '未验证' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.status === 'active' ? 'success' : 'danger'" size="small">
            {{ row.status === 'active' ? '正常' : '封禁' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="注册时间" width="120">
        <template #default="{ row }">
          <span style="font-size:12px">{{ row.created_at?.slice(0,10) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="280" fixed="right">
        <template #default="{ row }">
          <el-button size="small" type="primary" plain @click="openVIP(row)">设置VIP</el-button>
          <el-button size="small" @click="openResetPwd(row)">重置密码</el-button>
          <el-button size="small"
            :type="row.status === 'active' ? 'warning' : 'success'"
            @click="toggleBan(row)">
            {{ row.status === 'active' ? '封禁' : '解封' }}
          </el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      v-model:current-page="page" :page-size="20" :total="total"
      layout="total, prev, pager, next" style="margin-top:16px"
      @current-change="loadUsers"
    />

    <!-- 手动添加用户对话框 -->
    <el-dialog v-model="createVisible" title="手动添加用户" width="460px">
      <el-form :model="createForm" label-width="100px">
        <el-form-item label="用户名">
          <el-input v-model="createForm.username" placeholder="3-20位字母数字" />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="createForm.email" placeholder="用户邮箱" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="createForm.password" type="password" show-password placeholder="至少8位" />
        </el-form-item>
        <el-form-item label="VIP等级">
          <el-select v-model="createForm.vip_level" style="width:100%">
            <el-option label="Free" :value="0" />
            <el-option label="Basic" :value="1" />
            <el-option label="Pro" :value="2" />
            <el-option label="Ultimate" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item label="VIP天数" v-if="createForm.vip_level > 0">
          <el-input-number v-model="createForm.vip_days" :min="1" :max="3650" style="width:100%" />
        </el-form-item>
        <el-form-item label="跳过验证">
          <el-switch v-model="createForm.email_verified" />
          <span style="margin-left:8px;color:#999;font-size:12px">开启则无需邮箱验证直接激活</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="handleCreate">创建用户</el-button>
      </template>
    </el-dialog>

    <!-- 设置VIP对话框 -->
    <el-dialog v-model="vipVisible" title="设置VIP" width="400px">
      <el-form :model="vipForm" label-width="100px">
        <el-form-item label="用户">
          <strong>{{ currentUser?.username }}</strong>（{{ currentUser?.email }}）
        </el-form-item>
        <el-form-item label="VIP等级">
          <el-select v-model="vipForm.vip_level" style="width:100%">
            <el-option label="Free（取消VIP）" :value="0" />
            <el-option label="Basic" :value="1" />
            <el-option label="Pro" :value="2" />
            <el-option label="Ultimate" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item label="有效天数">
          <el-input-number v-model="vipForm.days" :min="1" :max="3650" style="width:100%" />
          <div style="font-size:12px;color:#999;margin-top:4px">
            若用户已有有效VIP，将在现有基础上延期
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="vipVisible = false">取消</el-button>
        <el-button type="primary" :loading="settingVip" @click="handleSetVIP">确认</el-button>
      </template>
    </el-dialog>

    <!-- 重置密码对话框 -->
    <el-dialog v-model="resetPwdVisible" title="重置密码" width="400px">
      <el-form label-width="100px">
        <el-form-item label="用户">
          <strong>{{ currentUser?.username }}</strong>
        </el-form-item>
        <el-form-item label="新密码">
          <el-input v-model="newPassword" type="password" show-password placeholder="至少8位" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="resetPwdVisible = false">取消</el-button>
        <el-button type="primary" :loading="resettingPwd" @click="handleResetPwd">确认重置</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminApi } from '../../api'

const users = ref([])
const total = ref(0)
const page = ref(1)
const keyword = ref('')

// 创建用户
const createVisible = ref(false)
const creating = ref(false)
const createForm = ref({
  username: '', email: '', password: '',
  vip_level: 0, vip_days: 30, email_verified: true,
})

// 设置VIP
const vipVisible = ref(false)
const settingVip = ref(false)
const currentUser = ref<any>(null)
const vipForm = ref({ vip_level: 0, days: 30 })

// 重置密码
const resetPwdVisible = ref(false)
const resettingPwd = ref(false)
const newPassword = ref('')

async function loadUsers() {
  const res: any = await adminApi.listUsers({ page: page.value, size: 20, keyword: keyword.value })
  users.value = res.data.list
  total.value = res.data.total
}

function openCreate() {
  createForm.value = { username: '', email: '', password: '', vip_level: 0, vip_days: 30, email_verified: true }
  createVisible.value = true
}

async function handleCreate() {
  if (!createForm.value.username || !createForm.value.email || !createForm.value.password) {
    ElMessage.warning('请填写完整信息')
    return
  }
  creating.value = true
  try {
    await adminApi.createUser(createForm.value)
    ElMessage.success('用户创建成功')
    createVisible.value = false
    loadUsers()
  } finally {
    creating.value = false
  }
}

function openVIP(user: any) {
  currentUser.value = user
  vipForm.value = { vip_level: user.vip_level, days: 30 }
  vipVisible.value = true
}

async function handleSetVIP() {
  settingVip.value = true
  try {
    await adminApi.setVIP(currentUser.value.id, vipForm.value)
    ElMessage.success('VIP 设置成功')
    vipVisible.value = false
    loadUsers()
  } finally {
    settingVip.value = false
  }
}

function openResetPwd(user: any) {
  currentUser.value = user
  newPassword.value = ''
  resetPwdVisible.value = true
}

async function handleResetPwd() {
  if (newPassword.value.length < 8) {
    ElMessage.warning('密码至少8位')
    return
  }
  resettingPwd.value = true
  try {
    await adminApi.resetUserPassword(currentUser.value.id, newPassword.value)
    ElMessage.success('密码已重置')
    resetPwdVisible.value = false
  } finally {
    resettingPwd.value = false
  }
}

async function toggleBan(user: any) {
  const ban = user.status === 'active'
  await adminApi.banUser(user.id, ban)
  ElMessage.success(ban ? '已封禁' : '已解封')
  loadUsers()
}

async function handleDelete(user: any) {
  await ElMessageBox.confirm(`确定删除用户 "${user.username}"？此操作不可恢复。`, '警告', { type: 'warning' })
  await adminApi.deleteUser(user.id)
  ElMessage.success('用户已删除')
  loadUsers()
}

onMounted(loadUsers)
</script>
