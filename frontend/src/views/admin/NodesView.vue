<template>
  <div>
    <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:16px">
      <h2>节点管理</h2>
      <el-button type="primary" @click="openCreate">+ 添加节点</el-button>
    </div>

    <el-table :data="nodes" border stripe>
      <el-table-column prop="name" label="节点名称" min-width="120" />
      <el-table-column prop="slug" label="标识" width="110" />
      <el-table-column prop="ip" label="IP地址" width="140" />
      <el-table-column prop="region" label="地区" width="110" />
      <el-table-column prop="frps_port" label="frps端口" width="90" />
      <el-table-column label="端口池" width="140">
        <template #default="{ row }">{{ row.port_range_start }} - {{ row.port_range_end }}</template>
      </el-table-column>
      <el-table-column label="负载" width="160">
        <template #default="{ row }">
          <div v-if="row.status === 'online'" style="font-size:12px;line-height:1.8">
            <div>CPU: {{ row.cpu_usage?.toFixed(1) }}%</div>
            <div>内存: {{ row.memory_usage?.toFixed(1) }}%</div>
            <div>连接: {{ row.current_conns }}</div>
          </div>
          <span v-else style="color:#999">-</span>
        </template>
      </el-table-column>
      <el-table-column label="最后心跳" width="110">
        <template #default="{ row }">
          <span style="font-size:12px">{{ row.last_heartbeat ? formatTime(row.last_heartbeat) : '从未' }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="80">
        <template #default="{ row }">
          <el-tag size="small" :type="({ online: 'success', offline: 'danger', maintain: 'warning' } as Record<string,any>)[row.status]">
            {{ ({ online: '在线', offline: '离线', maintain: '维护' } as Record<string,string>)[row.status] }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="280" fixed="right">
        <template #default="{ row }">
          <el-button size="small" type="primary" plain @click="openEdit(row)">编辑</el-button>
          <el-button size="small" @click="showInstallCmd(row)">安装</el-button>
          <el-button size="small" type="warning" @click="showUninstallCmd()">卸载</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row.id)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 创建 / 编辑节点对话框 -->
    <el-dialog v-model="formVisible" :title="isEdit ? '编辑节点' : '添加节点'" width="620px">
      <el-form :model="nodeForm" label-width="120px">
        <el-form-item label="节点名称" required>
          <el-input v-model="nodeForm.name" placeholder="如：上海节点01" />
        </el-form-item>
        <el-form-item label="节点标识" required>
          <el-input v-model="nodeForm.slug" placeholder="必填，唯一标识，如：sh-01、hz-01（创建后不可修改）" :disabled="isEdit" />
        </el-form-item>
        <el-form-item label="IP地址" required>
          <el-input v-model="nodeForm.ip" placeholder="公网 IP" />
        </el-form-item>
        <el-form-item label="地区" required>
          <el-input v-model="nodeForm.region" placeholder="如：中国·上海" />
        </el-form-item>
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="frps端口">
              <el-input-number v-model="nodeForm.frps_port" :min="1" :max="65535" style="width:100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="Agent端口">
              <el-input-number v-model="nodeForm.agent_port" :min="1" :max="65535" style="width:100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="端口池起始">
              <el-input-number v-model="nodeForm.port_range_start" :min="1024" :max="65535" style="width:100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="端口池结束">
              <el-input-number v-model="nodeForm.port_range_end" :min="1024" :max="65535" style="width:100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="排除端口">
          <el-input v-model="nodeForm.port_excludes" placeholder="逗号分隔，如：10080,10443" />
        </el-form-item>
        <el-form-item label="最低VIP等级">
          <el-select v-model="nodeForm.min_vip_level" style="width:100%">
            <el-option label="免费用户可用" :value="0" />
            <el-option label="Basic 及以上" :value="1" />
            <el-option label="Pro 及以上" :value="2" />
            <el-option label="Ultimate 专属" :value="3" />
          </el-select>
        </el-form-item>
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="最大连接数">
              <el-input-number v-model="nodeForm.max_connections" :min="1" style="width:100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="带宽上限(Mbps)">
              <el-input-number v-model="nodeForm.bandwidth_limit" :min="0" style="width:100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="节点状态">
          <el-select v-model="nodeForm.status" style="width:100%">
            <el-option label="在线" value="online" />
            <el-option label="离线" value="offline" />
            <el-option label="维护中" value="maintain" />
          </el-select>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="nodeForm.remark" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">
          {{ isEdit ? '保存修改' : '创建节点' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 安装命令对话框 -->
    <el-dialog v-model="cmdVisible" title="一键安装命令" width="700px">
      <el-alert type="info" :closable="false" style="margin-bottom:16px">
        在目标服务器（Ubuntu）上以 root 身份执行以下命令，将自动安装 frps 和节点 Agent 并注册为系统服务。
      </el-alert>
      <el-input v-model="installCmd" type="textarea" :rows="4" readonly />
      <template #footer>
        <el-button type="primary" @click="copyCmd">复制命令</el-button>
        <el-button @click="cmdVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 卸载命令对话框 -->
    <el-dialog v-model="uninstallVisible" title="节点卸载命令" width="700px">
      <el-alert type="warning" :closable="false" style="margin-bottom:16px">
        在目标服务器上以 root 身份执行以下命令，将停止并卸载 frps 和 Agent 服务，删除所有相关文件。
      </el-alert>
      <el-input v-model="uninstallCmd" type="textarea" :rows="3" readonly />
      <template #footer>
        <el-button type="primary" @click="copyUninstallCmd">复制命令</el-button>
        <el-button @click="uninstallVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminApi } from '../../api'

const nodes = ref([])
const formVisible = ref(false)
const cmdVisible = ref(false)
const uninstallVisible = ref(false)
const isEdit = ref(false)
const saving = ref(false)
const installCmd = ref('')
const uninstallCmd = ref('')
const editId = ref<number | null>(null)
let refreshTimer: ReturnType<typeof setInterval>

const defaultForm = () => ({
  name: '', slug: '', ip: '', region: '',
  frps_port: 7000, agent_port: 7500,
  port_range_start: 10000, port_range_end: 20000,
  port_excludes: '', min_vip_level: 0,
  max_connections: 100, bandwidth_limit: 0,
  status: 'offline', remark: '',
})
const nodeForm = ref(defaultForm())

async function loadNodes() {
  const res: any = await adminApi.listNodes()
  nodes.value = res.data
}

function formatTime(t: string) {
  const d = new Date(t)
  const diff = Math.floor((Date.now() - d.getTime()) / 1000)
  if (diff < 60) return `${diff}秒前`
  if (diff < 3600) return `${Math.floor(diff / 60)}分钟前`
  return d.toLocaleString()
}

function openCreate() {
  isEdit.value = false
  editId.value = null
  nodeForm.value = defaultForm()
  formVisible.value = true
}

function openEdit(row: any) {
  isEdit.value = true
  editId.value = row.id
  nodeForm.value = { ...defaultForm(), ...row }
  formVisible.value = true
}

async function handleSave() {
  // 验证必填字段
  if (!nodeForm.value.name) {
    ElMessage.warning('请输入节点名称')
    return
  }
  if (!nodeForm.value.slug) {
    ElMessage.warning('请输入节点标识')
    return
  }
  if (!nodeForm.value.ip) {
    ElMessage.warning('请输入 IP 地址')
    return
  }
  if (!nodeForm.value.region) {
    ElMessage.warning('请输入地区')
    return
  }

  saving.value = true
  try {
    if (isEdit.value && editId.value) {
      await adminApi.updateNode(editId.value, nodeForm.value)
      ElMessage.success('节点已更新')
    } else {
      await adminApi.createNode(nodeForm.value)
      ElMessage.success('节点创建成功')
    }
    formVisible.value = false
    loadNodes()
  } finally {
    saving.value = false
  }
}

async function handleDelete(id: number) {
  await ElMessageBox.confirm('确定删除该节点？删除后该节点上的隧道将无法使用。', '警告', { type: 'warning' })
  await adminApi.deleteNode(id)
  ElMessage.success('删除成功')
  loadNodes()
}

async function showInstallCmd(row: any) {
  const res: any = await adminApi.getInstallCmd(row.id)
  installCmd.value = res.data.command
  cmdVisible.value = true
}

function showUninstallCmd() {
  uninstallCmd.value = `bash <(wget -qO- https://api.jumpfrp.top/uninstall.sh)`
  uninstallVisible.value = true
}

function copyCmd() {
  navigator.clipboard.writeText(installCmd.value)
  ElMessage.success('已复制到剪贴板')
}

function copyUninstallCmd() {
  navigator.clipboard.writeText(uninstallCmd.value)
  ElMessage.success('已复制到剪贴板')
}

onMounted(() => {
  loadNodes()
  refreshTimer = setInterval(loadNodes, 30000)
})
onUnmounted(() => clearInterval(refreshTimer))
</script>
