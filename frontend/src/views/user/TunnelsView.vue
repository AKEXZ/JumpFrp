<template>
  <div class="tunnels-page">
    <div class="page-header">
      <h2>我的隧道</h2>
      <el-button type="primary" @click="openCreate">+ 创建隧道</el-button>
    </div>

    <!-- 配额提示 -->
    <el-alert v-if="user" :closable="false" style="margin-bottom:16px"
      :title="`当前套餐：${vipNames[user.vip_level]} | 已用 ${tunnels.length} / ${quotas[user.vip_level].maxTunnels} 条隧道`"
      type="info" />

    <!-- 隧道列表 -->
    <div v-if="tunnels.length === 0" class="empty">
      <el-empty description="还没有隧道，点击右上角创建一个吧" />
    </div>

    <div v-else class="tunnel-grid">
      <el-card v-for="t in tunnels" :key="t.id" class="tunnel-card" shadow="hover">
        <div class="tunnel-header">
          <span class="tunnel-name">{{ t.name }}</span>
          <el-tag :type="t.status === 'active' ? 'success' : 'info'" size="small">
            {{ t.status === 'active' ? '在线' : '离线' }}
          </el-tag>
        </div>
        <div class="tunnel-info">
          <div><span class="label">节点</span>{{ t.node?.name }} ({{ t.node?.region }})</div>
          <div><span class="label">协议</span><el-tag size="small">{{ t.protocol.toUpperCase() }}</el-tag></div>
          <div><span class="label">本地</span>{{ t.local_ip }}:{{ t.local_port }}</div>
          <div v-if="t.subdomain">
            <span class="label">域名</span>
            <el-link :href="`http://${t.subdomain}.jumpfrp.top`" target="_blank">{{ t.subdomain }}.jumpfrp.top</el-link>
          </div>
          <div v-else><span class="label">远程端口</span>{{ t.node?.ip }}:{{ t.remote_port }}</div>
          <div><span class="label">带宽</span>{{ t.bandwidth_limit }} Mbps</div>
        </div>
        <div class="tunnel-actions">
          <el-button size="small" @click="downloadConfig(t)">下载配置</el-button>
          <el-button size="small" @click="showHelp(t)">使用教程</el-button>
          <el-button size="small" type="danger" @click="handleDelete(t)">删除</el-button>
        </div>
      </el-card>
    </div>

    <!-- 创建隧道对话框 -->
    <el-dialog v-model="createVisible" title="创建隧道" width="520px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="隧道名称">
          <el-input v-model="form.name" placeholder="如：我的Web服务" />
        </el-form-item>
        <el-form-item label="选择节点">
          <el-select v-model="form.node_id" placeholder="请选择节点" style="width:100%">
            <el-option v-for="n in nodes" :key="n.id" :value="n.id"
              :label="`${n.name} (${n.region})`"
              :disabled="n.min_vip_level > (user?.vip_level || 0)">
              <span>{{ n.name }}</span>
              <span style="float:right;color:#999;font-size:12px">
                {{ n.region }}
                <el-tag v-if="n.min_vip_level > 0" size="small" type="warning">
                  VIP{{ n.min_vip_level }}+
                </el-tag>
              </span>
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="协议类型">
          <el-radio-group v-model="form.protocol">
            <el-radio-button v-for="p in allowedProtocols" :key="p" :value="p">
              {{ p.toUpperCase() }}
            </el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="本地IP">
          <el-input v-model="form.local_ip" placeholder="127.0.0.1" />
        </el-form-item>
        <el-form-item label="本地端口">
          <el-input-number v-model="form.local_port" :min="1" :max="65535" style="width:100%" />
        </el-form-item>
        <el-form-item v-if="form.protocol === 'http' || form.protocol === 'https'" label="子域名">
          <el-input v-model="form.subdomain" placeholder="留空则使用随机端口">
            <template #append>.jumpfrp.top</template>
          </el-input>
          <div style="font-size:12px;color:#999;margin-top:4px">
            需要 Pro 及以上套餐，留空则使用端口转发
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="handleCreate">创建</el-button>
      </template>
    </el-dialog>

    <!-- 使用教程对话框 -->
    <el-dialog v-model="helpVisible" title="使用教程" width="600px">
      <el-steps :active="2" direction="vertical" style="margin-bottom:16px">
        <el-step title="下载 frpc 客户端">
          <template #description>
            前往 <el-link href="https://github.com/fatedier/frp/releases" target="_blank">frp 官方 Releases</el-link>
            下载对应系统的 frpc
          </template>
        </el-step>
        <el-step title="下载配置文件">
          <template #description>
            点击"下载配置"按钮，将 frpc.ini 保存到 frpc 同目录
          </template>
        </el-step>
        <el-step title="启动 frpc">
          <template #description>
            <el-code>frpc -c frpc.ini</el-code>
          </template>
        </el-step>
      </el-steps>
      <el-alert v-if="currentTunnel" type="success" :closable="false">
        <template #title>
          连接地址：
          <strong v-if="currentTunnel.subdomain">
            {{ currentTunnel.protocol }}://{{ currentTunnel.subdomain }}.jumpfrp.top
          </strong>
          <strong v-else>
            {{ currentTunnel.node?.ip }}:{{ currentTunnel.remote_port }}
          </strong>
        </template>
      </el-alert>
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

