<template>
  <div>
    <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:16px">
      <h2>VIP 订单管理</h2>
      <el-button type="primary" @click="grantVisible = true">手动开通 VIP</el-button>
    </div>

    <el-tabs v-model="statusFilter" @tab-change="loadOrders">
      <el-tab-pane label="全部" name="" />
      <el-tab-pane label="已完成" name="paid" />
      <el-tab-pane label="待支付" name="pending" />
      <el-tab-pane label="已取消" name="cancelled" />
    </el-tabs>

    <el-table :data="orders" border stripe>
      <el-table-column prop="id" label="订单号" width="80" />
      <el-table-column label="用户" width="130">
        <template #default="{ row }">{{ row.user?.username }}</template>
      </el-table-column>
      <el-table-column label="套餐" width="100">
        <template #default="{ row }">
          <el-tag>{{ vipNames[row.vip_level] }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="duration_days" label="时长" width="80">
        <template #default="{ row }">{{ row.duration_days }}天</template>
      </el-table-column>
      <el-table-column prop="price" label="金额" width="100">
        <template #default="{ row }">
          {{ row.price === 0 ? '赠送' : `¥${row.price.toFixed(2)}` }}
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="({ paid: 'success', pending: 'warning', cancelled: 'info' } as any)[row.status]">
            {{ ({ paid: '已完成', pending: '待支付', cancelled: '已取消' } as any)[row.status] }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="到期时间" width="160">
        <template #default="{ row }">{{ row.expire_at ? formatDate(row.expire_at) : '-' }}</template>
      </el-table-column>
      <el-table-column label="创建时间">
        <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
      </el-table-column>
    </el-table>

    <!-- 手动开通 VIP 对话框 -->
    <el-dialog v-model="grantVisible" title="手动开通 VIP" width="420px">
      <el-form :model="grantForm" label-width="100px">
        <el-form-item label="用户ID">
          <el-input-number v-model="grantForm.user_id" :min="1" style="width:100%" />
        </el-form-item>
        <el-form-item label="VIP等级">
          <el-select v-model="grantForm.vip_level" style="width:100%">
            <el-option label="Free (取消VIP)" :value="0" />
            <el-option label="Basic" :value="1" />
            <el-option label="Pro" :value="2" />
            <el-option label="Ultimate" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item label="有效天数">
          <el-input-number v-model="grantForm.duration_days" :min="1" :max="3650" style="width:100%" />
          <div style="font-size:12px;color:#999;margin-top:4px">
            若用户已有有效 VIP，将在现有基础上延期
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="grantVisible = false">取消</el-button>
        <el-button type="primary" :loading="granting" @click="handleGrant">确认开通</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { adminApi } from '../../api'

const orders = ref([])
const statusFilter = ref('')
const grantVisible = ref(false)
const granting = ref(false)
const grantForm = ref({ user_id: 1, vip_level: 1, duration_days: 30 })

const vipNames: Record<number, string> = { 0: 'Free', 1: 'Basic', 2: 'Pro', 3: 'Ultimate' }

async function loadOrders() {
  const res: any = await adminApi.listOrders(statusFilter.value)
  orders.value = res.data
}

async function handleGrant() {
  granting.value = true
  try {
    await adminApi.grantVIP(grantForm.value)
    ElMessage.success('VIP 开通成功')
    grantVisible.value = false
    loadOrders()
  } finally {
    granting.value = false
  }
}

function formatDate(d: string) {
  return new Date(d).toLocaleString('zh-CN', {
    year: 'numeric', month: '2-digit', day: '2-digit',
    hour: '2-digit', minute: '2-digit',
  })
}

onMounted(loadOrders)
</script>
