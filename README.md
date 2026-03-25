# JumpFrp ⚡

> 高速内网穿透托管平台，基于 [frp](https://github.com/fatedier/frp) 构建

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev)
[![Vue](https://img.shields.io/badge/Vue-3-4FC08D?logo=vue.js)](https://vuejs.org)
[![License](https://img.shields.io/badge/License-MIT-blue)](LICENSE)

---

## ✨ 功能特性

| 功能 | 说明 |
|------|------|
| 🌐 **多节点管理** | 添加多台节点服务器，一键安装脚本，实时监控 CPU/内存/连接数 |
| 👤 **用户系统** | 邮箱注册（验证码）、用户名/邮箱登录、找回密码 |
| 🎫 **VIP 制度** | Free / Basic / Pro / Ultimate 四档，差异化端口数/带宽/协议 |
| 🔌 **隧道管理** | TCP/UDP/HTTP/HTTPS 全协议，自动随机分配端口，一键下载 frpc 配置 |
| 🌍 **子域名穿透** | Pro+ 用户支持自定义子域名（需通配符域名） |
| 📊 **实时监控** | 节点 CPU/内存/连接数 30s 自动上报，离线自动检测 |
| 📧 **邮件通知** | 注册验证码、VIP 到期提醒（7/3/1天），SMTP 后台可视化配置 |
| 🛡️ **安全加固** | 全局限流、登录防暴力破解、JWT 认证、安全响应头 |
| ⚙️ **系统设置** | SMTP 邮件配置、站点配置、开放注册开关，全部后台管理 |
| 🚀 **免费部署** | 主控服务支持 Fly.io 免费托管，节点支持 Ubuntu VPS 一键安装 |

---

## 🖥️ 界面预览

**管理后台**
- 仪表盘：用户数/节点数/隧道数/VIP 分布图/节点在线率
- 用户管理：手动添加用户、设置 VIP、重置密码、封禁/解封
- 节点管理：添加/编辑节点、生成安装命令、实时负载监控
- 系统设置：SMTP 邮件配置（支持 QQ/163/Gmail 等快速填写）

**用户前台**
- 首页：产品介绍 + 套餐对比
- 控制台：隧道概览 + VIP 状态
- 隧道管理：创建隧道、下载 frpc 配置、使用教程
- VIP 中心：套餐对比 + 订单记录

---

## 📦 项目结构

```
JumpFrp/
├── master/        # 主控服务 (Go + Gin + SQLite)
│   ├── cmd/       # 入口
│   ├── config/    # 配置
│   └── internal/
│       ├── api/   # 路由（admin + user + agent）
│       ├── middleware/  # JWT / 限流 / 安全头
│       ├── model/       # 数据模型
│       ├── scheduler/   # 定时任务
│       └── service/     # 业务逻辑
├── agent/         # 节点 Agent (Go)
├── frontend/      # 前端 (Vue 3 + Element Plus)
├── scripts/       # install.sh / uninstall.sh
├── docs/          # deployment.md
├── Dockerfile     # 多阶段构建
├── fly.toml       # Fly.io 配置
└── dev.sh         # 本地开发一键启动
```

---

## 🚀 快速开始（本地开发）

### 环境要求

- Go 1.22+
- Node.js 18+

### 一键启动

```bash
git clone https://github.com/yourname/jumpfrp.git
cd jumpfrp
bash dev.sh
```

| 地址 | 说明 |
|------|------|
| http://localhost:5173 | 前台页面 |
| http://localhost:5173/admin | 管理后台 |
| http://localhost:8080/api | API 接口 |

**默认管理员账号：`admin` / `admin123456`**
> ⚠️ 上线前必须修改默认密码！

---

## 🎫 VIP 套餐

| 套餐 | 隧道数 | 端口数 | 带宽 | 协议 | 子域名 | 固定端口 |
|------|--------|--------|------|------|--------|---------|
| **Free** | 1 | 3 | 1 Mbps | TCP | ✗ | ✗ |
| **Basic** | 5 | 10 | 5 Mbps | TCP/UDP | ✗ | ✗ |
| **Pro** | 20 | 50 | 20 Mbps | 全协议 | ✓ | ✗ |
| **Ultimate** | ∞ | 200 | 100 Mbps | 全协议 | ✓ | ✓ |

---

## 🔧 节点安装

在管理后台 → 节点管理 → 点击「安装命令」，复制后在 Ubuntu 服务器上执行：

```bash
bash <(wget -qO- https://api.jumpfrp.top/install.sh) \
  --node-id sh-01 \
  --token <your-token> \
  --master-url https://api.jumpfrp.top
```

**支持系统：** Ubuntu 20.04 / 22.04 / 24.04（amd64 / arm64）

---

## 📧 邮件配置

进入管理后台 → 系统设置 → 邮件配置，填写 SMTP 信息后点击「发送测试邮件」验证。

支持快速填写：QQ 邮箱 / 163 邮箱 / Gmail / Outlook / 阿里云邮件推送

---

## 🌐 生产部署

详见 [docs/deployment.md](docs/deployment.md)

**推荐方案：**
- 主控服务 → [Fly.io](https://fly.io)（永久免费）
- 节点服务器 → 任意 Ubuntu VPS

```bash
# 部署到 Fly.io
fly launch
fly volumes create jumpfrp_data --size 1
fly secrets set JWT_SECRET="your-strong-secret"
fly deploy
```

---

## 📋 API 文档

主要接口：

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/user/auth/register` | 用户注册 |
| POST | `/api/user/auth/login` | 用户登录 |
| GET | `/api/user/tunnels` | 获取隧道列表 |
| POST | `/api/user/tunnels` | 创建隧道 |
| GET | `/api/user/tunnels/:id/frpc-config` | 下载 frpc 配置 |
| GET | `/api/admin/dashboard` | 管理员仪表盘 |
| POST | `/api/admin/nodes` | 添加节点 |
| POST | `/api/admin/settings/smtp` | 保存 SMTP 配置 |

---

## 🛡️ 安全说明

- 密码使用 bcrypt 加密存储
- JWT Token 有效期 72 小时
- 登录接口限流：10 次/分钟/IP
- 全局限流：120 次/分钟/IP
- 节点 Agent 使用独立 Token 鉴权

---

## 📄 License

MIT © 2026 JumpFrp