const tunnels = ref<any[]>([])
const nodes = ref<any[]>([])
const createVisible = ref(false)
const helpVisible = ref(false)
const creating = ref(false)
const currentTunnel = ref<any>(null)

const vipNames: Record<number, string> = { 0: 'Free', 1: 'Basic', 2: 'Pro', 3: 'Ultimate' }
const quotas: Record<number, any> = {
  0: { maxTunnels: 1, protocols: ['tcp'] },
  1: { maxTunnels: 5, protocols: ['tcp', 'udp'] },
  2: { maxTunnels: 20, protocols: ['tcp', 'udp', 'http', 'https'] },
  3: { maxTunnels: 9999, protocols: ['tcp', 'udp', 'http', 'https'] },
}

const allowedProtocols = computed(() => {
  const level = user.value?.vip_level || 0
  return quotas[level]?.protocols || ['tcp']
})

const form = ref({
  node_id: null as number | null,
  name: '',
  protocol: 'tcp',
  local_ip: '127.0.0.1',
  local_port: 8080,
  subdomain: '',
})

async function loadData() {
  const [tunnelRes, nodeRes]: any[] = await Promise.all([
    userApi.listTunnels(),
    userApi.listAvailableNodes(),
  ])
  tunnels.value = tunnelRes.data
  nodes.value = nodeRes.data
}

function openCreate() {
  form.value = { node_id: null, name: '', protocol: allowedProtocols.value[0], local_ip: '127.0.0.1', local_port: 8080, subdomain: '' }
  createVisible.value = true
}

async function handleCreate() {
  if (!form.value.node_id) {
    ElMessage.warning('请选择节点')
    return
  }
  creating.value = true
  try {
    await userApi.createTunnel(form.value)
    ElMessage.success('隧道创建成功')
    createVisible.value = false
    loadData()
  } finally {
    creating.value = false
  }
}

async function handleDelete(t: any) {
  await ElMessageBox.confirm(`确定删除隧道「${t.name}」？`, '警告', { type: 'warning' })
  await userApi.deleteTunnel(t.id)
  ElMessage.success('删除成功')
  loadData()
}

async function downloadConfig(t: any) {
  const token = localStorage.getItem('token')
  const res = await fetch(`/api/user/tunnels/${t.id}/frpc-config`, {
    headers: { 'Authorization': `Bearer ${token}` }
  })
  const text = await res.text()
  const blob = new Blob([text], { type: 'text/plain;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `frpc-${t.name}.toml`
  a.click()
  URL.revokeObjectURL(url)
}

function showHelp(t: any) {
  currentTunnel.value = t
  helpVisible.value = true
}

onMounted(loadData)
</script>

<style scoped>
.tunnels-page { padding: 0 }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.page-header h2 { margin: 0; }
.tunnel-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 16px; }
.tunnel-card { border-radius: 8px; }
.tunnel-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.tunnel-name { font-size: 16px; font-weight: bold; }
.tunnel-info { font-size: 14px; line-height: 2; }
.tunnel-info .label { color: #999; margin-right: 8px; min-width: 60px; display: inline-block; }
.tunnel-actions { margin-top: 12px; display: flex; gap: 8px; }
</style>
