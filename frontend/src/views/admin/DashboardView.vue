<template>
  <div>
    <h2 style="margin-bottom:20px">仪表盘</h2>

    <!-- 核心指标 -->
    <el-row :gutter="16" style="margin-bottom:20px">
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <el-statistic title="总用户数" :value="stats.users">
            <template #prefix><el-icon color="#409eff"><User /></el-icon></template>
          </el-statistic>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <el-statistic title="节点总数" :value="stats.nodes">
            <template #suffix style="font-size:14px;color:#67c23a">
              ({{ stats.online_nodes }} 在线)
            </template>
          </el-statistic>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <el-statistic title="隧道总数" :value="stats.tunnels" />
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <el-statistic title="付费用户" :value="(stats.vip_dist?.basic || 0) + (stats.vip_dist?.pro || 0) + (stats.vip_dist?.ultimate || 0)" />
        </el-card>
      </el-col>
    </el-row>

    <!-- VIP 分布 -->
    <el-row :gutter="16">
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>VIP 用户分布</template>
          <div class="vip-dist">
            <div v-for="item in vipDist" :key="item.name" class="vip-dist-item">
              <div class="vip-dist-label">
                <el-tag :type="item.type" size="small">{{ item.name }}</el-tag>
              </div>
              <el-progress
                :percentage="stats.users ? Math.round(item.count / stats.users * 100) : 0"
                :color="item.color"
                :stroke-width="16"
                style="flex:1"
              />
              <span class="vip-dist-count">{{ item.count }}</span>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>节点状态</template>
          <div class="node-status">
            <div class="node-status-item">
              <div class="node-dot online" />
              <span>在线</span>
              <strong>{{ stats.online_nodes }}</strong>
            </div>
            <div class="node-status-item">
              <div class="node-dot offline" />
              <span>离线</span>
              <strong>{{ (stats.nodes || 0) - (stats.online_nodes || 0) }}</strong>
            </div>
          </div>
          <el-progress
            :percentage="stats.nodes ? Math.round((stats.online_nodes || 0) / stats.nodes * 100) : 0"
            status="success"
            :stroke-width="20"
            style="margin-top:16px"
          />
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { adminApi } from '../../api'

const stats = ref<any>({})

const vipDist = computed(() => [
  { name: 'Free',     count: stats.value.vip_dist?.free     || 0, type: 'info',    color: '#909399' },
  { name: 'Basic',    count: stats.value.vip_dist?.basic    || 0, type: '',        color: '#409eff' },
  { name: 'Pro',      count: stats.value.vip_dist?.pro      || 0, type: 'warning', color: '#e6a23c' },
  { name: 'Ultimate', count: stats.value.vip_dist?.ultimate || 0, type: 'danger',  color: '#8e44ad' },
])

onMounted(async () => {
  const res: any = await adminApi.dashboard()
  stats.value = res.data
})
</script>

<style scoped>
.stat-card { text-align: center; }
.vip-dist { display: flex; flex-direction: column; gap: 12px; }
.vip-dist-item { display: flex; align-items: center; gap: 12px; }
.vip-dist-label { width: 70px; }
.vip-dist-count { width: 30px; text-align: right; font-weight: bold; }
.node-status { display: flex; gap: 40px; justify-content: center; padding: 16px 0; }
.node-status-item { display: flex; align-items: center; gap: 8px; font-size: 16px; }
.node-dot { width: 12px; height: 12px; border-radius: 50%; }
.node-dot.online { background: #67c23a; }
.node-dot.offline { background: #f56c6c; }
</style>
