# JumpFrp ⚡

> 高速内网穿透托管平台，基于 [frp](https://github.com/fatedier/frp) 构建

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev)
[![Vue](https://img.shields.io/badge/Vue-3-4FC08D?logo=vue.js)](https://vuejs.org)
[![frp](https://img.shields.io/badge/frp-0.61.0-orange)](https://github.com/fatedier/frp)
[![License](https://img.shields.io/badge/License-MIT-blue)](LICENSE)

---

## ✨ 功能特性

| 功能 | 说明 |
|------|------|
| 🌐 **多节点管理** | 添加多台节点服务器，一键安装/卸载，实时监控 CPU/内存/连接数 |
| 👤 **用户系统** | 邮箱注册（验证码）、用户名/邮箱登录、找回密码 |
| 🎫 **VIP 制度** | Free / Basic / Pro / Ultimate 四档，差异化端口数/带宽/协议 |
| 🔌 **隧道管理** | TCP/UDP/HTTP/HTTPS 全协议，自动分配端口，绑定节点/域名，一键下载 frpc 配置 |
| 🌍 **Web 穿透** | 支持 HTTP/HTTPS 隧道映射 Web 网站，Pro+ 支持自定义子域名 |
| 🔗 **域名管理** | 用户绑定自定义子域名（Pro+），管理员审批/管理所有域名 |
| 📊 **实时监控** | 节点 CPU/内存/连接数 30s 自动上报，离线自动检测 |
| 📧 **邮件通知** | 注册验证码、VIP 到期提醒，SMTP 后台可视化配置 |
| 🛡️ **安全加固** | 全局限流、登录防暴力破解、JWT 认证、安全响应头 |
| ⚙️ **系统设置** | SMTP / 站点配置全部后台管理，无需改配置文件 |
| 🚀 **免费部署** | 主控支持 Fly.io 免费托管，节点一键安装脚本 |

---

## 🖥️ 管理后台

| 页面 | 功能 |
|------|------|
| 仪表盘 | 用户数/节点数/隧道数/VIP 分布/节点在线率 |
| 用户管理 | 手动添加用户、设置 VIP、重置密码、封禁/解封 |
| 节点管理 | 添加/编辑节点、**安装命令**、**卸载命令**、实时负载监控 |
| 系统设置 | SMTP 邮件配置（QQ/163/Gmail 快速填写）、站点配置 |

---

## 📦 项目结构

```
JumpFrp/
├── master/        # 主控服务 (Go + Gin + SQLite)
├── agent/         # 节点 Agent (Go)
├── frontend/      # 前端 (Vue 3 + Element Plus)
├── scripts/
│   ├── install.sh    # 节点一键安装（含进度显示）
│   └── uninstall.sh  # 节点卸载
├── Dockerfile     # 多阶段构建（Node + Go + Alpine）
├── fly.toml       # Fly.io 配置
└── dev.sh         # 本地开发一键启动
```

---

## 🚀 快速开始（本地开发）

```bash
git clone https://github.com/AKEXZ/JumpFrp.git
cd JumpFrp
bash dev.sh
```

| 地址 | 说明 |
|------|------|
| http://localhost:5173 | 前台页面 |
| http://localhost:5173/admin | 管理后台 |
| http://localhost:8080/api | API 接口 |

**默认管理员：`admin` / `admin123456`**
> ⚠️ 上线前必须修改默认密码！

---

## 🎫 VIP 套餐

| 套餐 | 隧道数 | 端口数 | 带宽 | 协议 | 子域名 |
|------|--------|--------|------|------|--------|
| **Free** | 1 | 3 | 1 Mbps | TCP | ✗ |
| **Basic** | 5 | 10 | 5 Mbps | TCP/UDP | ✗ |
| **Pro** | 20 | 50 | 20 Mbps | 全协议 | ✓ |
| **Ultimate** | ∞ | 200 | 100 Mbps | 全协议 | ✓ |

---

## 🔧 节点管理

### 安装节点

在管理后台 → 节点管理 → 填写节点信息 → 点击「安装」复制命令，在 Ubuntu 服务器上执行：

```bash
bash <(wget -qO- https://api.jumpfrp.top/install.sh) \
  --node-id sh-01 \
  --token <your-token> \
  --master-url https://api.jumpfrp.top \
  --frps-port 7000 \
  --agent-port 7500
```

安装过程会显示下载进度和服务启动状态。

### 卸载节点

在管理后台 → 节点管理 → 点击「卸载」复制命令，在服务器上执行：

```bash
bash <(wget -qO- https://api.jumpfrp.top/uninstall.sh)
```

**支持系统：** Ubuntu 20.04 / 22.04 / 24.04（amd64 / arm64）

> ⚠️ 节点标识（slug）必须填写，创建后不可修改，安装命令依赖此字段

---

## 🌐 Web 网站穿透

创建隧道时选择 HTTP 或 HTTPS 协议，填写本地端口（如 80/443/8080），即可将本地 Web 服务暴露到公网。

- **Free/Basic**：分配随机端口，通过 `节点IP:端口` 访问
- **Pro/Ultimate**：可绑定子域名，通过 `xxx.jumpfrp.top` 访问

---

## 📧 邮件配置

管理后台 → 系统设置 → 邮件配置，支持快速填写：

| 服务商 | SMTP 地址 | 端口 | SSL |
|--------|-----------|------|-----|
| QQ 邮箱 | smtp.qq.com | 465 | ✓ |
| 163 邮箱 | smtp.163.com | 465 | ✓ |
| Gmail | smtp.gmail.com | 587 | ✗ |
| Outlook | smtp.office365.com | 587 | ✗ |

---

## 🌐 生产部署（Fly.io）

```bash
fly auth login
fly apps create jumpfrp
fly volumes create jumpfrp_data --size 1 --region sin --app jumpfrp
fly secrets set JWT_SECRET="$(openssl rand -base64 48)" --app jumpfrp
fly secrets set GIN_MODE="release" --app jumpfrp
fly deploy --app jumpfrp
```

### ⚠️ 已知坑（避免踩雷）

| 问题 | 原因 | 解决 |
|------|------|------|
| SQLite CGO 报错 | 默认驱动需要 CGO | ✅ 已用 `glebarez/sqlite`（纯 Go） |
| 前端 404 | 镜像无前端文件 | ✅ Dockerfile 多阶段构建 |
| API 请求 localhost | 生产用了开发地址 | ✅ `.env.production` 用相对路径 `/api` |
| hkg 区域不可用 | Fly.io 已废弃 | ✅ 改用 `sin`（新加坡） |
| go.mod 版本冲突 | 本地 Go 版本高于镜像 | ✅ Dockerfile 用 `golang:1.24-alpine` |
| fly.toml 格式错误 | `[[http_service]]` 写法错误 | ✅ 改为 `[http_service]` |

---

## 🔩 节点安装注意事项

| 问题 | 原因 | 解决 |
|------|------|------|
| frps 启动失败 | frp 0.61.0 改用 TOML 格式 | ✅ 生成 `frps.toml`，不再用 INI |
| Agent Exec format error | 下载到 HTML 而非二进制 | ✅ 从 GitHub Releases 下载 |
| GitHub 下载卡住 | 服务器无法访问 GitHub | ✅ 优先用 `gitproxy.ake.cx` 代理 |
| node-id 为空 | 创建节点未填 slug | ✅ 前端标注必填，后端校验 |

---

## 📦 Agent 发布（维护者）

每次更新 Agent 代码后需重新编译并上传到 GitHub Releases：

```bash
cd agent
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o jumpfrp-agent-linux-amd64 ./cmd/main.go
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o jumpfrp-agent-linux-arm64 ./cmd/main.go
# 上传到 GitHub Releases: https://github.com/AKEXZ/JumpFrp/releases
```

---

## 📋 主要 API

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/user/auth/register` | 用户注册 |
| POST | `/api/user/auth/login` | 用户登录 |
| GET | `/api/user/nodes` | 可用节点列表 |
| GET | `/api/user/tunnels` | 隧道列表 |
| POST | `/api/user/tunnels` | 创建隧道 |
| DELETE | `/api/user/tunnels/:id` | 删除隧道 |
| GET | `/api/user/tunnels/:id/frpc-config` | 下载 frpc 配置 |
| GET | `/api/user/subdomains` | 我的域名列表 |
| POST | `/api/user/subdomains` | 绑定域名（Pro+） |
| DELETE | `/api/user/subdomains/:id` | 解绑域名 |
| GET | `/api/admin/dashboard` | 仪表盘统计 |
| GET/POST/PUT/DELETE | `/api/admin/users/*` | 用户管理 |
| GET/POST/PUT/DELETE | `/api/admin/nodes/*` | 节点管理 |
| GET | `/api/admin/subdomains` | 域名管理列表 |
| POST | `/api/admin/subdomains` | 手动添加域名 |
| PUT | `/api/admin/subdomains/:id/approve` | 审批域名 |
| POST | `/api/admin/settings/smtp/test` | 发送测试邮件 |

---

## 🛡️ 安全

- 密码 bcrypt 加密
- JWT Token 有效期 72 小时
- 登录限流：10 次/分钟/IP
- 全局限流：120 次/分钟/IP
- 节点 Agent 独立 Token 鉴权

---

## 📄 License

MIT © 2026 JumpFrp
