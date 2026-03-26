# JumpFrp — 内网穿透托管平台 项目规划方案

> 版本：v1.5
> 日期：2026-03-26
> 状态：✅ 全部完成

---

## 一、项目概述

基于 [fatedier/frp](https://github.com/fatedier/frp) 服务端进行二次开发，构建一个多节点、多用户的内网穿透托管平台，品牌名 **JumpFrp**。用户通过前台注册账号，选择节点服务器，创建隧道配置，使用官方 frpc 客户端连接。平台支持 VIP 会员制度，管理员可通过后台管理服务器节点、用户和权限。

---

## 二、系统架构

```
┌─────────────────────────────────────────────────────────┐
│                     用户浏览器                           │
│              前台页面 + 管理员后台                        │
└──────────────────────┬──────────────────────────────────┘
                       │ HTTP/HTTPS
┌──────────────────────▼──────────────────────────────────┐
│                   主控服务 (Master)                       │
│         Go + Gin + SQLite + Vue 3 前端                   │
│   - 用户认证 / VIP 管理 / 隧道配置下发                    │
│   - 节点管理 / 一键安装脚本生成                           │
│   - 系统设置（SMTP / 站点配置）                           │
│   - 限流 / 安全加固 / 邮件通知                            │
└──────────┬───────────────────────┬──────────────────────┘
           │ API 通信               │ API 通信
┌──────────▼──────────┐  ┌─────────▼──────────────────────┐
│   节点服务器 A       │  │   节点服务器 B / C / ...        │
│  frps 0.61.0        │  │  frps 0.61.0                    │
│  jumpfrp-agent      │  │  jumpfrp-agent                  │
│  Ubuntu + systemd   │  │  Ubuntu + systemd               │
└─────────────────────┘  └────────────────────────────────┘
           ▲                         ▲
           │ frpc 连接               │ frpc 连接
    ┌──────┴──────┐           ┌──────┴──────┐
    │  用户设备 A  │           │  用户设备 B  │
    │ 官方 frpc   │           │ 官方 frpc   │
    └─────────────┘           └─────────────┘
```

---

## 三、技术栈

| 层级 | 技术选型 | 说明 |
|------|---------|------|
| **后端主控** | Go 1.22 + Gin | 高性能，与 frp 同语言 |
| **数据库** | SQLite (glebarez/sqlite) | **纯 Go 实现，无需 CGO**，Fly.io 兼容 |
| **前端** | Vue 3 + Vite + Element Plus | 成熟生态，后台管理友好 |
| **节点 Agent** | Go 轻量 HTTP 服务 | 部署在每台节点，接收主控指令 |
| **一键安装** | Shell 脚本 (Bash) | Ubuntu 专用，systemd 服务注册 |
| **认证** | JWT Token | 前后端分离认证 |
| **邮件** | SMTP（后台可配置） | HTML 模板邮件 |
| **部署** | Fly.io（主控）+ Ubuntu VPS（节点） | 主控免费托管 |

---

## 四、VIP 等级设计

| 等级 | 名称 | 隧道数 | 端口数 | 带宽 | 协议 | 子域名 | 固定端口 |
|------|------|--------|--------|------|------|--------|---------|
| 0 | Free | 1 | 3 | 1 Mbps | TCP | ✗ | ✗ |
| 1 | Basic | 5 | 10 | 5 Mbps | TCP/UDP | ✗ | ✗ |
| 2 | Pro | 20 | 50 | 20 Mbps | 全协议 | ✓ | ✗ |
| 3 | Ultimate | 无限 | 200 | 100 Mbps | 全协议 | ✓ | ✓ |

---

## 五、功能模块详细设计

### 5.1 前台（用户端）

| 页面 | 功能 |
|------|------|
| 首页 | 产品介绍、套餐展示、快速开始 |
| 注册/登录 | 邮箱注册（验证码）、用户名或邮箱登录、找回密码 |
| 控制台 | 隧道概览、VIP 状态、快速操作 |
| 隧道管理 | 创建/编辑/删除隧道、绑定节点、绑定子域名、下载 frpc 配置、使用教程 |
| VIP 中心 | 套餐对比、当前权益、订单记录 |
| 个人设置 | 修改密码 |

### 5.2 管理员后台

| 页面 | 功能 |
|------|------|
| 仪表盘 | 用户数、节点数、隧道数、VIP 分布图、节点在线率 |
| 用户管理 | 列表搜索、手动添加用户、设置 VIP、重置密码、封禁/解封 |
| 节点管理 | 添加/编辑/删除节点、**安装命令**、**卸载命令**、实时监控（CPU/内存/连接数） |
| 隧道管理 | 全局隧道列表、强制删除 |
| VIP 订单 | 订单列表（按状态筛选）、手动开通 VIP |
| 系统设置 | SMTP 邮件配置（后台可视化）、站点配置、开放注册开关 |

### 5.3 节点服务器管理

节点操作按钮：**编辑 | 安装 | 卸载 | 删除**

节点信息字段（完整）：

| 字段 | 说明 |
|------|------|
| 节点名称 / 标识 | 显示名 + 唯一 slug（必填，创建后不可修改） |
| IP 地址 / 地区 | 公网 IP + 所在地区 |
| frps 端口 / Agent 端口 | 默认 7000 / 7500 |
| Agent Token | 主控与节点通信密钥（自动生成） |
| 端口池范围 | 起始 - 结束，用户随机分配范围 |
| 排除端口 | 逗号分隔，不参与随机分配 |
| 最低 VIP 等级 | 控制哪些用户可使用此节点 |
| 带宽上限 / 最大连接数 | 节点容量限制 |
| 节点状态 | 在线 / 离线 / 维护（可手动设置） |
| 实时监控 | CPU / 内存 / 当前连接数 / 最后心跳时间 |
| 备注 | 内部备注 |

### 5.4 一键安装 / 卸载脚本

**安装**（从管理后台复制命令）：
```bash
bash <(wget -qO- https://api.jumpfrp.top/install.sh) \
  --node-id sh-01 \
  --token xxxxxxxx \
  --master-url https://api.jumpfrp.top \
  --frps-port 7000 \
  --agent-port 7500
```

**卸载**：
```bash
bash <(wget -qO- https://api.jumpfrp.top/uninstall.sh)
```

脚本安装内容：
1. 检测系统（Ubuntu 20.04+）及架构（amd64 / arm64）
2. 通过 `gitproxy.ake.cx` 代理下载 frps 0.61.0（显示下载进度）
3. 从 GitHub Releases 下载 jumpfrp-agent 二进制
4. 生成 `frps.toml` 配置文件（TOML 格式，frp 0.61.0+）
5. 注册 systemd 服务（`frps.service` + `jumpfrp-agent.service`）
6. 启动服务并显示状态，提示防火墙端口开放

---

## 六、数据库设计（SQLite）

```sql
users          -- 用户表（含 VIP 等级、到期时间、API Token）
nodes          -- 节点表（含端口池、监控数据、Agent Token）
tunnels        -- 隧道表（含协议、端口分配、带宽限制）
vip_orders     -- VIP 订单表
admin_logs     -- 管理员操作日志
subdomains     -- 子域名申请表
traffic_logs   -- 流量统计表（按天聚合）
system_configs -- 系统配置表（KV 存储，含 SMTP / 站点配置）
```

---

## 七、项目目录结构

```
JumpFrp/
├── master/                    # 主控服务 (Go)
│   ├── cmd/server/main.go
│   ├── config/
│   ├── internal/
│   │   ├── api/admin/         # 用户/节点/隧道/VIP/设置
│   │   ├── api/user/          # 认证/隧道/VIP
│   │   ├── middleware/        # JWT / 限流 / 安全头 / CORS
│   │   ├── model/             # 数据模型
│   │   ├── scheduler/         # 定时任务
│   │   └── service/           # auth/tunnel/vip/mail/system
│   └── web/                   # 前端构建产物
├── agent/                     # 节点 Agent (Go)
├── frontend/                  # 前端 (Vue 3)
│   └── src/views/
│       ├── user/              # 首页/登录/注册/控制台/隧道/VIP
│       └── admin/             # 仪表盘/用户/节点/隧道/订单/设置
├── scripts/
│   ├── install.sh             # 节点一键安装（含进度显示）
│   └── uninstall.sh           # 节点卸载
├── docs/deployment.md
├── Dockerfile                 # 多阶段构建（Node + Go + Alpine）
├── fly.toml
├── dev.sh
└── README.md
```

---

## 八、安全设计

| 措施 | 实现 |
|------|------|
| JWT 认证 | 所有需登录接口验证 Token |
| 全局限流 | 120 次/分钟/IP |
| 登录限流 | 10 次/分钟/IP，防暴力破解 |
| 安全响应头 | X-Frame-Options / X-XSS-Protection / X-Content-Type-Options |
| Agent 鉴权 | 节点与主控通信使用独立 Token |
| 密码加密 | bcrypt 哈希存储 |
| CORS | 跨域请求控制 |

---

## 九、部署方案

| 组件 | 部署位置 | 说明 |
|------|---------|------|
| 主控服务 | Fly.io（免费） | SQLite 持久化 Volume，sin 区域 |
| 前端 | 嵌入主控服务 | 静态文件由 Go 服务托管，SPA fallback |
| 节点服务器 | Ubuntu VPS | 一键安装脚本部署 |

**域名配置：**
- `jumpfrp.top` → 主控服务（前台 + 后台）
- `api.jumpfrp.top` → 主控 API
- `*.jumpfrp.top` → 节点服务器（子域名穿透）

---

## 十、Fly.io 部署注意事项（重要）

### 10.1 已踩过的坑

| 坑 | 原因 | 解决方案 |
|----|------|---------|
| SQLite CGO 报错 | `gorm.io/driver/sqlite` 依赖 CGO，Alpine 无 gcc | 换 `github.com/glebarez/sqlite`（纯 Go） |
| 前端 404 | Dockerfile 未复制前端产物 | 多阶段构建，Node 阶段构建前端 |
| API 请求 localhost | 生产环境用了开发 API 地址 | `.env.production` 设置 `VITE_API_URL=/api` |
| 健康检查超时 | 自定义 check 配置太严格 | 删除 `[checks]`，用 `[http_service]` 默认检查 |
| hkg 区域废弃 | Fly.io 已弃用香港区域 | 改用 `sin`（新加坡） |
| 构建超时 | Depot builder 网络抖动 | 重试或加 `--builder local` |
| go.mod 版本冲突 | 本地 Go 1.25+，Docker 镜像 Go 1.22 | Dockerfile 改用 `golang:1.24-alpine` |
| fly.toml 格式错误 | `[[http_service]]` 应为 `[http_service]` | 改为单括号 |

### 10.2 完整部署命令

```bash
fly auth login
fly apps create jumpfrp
fly volumes create jumpfrp_data --size 1 --region sin --app jumpfrp
fly secrets set JWT_SECRET="$(openssl rand -base64 48)" --app jumpfrp
fly secrets set GIN_MODE="release" --app jumpfrp
fly deploy --app jumpfrp
curl https://jumpfrp.fly.dev/health
```

### 10.3 部署前检查清单

- [ ] `fly.toml` 中 `primary_region = "sin"`（不是 hkg）
- [ ] `Dockerfile` 包含 Node + Go 多阶段构建
- [ ] `frontend/.env.production` 存在且 `VITE_API_URL=/api`
- [ ] 已创建 Volume
- [ ] 已设置 JWT_SECRET

---

## 十一、节点安装注意事项（重要）

### 11.1 已踩过的坑

| 坑 | 原因 | 解决方案 |
|----|------|---------|
| frps 启动失败 | frp 0.61.0 改用 TOML 格式，旧 INI 不兼容 | 生成 `frps.toml`，字段名全部更新 |
| Agent Exec format error | 主控 `/download/agent` 返回 HTML 而非二进制 | 从 GitHub Releases 下载正确二进制 |
| GitHub 下载卡住 | 节点服务器无法直连 GitHub | 优先使用 `gitproxy.ake.cx` 代理 |
| node-id 为空 | 创建节点时未填写 slug | 后端自动生成，前端标注必填 |
| install.sh 参数解析错误 | `--node-id` 后面没有值 | 创建节点时 slug 必填，命令才完整 |

### 11.2 frps.toml 配置格式（frp 0.61.0+）

```toml
bindPort = 7000
auth.method = "token"
auth.token = "your-token"

webServer.addr = "0.0.0.0"
webServer.port = 7001
webServer.user = "admin"
webServer.password = "jumpfrp-dashboard"

log.to = "/var/log/frps.log"
log.level = "info"
log.maxDays = 3
```

> ⚠️ frp 0.52 以前用 INI 格式（`[common]`），0.52+ 改为 TOML，两者不兼容

### 11.3 Agent 发布流程

每次更新 Agent 代码后需重新编译并上传到 GitHub Releases：

```bash
cd agent
export PATH="/opt/homebrew/bin:$PATH"
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o jumpfrp-agent-linux-amd64 ./cmd/main.go
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o jumpfrp-agent-linux-arm64 ./cmd/main.go
```

上传到：https://github.com/AKEXZ/JumpFrp/releases/tag/v1.0.0

---

## 十二、开发注意事项

### 12.1 本地开发

```bash
bash dev.sh
# 前端: http://localhost:5173
# 后端: http://localhost:8080
```

### 12.2 SQLite 驱动

| 驱动 | CGO | Fly.io | 说明 |
|------|-----|--------|------|
| `gorm.io/driver/sqlite` | 需要 | ❌ | 底层 go-sqlite3，需要 gcc |
| `github.com/glebarez/sqlite` | 不需要 | ✅ | **当前使用**，纯 Go |

---

## 十三、已确认配置

| 项目 | 决定 |
|------|------|
| 平台名称 | JumpFrp |
| 域名 | jumpfrp.top |
| 主控部署 | Fly.io（sin 区域） |
| frps 版本 | 0.61.0（TOML 配置） |
| VIP 付费 | 管理员手动开通（预留支付接口） |
| 邮件通知 | SMTP（后台可视化配置） |
| GitHub 代理 | gitproxy.ake.cx |

---

## 十四、开发进度

- [x] Phase 1 — 基础框架
- [x] Phase 2 — 节点管理（Agent、一键安装、心跳监控）
- [x] Phase 3 — 隧道核心（端口分配、VIP 权限、frpc 配置生成）
- [x] Phase 4 — VIP 系统
- [x] Phase 5 — 完善优化（邮件、流量统计、安全加固）
- [x] 补丁 — 节点编辑/卸载命令、手动添加用户、SMTP 后台配置
- [x] 部署 — Fly.io 上线，修复所有部署问题
- [ ] **待完成** — 编译 Agent 并上传 GitHub Releases v1.0.0
