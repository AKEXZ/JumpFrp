<template>
  <div>
    <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:16px">
      <h2>域名管理</h2>
      <div style="display:flex;gap:8px">
        <el-select v-model="statusFilter" placeholder="状态筛选" style="width:120px" @change="loadData">
          <el-option label="全部" value="all" />
          <el-option label="已批准" value="approved" />
          <el-option label="待审批" value="pending" />
          <el-option label="已拒绝" value="rejected" />
        </el-select>
        <el-button type="primary" @click="openCreate">+ 添加域名</el-button>
      </div>
    </div>

    <el-table :data="subdomains" border stripe>
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column label="域名" min-width="180">
        <template #default="{ row }">
          <el-link :href="`http://${row.subdomain}.jumpfrp.top`" target="_blank">
            {{ row.subdomain }}.jumpfrp.top
          </el-link>
        </template>
      </el-table-column>
      <el-table-column label="用户" width="140">
        <template #default="{ row }">
          {{ row.user?.username || '-' }}
        </template>
      </el-table-column>
      <el-table-column label="隧道" width="100">
        <template #default="{ row }">
          {{ row.tunnel_id || '-' }}
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="statusType[row.status]" size="small">
            {{ statusName[row.status] }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="申请时间" width="160">
        <template #default="{ row }">
          {{ row.created_at?.replace('T', ' ').slice(0, 16) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <template v-if="row.status === 'pending'">
            <el-button size="small" type="success" @click="handleApprove(row, true)">批准</el-button>
            <el-button size="small" type="warning" @click="handleApprove(row, false)">拒绝</el-button>
          </template>
          <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 添加域名对话框 -->
    <el-dialog v-model="createVisible" title="添加域名" width="460px">
      <el-form :model="createForm" label-width="80px">
        <el-form-item label="用户ID">
          <el-input-number v-model="createForm.user_id" :min="1" style="width:100%" />
        </el-form-item>
        <el-form-item label="隧道ID">
          <el-input-number v-model="createForm.tunnel_id" :min="0" style="width:100%" />
          <div style="font-size:12px;color:#999">可选，绑定到指定隧道</div>
        </el-form-item>
        <el-form-item label="子域名">
          <el-input v-model="createForm.subdomain" placeholder="如：myapp">
            <template #append>.jumpfrp.top</template>
          </el-input>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="handleCreate">添加</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminApi } from '../../api'

const subdomains = ref<any[]>([])
const statusFilter = ref('all')
const createVisible = ref(false)
const creating = ref(false)

const statusName: Record<string, string> = {
  approved: '已批准',
  pending: '待审批',
  rejected: '已拒绝',
}
const statusType: Record<string, string> = {
  approved: 'success',
  pending: 'warning',
  rejected: 'danger',
}

const createForm = ref({
  user_id: 1,
  tunnel_id: 0,
  subdomain: '',
})

async function loadData() {
  const res: any = await adminApi.listSubdomains(statusFilter.value)
  subdomains.value = res.data
}

function openCreate() {
  createForm.value = { user_id: 1, tunnel_id: 0, subdomain: '' }
  createVisible.value = true
}

async function handleCreate() {
  if (!createForm.value.subdomain) {
    ElMessage.warning('请输入子域名')
    return
  }
  creating.value = true
  try {
    await adminApi.createSubdomain(createForm.value)
    ElMessage.success('添加成功')
    createVisible.value = false
    loadData()
  } finally {
    creating.value = false
  }
}

async function handleApprove(row: any, approve: boolean) {
  await adminApi.approveSubdomain(row.id, approve)
  ElMessage.success(approve ? '已批准' : '已拒绝')
  loadData()
}

async function handleDelete(row: any) {
  await ElMessageBox.confirm(`确定删除域名 "${row.subdomain}.jumpfrp.top"？`, '警告', { type: 'warning' })
  await adminApi.deleteSubdomain(row.id)
  ElMessage.success('删除成功')
  loadData()
}

onMounted(loadData)
</script>