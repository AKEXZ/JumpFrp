<template>
  <div>
    <h2 style="margin-bottom:16px">隧道管理</h2>
    <el-table :data="tunnels" border stripe>
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column label="用户" width="120">
        <template #default="{ row }">{{ row.user?.username }}</template>
      </el-table-column>
      <el-table-column prop="name" label="隧道名称" />
      <el-table-column label="节点" width="120">
        <template #default="{ row }">{{ row.node?.name }}</template>
      </el-table-column>
      <el-table-column prop="protocol" label="协议" width="80">
        <template #default="{ row }">
          <el-tag size="small">{{ row.protocol?.toUpperCase() }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="本地" width="160">
        <template #default="{ row }">{{ row.local_ip }}:{{ row.local_port }}</template>
      </el-table-column>
      <el-table-column label="远程端口" width="100">
        <template #default="{ row }">{{ row.remote_port || row.subdomain }}</template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="row.status === 'active' ? 'success' : 'info'" size="small">
            {{ row.status === 'active' ? '在线' : '离线' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="100">
        <template #default="{ row }">
          <el-button size="small" type="danger" @click="handleDelete(row.id)">强制删除</el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminApi } from '../../api'

const tunnels = ref([])

async function loadTunnels() {
  const res: any = await adminApi.listTunnels()
  tunnels.value = res.data
}

async function handleDelete(id: number) {
  await ElMessageBox.confirm('确定强制删除该隧道？', '警告', { type: 'warning' })
  await adminApi.deleteTunnel(id)
  ElMessage.success('已删除')
  loadTunnels()
}

onMounted(loadTunnels)
</script>
