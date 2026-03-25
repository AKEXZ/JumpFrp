<template>
  <div class="home">
    <nav>
      <span class="brand">⚡ JumpFrp</span>
      <div>
        <router-link to="/login">登录</router-link>
        <router-link to="/register" class="btn-primary">免费注册</router-link>
      </div>
    </nav>
    <section class="hero">
      <h1>高速内网穿透服务</h1>
      <p>基于 frp 构建，稳定、快速、易用。免费用户即可开始使用。</p>
      <router-link to="/register" class="btn-hero">立即免费使用</router-link>
    </section>
    <section class="plans">
      <h2>选择套餐</h2>
      <div class="plan-grid">
        <div class="plan" v-for="plan in plans" :key="plan.name" :class="{ featured: plan.featured }">
          <div class="plan-name">{{ plan.name }}</div>
          <div class="plan-price">{{ plan.price }}</div>
          <ul>
            <li v-for="f in plan.features" :key="f">✓ {{ f }}</li>
          </ul>
          <router-link to="/register" class="plan-btn">{{ plan.cta }}</router-link>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
const plans = [
  {
    name: 'Free',
    price: '免费',
    features: ['1条隧道', '3个随机端口', '1 Mbps带宽', '仅TCP协议'],
    cta: '免费开始',
    featured: false,
  },
  {
    name: 'Basic',
    price: '¥9.9/月',
    features: ['5条隧道', '10个随机端口', '5 Mbps带宽', 'TCP + UDP'],
    cta: '立即升级',
    featured: false,
  },
  {
    name: 'Pro',
    price: '¥29.9/月',
    features: ['20条隧道', '50个随机端口', '20 Mbps带宽', '全协议支持', '自定义子域名'],
    cta: '推荐选择',
    featured: true,
  },
  {
    name: 'Ultimate',
    price: '¥99.9/月',
    features: ['无限隧道', '200个随机端口', '100 Mbps带宽', '全协议支持', '固定端口申请'],
    cta: '旗舰体验',
    featured: false,
  },
]
</script>

<style scoped>
.home { font-family: sans-serif; }
nav {
  display: flex; justify-content: space-between; align-items: center;
  padding: 16px 40px; background: #1a1a2e; color: white;
}
.brand { font-size: 22px; font-weight: bold; color: #409eff; }
nav a { color: #ccc; text-decoration: none; margin-left: 16px; }
.btn-primary {
  background: #409eff; color: white !important; padding: 6px 16px;
  border-radius: 6px;
}
.hero {
  text-align: center; padding: 80px 20px;
  background: linear-gradient(135deg, #1a1a2e, #0f3460);
  color: white;
}
.hero h1 { font-size: 48px; margin-bottom: 16px; }
.hero p { font-size: 18px; color: #aaa; margin-bottom: 32px; }
.btn-hero {
  background: #409eff; color: white; padding: 14px 40px;
  border-radius: 8px; text-decoration: none; font-size: 18px;
}
.plans { padding: 60px 40px; background: #f5f7fa; text-align: center; }
.plans h2 { font-size: 32px; margin-bottom: 40px; }
.plan-grid { display: flex; gap: 20px; justify-content: center; flex-wrap: wrap; }
.plan {
  background: white; border-radius: 12px; padding: 32px 24px;
  width: 220px; box-shadow: 0 4px 20px rgba(0,0,0,0.08);
}
.plan.featured { border: 2px solid #409eff; transform: scale(1.05); }
.plan-name { font-size: 20px; font-weight: bold; margin-bottom: 8px; }
.plan-price { font-size: 28px; color: #409eff; font-weight: bold; margin-bottom: 20px; }
.plan ul { list-style: none; padding: 0; text-align: left; margin-bottom: 24px; }
.plan li { padding: 4px 0; color: #555; }
.plan-btn {
  display: block; background: #409eff; color: white;
  padding: 10px; border-radius: 6px; text-decoration: none;
}
</style>
