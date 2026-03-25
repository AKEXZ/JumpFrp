<template>
  <div>
    <h2 style="margin-bottom:20px">系统设置</h2>

    <el-tabs v-model="activeTab">
      <!-- SMTP 邮件配置 -->
      <el-tab-pane label="📧 邮件配置" name="smtp">
        <el-card>
          <el-form :model="smtp" label-width="130px" style="max-width:600px">
            <el-form-item label="启用邮件通知">
              <el-switch v-model="smtp.enabled" />
            </el-form-item>
            <el-divider />
            <el-form-item label="SMTP 服务器">
              <el-input v-model="smtp.host" placeholder="如：smtp.qq.com" :disabled="!smtp.enabled" />
            </el-form-item>
            <el-form-item label="端口">
              <el-input-number v-model="smtp.port" :min="1" :max="65535" :disabled="!smtp.enabled" />
              <span style="margin-left:8px;color:#999;font-size:12px">SSL: 465 / TLS: 587 / 普通: 25</span>
            </el-form-item>
            <el-form-item label="SSL 加密">
              <el-switch v-model="smtp.ssl" :disabled="!smtp.enabled" />
            </el-form-item>
            <el-form-item label="用户名">
              <el-input v-model="smtp.username" placeholder="邮箱账号" :disabled="!smtp.enabled" />
            </el-form-item>
            <el-form-item label="密码 / 授权码">
              <el-input v-model="smtp.password" type="password" show-password
                placeholder="邮箱密码或应用授权码" :disabled="!smtp.enabled" />
            </el-form-item>
            <el-form-item label="发件人地址">
              <el-input v-model="smtp.from" placeholder="noreply@jumpfrp.top" :disabled="!smtp.enabled" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="savingSmtp" @click="saveSmtp">保存配置</el-button>
              <el-button :loading="testingSmtp" :disabled="!smtp.enabled" @click="testSmtpVisible = true">
                发送测试邮件
              </el-button>
            </el-form-item>
          </el-form>

          <!-- 常用邮件服务商快速填写 -->
          <el-divider>常用服务商快速填写</el-divider>
          <div style="display:flex;gap:8px;flex-wrap:wrap">
            <el-button size="small" v-for="p in providers" :key="p.name" @click="fillProvider(p)">
              {{ p.name }}
            </el-button>
          </div>
        </el-card>
      </el-tab-pane>

      <!-- 站点配置 -->
      <el-tab-pane label="🌐 站点配置" name="site">
        <el-card>
          <el-form :model="site" label-width="130px" style="max-width:600px">
            <el-form-item label="站点名称">
              <el-input v-model="site.site_name" placeholder="JumpFrp" />
            </el-form-item>
            <el-form-item label="站点地址">
              <el-input v-model="site.site_url" placeholder="https://jumpfrp.top" />
            </el-form-item>
            <el-form-item label="开放注册">
              <el-switch v-model="site.register_open" />
              <span style="margin-left:8px;color:#999;font-size:12px">关闭后新用户无法注册</span>
            </el-form-item>
            <el-form-item label="ICP 备案号">
              <el-input v-model="site.icp" placeholder="如：京ICP备XXXXXXXX号" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="savingSite" @click="saveSite">保存配置</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <!-- 测试邮件对话框 -->
    <el-dialog v-model="testSmtpVisible" title="发送测试邮件" width="400px">
      <el-form label-width="80px">
        <el-form-item label="收件邮箱">
          <el-input v-model="testEmail" placeholder="输入接收测试邮件的地址" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="testSmtpVisible = false">取消</el-button>
        <el-button type="primary" :loading="testingSmtp" @click="sendTestMail">发送</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { adminApi } from '../../api'

const activeTab = ref('smtp')
const savingSmtp = ref(false)
const savingSite = ref(false)
const testingSmtp = ref(false)
const testSmtpVisible = ref(false)
const testEmail = ref('')

const smtp = ref({
  enabled: false, host: '', port: 587, ssl: false,
  username: '', password: '', from: 'noreply@jumpfrp.top',
})

const site = ref({
  site_name: 'JumpFrp', site_url: 'https://jumpfrp.top',
  register_open: true, icp: '',
})

// 常用邮件服务商预设
const providers = [
  { name: 'QQ 邮箱',    host: 'smtp.qq.com',      port: 465, ssl: true  },
  { name: '163 邮箱',   host: 'smtp.163.com',     port: 465, ssl: true  },
  { name: '126 邮箱',   host: 'smtp.126.com',     port: 465, ssl: true  },
  { name: 'Gmail',      host: 'smtp.gmail.com',   port: 587, ssl: false },
  { name: 'Outlook',    host: 'smtp.office365.com', port: 587, ssl: false },
  { name: 'Aliyun',     host: 'smtpdm.aliyun.com', port: 465, ssl: true  },
]

function fillProvider(p: any) {
  smtp.value.host = p.host
  smtp.value.port = p.port
  smtp.value.ssl  = p.ssl
}

async function loadSettings() {
  const res: any = await adminApi.getSettings()
  smtp.value = { ...smtp.value, ...res.data.smtp }
  site.value = { ...site.value, ...res.data.site }
}

async function saveSmtp() {
  savingSmtp.value = true
  try {
    await adminApi.saveSmtp(smtp.value)
    ElMessage.success('SMTP 配置已保存')
  } finally {
    savingSmtp.value = false
  }
}

async function saveSite() {
  savingSite.value = true
  try {
    await adminApi.saveSite(site.value)
    ElMessage.success('站点配置已保存')
  } finally {
    savingSite.value = false
  }
}

async function sendTestMail() {
  if (!testEmail.value) {
    ElMessage.warning('请输入收件邮箱')
    return
  }
  testingSmtp.value = true
  try {
    await adminApi.testSmtp(testEmail.value)
    ElMessage.success('测试邮件已发送，请检查收件箱')
    testSmtpVisible.value = false
  } finally {
    testingSmtp.value = false
  }
}

onMounted(loadSettings)
</script>
