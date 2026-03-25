<template>
  <div>
    <h2 style="margin-bottom:20px">控制台</h2>

    <!-- 用户信息卡片 -->
    <el-row :gutter="16" style="margin-bottom:20px">
      <el-col :span="8">
        <el-card shadow="hover">
          <div style="display:flex;align-items:center;gap:12px">
            <el-avatar :size="48" style="background:#409eff">{{ user?.username?.[0]?.toUpperCase() }}</el-avatar>
            <div>
              <div style="font-weight:bold;font-size:16px">{{ user?.username }}</div>
              <el-tag :type="['info','','warning','danger'][user?.vip_level || 0]">
                {{ ['Free','Basic','Pro','Ultimate'][user?.vip_level || 0] }}
              </el-tag>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover">
          <el-statistic title="我的隧道" :value="tunnels.length">
            <template #suffix>/ {{ quotas[user?.vip_level || 0]?.maxTunnels }}</template>
          </el-statistic>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover">
          <el-statistic title="在线隧道" :value="activeTunnels" />
        </el-card>
      </el-col>
    </el-row>

    <!-- 快速操作 -->
    <el-card style="margin-bottom:20px">
      <template #header>快速操作</template>
      <div style="display:flex;gap:12px">
        <el-button type="primary" @click="$router.push('/tunnels')">管理隧道</el-button>
        <el-button @click="$router.push('/vip')">升级 VIP</el-button>
      </div>
    </el-card>

    <!-- 最近隧道 -->
    <el-card>
      <template #header>
        <div style="display:flex;justify-content:space-between">
          <span>最近隧道</span>
          <el-button text @click="$router.push('/tunnels')">查看全部</el-button>
        </div>
      </template>
      <el-empty v-if="tunnels.length === 0" description="还没有隧道" />
      <el-table v-else :data="tunnels.slice(0,5)" :show-header="false">
        <el-table-column prop="name" />
        <el-table-column width="100">
          <template #default="{ row }">
            <el-tag size="small">{{ row.protocol?.toUpperCase() }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column width="120">
          <template #default="{ row }">{{ row.node?.name }}</template>
        </el-table-column>
        <el-table-column width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'info'" size="small">
              {{ row.status === 'active' ? '在线' : '离线' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { userApi } from '../../api'
import { useAuthStore } from '../../stores/auth'

const auth = useAuthStore()
const user = computed(() => auth.user)
const tunnels = ref<any[]>([])

const quotas: Record<number, any> = {
  0: { maxTunnels: 1 },
  1: { maxTunnels: 5 },
  2: { maxTunnels: 20 },
  3: { maxTunnels: 9999 },
}

const activeTunnels = computed(() => tunnels.value.filter(t => t.status === 'active').length)

onMounted(async () => {
  const res: any = await userApi.listTunnels()
  tunnels.value = res.data
})
</script>
