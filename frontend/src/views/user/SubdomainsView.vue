<template>
  <div class="subdomains-page">
    <div class="page-header">
      <h2>我的域名</h2>
      <el-button type="primary" @click="openCreate" :disabled="!canCreate">+ 绑定域名</el-button>
    </div>

    <!-- VIP 提示 -->
    <el-alert v-if="user && user.vip_level < 2" :closable="false" style="margin-bottom:16px" type="warning">
      自定义域名需要 Pro 或以上 VIP 等级，当前等级：{{ vipNames[user.vip_level] }}
    </el-alert>

    <!-- 域名列表 -->
    <div v-if="subdomains.length === 0" class="empty">
      <el-empty description="还没有绑定域名" />
    </div>

    <el-table v-else :data="subdomains" border stripe>
      <el-table-column label="域名" min-width="200">
        <template #default="{ row }">
          <el-link :href="`http://${row.subdomain}.jumpfrp.top`" target="_blank">
            {{ row.subdomain }}.jumpfrp.top
          </el-link>
        </template>
      </el-table-column>
      <el-table-column label="绑定隧道" width="120">
        <template #default="{ row }">
          {{ getTunnelName(row.tunnel_id) }}
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="statusType[row.status]" size="small">
            {{ statusName[row.status] }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="120">
        <template #default="{ row }">
          <el-button size="small" type="danger" @click="handleDelete(row)">解绑</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 绑定域名对话框 -->
    <el-dialog v-model="createVisible" title="绑定域名" width="460px">
      <el-form :model="createForm" label-width="80px">
        <el-form-item label="选择隧道">
          <el-select v-model="createForm.tunnel_id" placeholder="选择要绑定域名的隧道" style="width:100%">
            <el-option v-for="t in httpTunnels" :key="t.id" :value="t.id" :label="t.name">
              <span>{{ t.name }}</span>
              <span style="float:right;color:#999;font-size:12px">{{ t.protocol.toUpperCase() }}</span>
            </el-option>
          </el-select>
          <div style="font-size:12px;color:#999;margin-top:4px">仅支持 HTTP/HTTPS 隧道</div>
        </el-form-item>
        <el-form-item label="子域名">
          <el-input v-model="createForm.subdomain" placeholder="如：myapp">
            <template #append>.jumpfrp.top</template>
          </el-input>
          <div style="font-size:12px;color:#999;margin-top:4px">
            绑定后可通过 http://{{ createForm.subdomain || 'xxx' }}.jumpfrp.top 访问
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="handleCreate">绑定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { userApi } from '../../api'
import { useAuthStore } from '../../stores/auth'

const auth = useAuthStore()
const user = computed(() => auth.user)

const subdomains = ref<any[]>([])
const tunnels = ref<any[]>([])
const createVisible = ref(false)
const creating = ref(false)

const vipNames: Record<number, string> = { 0: 'Free', 1: 'Basic', 2: 'Pro', 3: 'Ultimate' }
const statusName: Record<string, string> = { approved: '已批准', pending: '待审批', rejected: '已拒绝' }
const statusType: Record<string, string> = { approved: 'success', pending: 'warning', rejected: 'danger' }

const canCreate = computed(() => (user.value?.vip_level || 0) >= 2)
const httpTunnels = computed(() => tunnels.value.filter(t => t.protocol === 'http' || t.protocol === 'https'))

const createForm = ref({
  tunnel_id: null as number | null,
  subdomain: '',
})

async function loadData() {
  const [subRes, tunnelRes]: any[] = await Promise.all([
    userApi.listSubdomains(),
    userApi.listTunnels(),
  ])
  subdomains.value = subRes.data
  tunnels.value = tunnelRes.data
}

function getTunnelName(tunnelId: number) {
  const t = tunnels.value.find(x => x.id === tunnelId)
  return t?.name || `#${tunnelId}`
}

function openCreate() {
  createForm.value = { tunnel_id: httpTunnels.value[0]?.id || null, subdomain: '' }
  createVisible.value = true
}

async function handleCreate() {
  if (!createForm.value.tunnel_id) {
    ElMessage.warning('请选择隧道')
    return
  }
  if (!createForm.value.subdomain) {
    ElMessage.warning('请输入子域名')
    return
  }
  if (!/^[a-z0-9][a-z0-9-]*[a-z0-9]$/i.test(createForm.value.subdomain)) {
    ElMessage.warning('域名格式不正确，只能包含字母、数字和连字符')
    return
  }
  creating.value = true
  try {
    await userApi.createSubdomain(createForm.value)
    ElMessage.success('域名绑定成功')
    createVisible.value = false
    loadData()
  } finally {
    creating.value = false
  }
}

async function handleDelete(row: any) {
  await ElMessageBox.confirm(`确定解绑域名 "${row.subdomain}.jumpfrp.top"？`, '提示', { type: 'warning' })
  await userApi.deleteSubdomain(row.id)
  ElMessage.success('已解绑')
  loadData()
}

onMounted(loadData)
</script>

<style scoped>
.subdomains-page { padding: 20px }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px }
.empty { padding: 40px 0 }
</style>