<template>
  <div class="vip-page">
    <h2 style="margin-bottom:20px">VIP 中心</h2>

    <!-- 当前套餐状态 -->
    <el-card class="current-vip" style="margin-bottom:24px">
      <div class="vip-status">
        <div>
          <div class="vip-badge" :class="`vip-${user?.vip_level}`">
            {{ vipNames[user?.vip_level || 0] }}
          </div>
          <div style="margin-top:8px;color:#666">
            <span v-if="user?.vip_level === 0">免费用户</span>
            <span v-else-if="user?.vip_expire_at">
              到期时间：{{ formatDate(user.vip_expire_at) }}
              <el-tag v-if="isExpiringSoon" type="warning" size="small" style="margin-left:8px">即将到期</el-tag>
            </span>
          </div>
        </div>
        <div class="quota-grid">
          <div class="quota-item">
            <div class="quota-val">{{ currentQuota.MaxTunnels === 9999 ? '∞' : currentQuota.MaxTunnels }}</div>
            <div class="quota-label">隧道数</div>
          </div>
          <div class="quota-item">
            <div class="quota-val">{{ currentQuota.MaxPorts }}</div>
            <div class="quota-label">端口数</div>
          </div>
          <div class="quota-item">
            <div class="quota-val">{{ currentQuota.MaxBandwidth }}M</div>
            <div class="quota-label">带宽</div>
          </div>
          <div class="quota-item">
            <div class="quota-val">{{ currentQuota.Protocols?.length }}</div>
            <div class="quota-label">协议数</div>
          </div>
        </div>
      </div>
    </el-card>

    <!-- 套餐对比 -->
    <h3 style="margin-bottom:16px">升级套餐</h3>
    <div class="plan-grid">
      <div v-for="plan in plans" :key="plan.Level"
        class="plan-card"
        :class="{ current: plan.Level === user?.vip_level, featured: plan.Level === 2 }">
        <div v-if="plan.Level === 2" class="plan-badge">推荐</div>
        <div class="plan-name">{{ plan.Name }}</div>
        <div class="plan-price">
          <span class="price-num">¥{{ plan.Price }}</span>
          <span class="price-unit">/月</span>
        </div>
        <div class="plan-desc">{{ plan.Description }}</div>
        <ul class="plan-features">
          <li v-for="f in getPlanFeatures(plan.Level)" :key="f">
            <el-icon color="#67c23a"><Check /></el-icon> {{ f }}
          </li>
        </ul>
        <el-button
          v-if="plan.Level !== user?.vip_level"
          type="primary"
          :plain="plan.Level < (user?.vip_level || 0)"
          style="width:100%;margin-top:16px"
          @click="openBuy(plan)">
          {{ plan.Level < (user?.vip_level || 0) ? '降级' : '立即升级' }}
        </el-button>
        <el-button v-else disabled style="width:100%;margin-top:16px">当前套餐</el-button>
      </div>
    </div>

    <!-- 订单记录 -->
    <el-card style="margin-top:32px">
      <template #header>订单记录</template>
      <el-empty v-if="orders.length === 0" description="暂无订单" />
      <el-table v-else :data="orders" border>
        <el-table-column prop="id" label="订单号" width="80" />
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
            <el-tag :type="({ paid: 'success', pending: 'warning', cancelled: 'info' } as Record<string,any>)[row.status]">
              {{ ({ paid: '已完成', pending: '待支付', cancelled: '已取消' } as Record<string,string>)[row.status] }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="expire_at" label="到期时间" width="160">
          <template #default="{ row }">{{ row.expire_at ? formatDate(row.expire_at) : '-' }}</template>
        </el-table-column>
        <el-table-column label="创建时间">
          <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 购买对话框 -->
    <el-dialog v-model="buyVisible" :title="`升级至 ${selectedPlan?.Name}`" width="440px">
      <div v-if="selectedPlan" class="buy-dialog">
        <el-alert type="info" :closable="false" style="margin-bottom:16px">
          当前为手动开通模式，请联系管理员完成支付后开通。
        </el-alert>
        <el-form label-width="80px">
          <el-form-item label="套餐">
            <strong>{{ selectedPlan.Name }}</strong>
          </el-form-item>
          <el-form-item label="时长">
            <el-radio-group v-model="buyDays">
              <el-radio-button :value="30">1个月 ¥{{ selectedPlan.Price }}</el-radio-button>
              <el-radio-button :value="90">3个月 ¥{{ (selectedPlan.Price * 3 * 0.9).toFixed(1) }}</el-radio-button>
              <el-radio-button :value="365">1年 ¥{{ (selectedPlan.Price * 12 * 0.8).toFixed(1) }}</el-radio-button>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="金额">
            <span class="buy-price">¥{{ calcPrice(selectedPlan, buyDays) }}</span>
          </el-form-item>
        </el-form>
        <el-divider />
        <div style="text-align:center;color:#999;font-size:13px">
          请联系管理员，报告您的用户名和所选套餐，管理员将为您手动开通。
        </div>
      </div>
      <template #footer>
        <el-button @click="buyVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { userApi } from '../../api'
import { useAuthStore } from '../../stores/auth'

const auth = useAuthStore()
const user = computed(() => auth.user)

const plans = ref<any[]>([])
const orders = ref<any[]>([])
const buyVisible = ref(false)
const selectedPlan = ref<any>(null)
const buyDays = ref(30)

const vipNames: Record<number, string> = { 0: 'Free', 1: 'Basic', 2: 'Pro', 3: 'Ultimate' }

const quotaMap: Record<number, any> = {
  0: { MaxTunnels: 1, MaxPorts: 3, MaxBandwidth: 1, Protocols: ['tcp'] },
  1: { MaxTunnels: 5, MaxPorts: 10, MaxBandwidth: 5, Protocols: ['tcp', 'udp'] },
  2: { MaxTunnels: 20, MaxPorts: 50, MaxBandwidth: 20, Protocols: ['tcp', 'udp', 'http', 'https'] },
  3: { MaxTunnels: 9999, MaxPorts: 200, MaxBandwidth: 100, Protocols: ['tcp', 'udp', 'http', 'https'] },
}

const currentQuota = computed(() => quotaMap[user.value?.vip_level || 0])

const isExpiringSoon = computed(() => {
  if (!user.value?.vip_expire_at) return false
  const diff = new Date(user.value.vip_expire_at).getTime() - Date.now()
  return diff > 0 && diff < 7 * 24 * 3600 * 1000
})

function getPlanFeatures(level: number): string[] {
  const q = quotaMap[level]
  const features = [
    `${q.MaxTunnels === 9999 ? '无限' : q.MaxTunnels} 条隧道`,
    `${q.MaxPorts} 个随机端口`,
    `${q.MaxBandwidth} Mbps 带宽`,
    `协议：${q.Protocols.map((p: string) => p.toUpperCase()).join(' / ')}`,
  ]
  if (level >= 2) features.push('自定义子域名')
  if (level >= 3) features.push('固定端口申请')
  return features
}

function formatDate(d: string) {
  return new Date(d).toLocaleString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

function calcPrice(plan: any, days: number) {
  const monthly = plan.Price
  if (days === 30) return monthly.toFixed(2)
  if (days === 90) return (monthly * 3 * 0.9).toFixed(2)
  if (days === 365) return (monthly * 12 * 0.8).toFixed(2)
  return (monthly * days / 30).toFixed(2)
}

function openBuy(plan: any) {
  selectedPlan.value = plan
  buyDays.value = 30
  buyVisible.value = true
}

onMounted(async () => {
  const [planRes, orderRes]: any[] = await Promise.all([
    userApi.getVIPPlans(),
    userApi.getVIPOrders(),
  ])
  plans.value = planRes.data
  orders.value = orderRes.data
})
</script>

<style scoped>
.vip-status { display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 16px; }
.vip-badge {
  display: inline-block; padding: 6px 20px; border-radius: 20px;
  font-size: 20px; font-weight: bold; color: white;
}
.vip-0 { background: #909399; }
.vip-1 { background: #409eff; }
.vip-2 { background: linear-gradient(135deg, #f6a623, #f05a28); }
.vip-3 { background: linear-gradient(135deg, #8e44ad, #3498db); }

.quota-grid { display: flex; gap: 24px; }
.quota-item { text-align: center; }
.quota-val { font-size: 24px; font-weight: bold; color: #409eff; }
.quota-label { font-size: 12px; color: #999; }

.plan-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 20px; }
.plan-card {
  position: relative; background: white; border-radius: 12px;
  padding: 28px 24px; border: 2px solid #eee;
  transition: transform 0.2s, box-shadow 0.2s;
}
.plan-card:hover { transform: translateY(-4px); box-shadow: 0 8px 30px rgba(0,0,0,0.1); }
.plan-card.featured { border-color: #409eff; }
.plan-card.current { border-color: #67c23a; }
.plan-badge {
  position: absolute; top: -12px; left: 50%; transform: translateX(-50%);
  background: #409eff; color: white; padding: 2px 16px;
  border-radius: 12px; font-size: 12px;
}
.plan-name { font-size: 22px; font-weight: bold; margin-bottom: 8px; }
.plan-price { margin-bottom: 8px; }
.price-num { font-size: 32px; font-weight: bold; color: #409eff; }
.price-unit { color: #999; }
.plan-desc { color: #666; font-size: 13px; margin-bottom: 16px; }
.plan-features { list-style: none; padding: 0; }
.plan-features li { padding: 4px 0; font-size: 14px; display: flex; align-items: center; gap: 6px; }
.buy-price { font-size: 24px; font-weight: bold; color: #f05a28; }
</style>
