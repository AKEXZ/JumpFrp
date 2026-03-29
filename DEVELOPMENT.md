# JumpFrp 项目完整开发文档

> 版本：v2.1
> 日期：2026-03-29
> 状态：✅ 功能完成 + 隧道编辑 + Token 自动生成

---

## 一、项目概述

### 1.1 项目简介

**JumpFrp** 是一款基于 [frp (fast reverse proxy)](https://github.com/fatedier/frp) 内网穿透服务二次开发的多节点内网穿透托管平台。

用户可以通过平台注册账号、选择节点服务器、创建隧道配置，使用官方 frpc 客户端连接即可将本地服务暴露到公网。平台支持 VIP 会员制度，管理员可通过后台管理服务器节点、用户和权限。

### 1.2 项目特色

- 🌐 **多节点管理** - 支持添加多台节点服务器，一键安装/卸载
- 👤 **用户系统** - 邮箱注册、JWT 认证、VIP 分级
- 🎫 **VIP 制度** - Free / Basic / Pro / Ultimate 四档，差异化权限
- 🔌 **隧道管理** - TCP/UDP/HTTP/HTTPS 全协议支持
- 🔗 **域名管理** - Pro+ 用户可绑定自定义子域名
- 📊 **实时监控** - 节点 CPU/内存/连接数 30s 自动上报
- 📧 **邮件通知** - 注册验证码、VIP 到期提醒
- 🚀 **免费部署** - 主控支持 Fly.io 免费托管

### 1.3 技术栈

| 层级 | 技术选型 | 版本 | 说明 |
|------|----------|------|------|
| **后端** | Go + Gin | 1.24 | HTTP 框架，轻量高性能 |
| **数据库** | SQLite | - | 使用 `glebarez/sqlite` 纯 Go 驱动 |
| **前端** | Vue 3 + Vite | 3.x / 5.x | 组合式 API + TypeScript |
| **UI 框架** | Element Plus | 2.x | Vue 3 组件库 |
| **状态管理** | Pinia | 2.x | Vue 3 官方推荐 |
| **路由** | Vue Router | 4.x | SPA 路由管理 |
| **HTTP 客户端** | Axios | 1.x | API 请求 |
| **frp 版本** | 0.61.0 | - | 使用 TOML 配置格式 |

### 1.4 项目结构

```
JumpFrp/
├── master/                    # 主控服务 (Go + Gin)
│   ├── cmd/server/main.go     # 程序入口
│   ├── internal/
│   │   ├── api/               # API 处理层
│   │   │   ├── admin/         # 管理员 API
│   │   │   │   ├── handlers.go       # 管理员处理器
│   │   │   │   ├── routes.go         # 路由注册
│   │   │   │   ├── settings.go       # 系统设置
│   │   │   │   ├── tunnel.go         # 隧道管理
│   │   │   │   └── subdomain.go      # 域名管理
│   │   │   ├── agent/         # 节点 Agent API
│   │   │   │   ├── heartbeat.go     # 心跳处理
│   │   │   │   └── register.go      # Agent 注册
│   │   │   ├── auth/          # 认证 API
│   │   │   │   ├── register.go      # 用户注册
│   │   │   │   ├── login.go         # 用户登录
│   │   │   │   └── password.go      # 密码管理
│   │   │   └── user/          # 用户 API
│   │   │       ├── routes.go        # 路由注册
│   │   │       ├── tunnel.go        # 隧道管理
│   │   │       ├── vip.go          # VIP 相关
│   │   │       └── profile.go      # 用户信息
│   │   ├── middleware/        # 中间件
│   │   │   ├── auth.go              # JWT 认证
│   │   │   ├── ratelimit.go         # 限流
│   │   │   └── cors.go              # 跨域
│   │   ├── model/             # 数据模型
│   │   │   └── models.go           # GORM 模型定义
│   │   ├── service/          # 业务逻辑层
│   │   │   ├── auth.go             # 认证服务
│   │   │   ├── mail.go             # 邮件服务
│   │   │   └── tunnel.go           # 隧道服务
│   │   └── config/           # 配置管理
│   │       └── config.go           # 配置加载
│   └── go.mod/go.sum         # Go 依赖
│
├── agent/                     # 节点 Agent
│   ├── cmd/main.go           # 程序入口
│   ├── internal/
│   │   ├── agent.go          # Agent 核心逻辑
│   │   ├── monitor.go        # 系统监控
│   │   ├── installer.go      # frps 安装
│   │   └── config.go         # 配置管理
│   └── scripts/
│       └── install.sh        # 一键安装脚本（生成）
│
├── frontend/                 # 前端 (Vue 3)
│   ├── public/               # 静态资源
│   ├── src/
│   │   ├── api/              # API 客户端
│   │   │   └── index.ts           # API 定义
│   │   ├── assets/           # 资源文件
│   │   ├── components/       # 公共组件
│   │   ├── router/          # 路由配置
│   │   │   └── index.ts           # 路由定义
│   │   ├── stores/          # Pinia 状态
│   │   │   └── auth.ts             # 认证状态
│   │   ├── views/           # 页面组件
│   │   │   ├── user/              # 用户页面
│   │   │   │   ├── HomeView.vue        # 首页
│   │   │   │   ├── LoginView.vue       # 登录
│   │   │   │   ├── RegisterView.vue     # 注册
│   │   │   │   ├── UserLayout.vue      # 用户布局
│   │   │   │   ├── DashboardView.vue   # 控制台
│   │   │   │   ├── TunnelsView.vue    # 隧道管理
│   │   │   │   ├── SubdomainsView.vue # 域名管理
│   │   │   │   └── VIPView.vue        # VIP 中心
│   │   │   └── admin/            # 管理员页面
│   │   │       ├── AdminLayout.vue     # 管理布局
│   │   │       ├── DashboardView.vue   # 仪表盘
│   │   │       ├── UsersView.vue       # 用户管理
│   │   │       ├── NodesView.vue       # 节点管理
│   │   │       ├── TunnelsView.vue     # 隧道管理
│   │   │       ├── SubdomainsView.vue  # 域名管理
│   │   │       ├── OrdersView.vue      # VIP 订单
│   │   │       └── SettingsView.vue    # 系统设置
│   │   ├── App.vue           # 根组件
│   │   └── main.ts           # 入口文件
│   ├── index.html
│   ├── vite.config.ts
│   ├── tsconfig.json
│   └── package.json
│
├── scripts/                  # 脚本
│   ├── install.sh            # Agent 安装脚本
│   └── uninstall.sh          # Agent 卸载脚本
│
├── Dockerfile                # 多阶段构建
├── fly.toml                  # Fly.io 配置
├── docker-compose.yml        # 本地开发
├── .env                      # 环境变量
├── .env.production           # 生产环境变量
├── PROJECT.md                # 项目规划文档
├── README.md                 # 项目说明文档
└── DEVELOP.md                # 本文档
```

---

## 二、数据库设计

### 2.1 数据库概览

- **数据库类型**：SQLite
- **驱动**：`github.com/glebarez/sqlite`（纯 Go，无 CGO 依赖）
- **ORM**：GORM v2
- **数据库文件**：`/data/jumpfrp.db`（Fly.io Volume）

### 2.2 数据模型

#### 2.2.1 用户表 (users)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| username | string(50) | 用户名，唯一索引 |
| email | string(100) | 邮箱，唯一索引 |
| password | string(255) | bcrypt 加密密码 |
| vip_level | int | VIP 等级：0=Free, 1=Basic, 2=Pro, 3=Ultimate |
| vip_expire_at | *time.Time | VIP 到期时间 |
| status | string(20) | 状态：active/pending/banned |
| last_login_at | *time.Time | 最后登录时间 |
| last_login_ip | string(50) | 最后登录 IP |
| created_at | time.Time | 创建时间 |
| updated_at | time.Time | 更新时间 |
| deleted_at | gorm.DeletedAt | 软删除 |

#### 2.2.2 节点表 (nodes)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| name | string(100) | 节点名称 |
| slug | string(50) | 节点标识，唯一索引，自动生成 |
| ip | string(50) | 公网 IP |
| region | string(100) | 地区 |
| frps_port | int | frps 端口，默认 7000 |
| agent_port | int | Agent 端口，默认 7500 |
| agent_token | string(64) | Agent 认证 Token |
| port_range_start | int | 端口池起始，默认 10000 |
| port_range_end | int | 端口池结束，默认 20000 |
| port_excludes | string(255) | 排除端口，逗号分隔 |
| min_vip_level | int | 最低 VIP 等级，默认 0 |
| bandwidth_limit | int | 带宽限制(Mbps)，默认 0 |
| max_connections | int | 最大连接数，默认 100 |
| status | string(20) | 状态：online/offline/maintain |
| version | string(20) | Agent 版本 |
| last_heartbeat | *time.Time | 最后心跳时间 |
| cpu_usage | float64 | CPU 使用率 |
| memory_usage | float64 | 内存使用率 |
| current_conns | int | 当前连接数 |
| installed | bool | 是否已安装 |
| remark | string(255) | 备注 |

#### 2.2.3 隧道表 (tunnels)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| user_id | uint | 用户 ID，外键 |
| node_id | uint | 节点 ID，外键 |
| name | string(100) | 隧道名称 |
| protocol | string(10) | 协议：tcp/udp/http/https |
| local_ip | string(50) | 本地 IP |
| local_port | int | 本地端口 |
| remote_port | int | 分配的远程端口 |
| subdomain | string(50) | 子域名（HTTP/HTTPS） |
| status | string(20) | 状态：active/stopped |
| bandwidth_limit | int | 带宽限制（Mbps） |
| enabled | bool | 是否启用（true/false） |
| created_at | time.Time | 创建时间 |
| updated_at | time.Time | 更新时间 |

#### 2.2.4 域名表 (subdomains)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| user_id | uint | 用户 ID，外键 |
| tunnel_id | uint | 关联隧道 ID |
| subdomain | string(50) | 子域名（不含后缀），唯一索引 |
| status | string(20) | 状态：pending/approved/rejected |
| created_at | time.Time | 创建时间 |
| updated_at | time.Time | 更新时间 |

#### 2.2.5 VIP 订单表 (vip_orders)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| user_id | uint | 用户 ID，外键 |
| plan_id | int | 套餐 ID |
| days | int | 天数 |
| amount | float64 | 金额 |
| status | string(20) | 状态：pending/completed/cancelled |
| created_at | time.Time | 创建时间 |
| updated_at | time.Time | 更新时间 |

#### 2.2.6 系统配置表 (system_configs)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| key | string(100) | 配置键，唯一索引 |
| value | text | 配置值 |
| created_at | time.Time | 创建时间 |
| updated_at | time.Time | 更新时间 |

**内置配置键**：
- `smtp_host` - SMTP 服务器
- `smtp_port` - SMTP 端口
- `smtp_user` - SMTP 用户
- `smtp_pass` - SMTP 密码
- `smtp_from` - 发件人邮箱
- `site_name` - 站点名称
- `site_domain` - 站点域名
- `allow_register` - 是否开放注册
- `verify_code:{email}` - 邮箱验证码 |

### 2.3 VIP 等级说明

| 等级 | 值 | 隧道数 | 端口数 | 带宽 | 协议 | 子域名 |
|------|-----|--------|--------|------|------|--------|
| Free | 0 | 1 | 3 | 1 Mbps | TCP | ✗ |
| Basic | 1 | 5 | 10 | 5 Mbps | TCP/UDP | ✗ |
| Pro | 2 | 20 | 50 | 20 Mbps | 全协议 | ✓ |
| Ultimate | 3 | ∞ | 200 | 100 Mbps | 全协议 | ✓ |

---

## 三、API 接口文档

### 3.1 用户端 API

#### 认证相关

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/user/auth/send-code | 发送验证码 |
| POST | /api/user/auth/register | 用户注册 |
| POST | /api/user/auth/login | 用户登录 |
| POST | /api/user/auth/reset-password | 重置密码 |

#### 用户信息

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/user/profile | 获取个人信息 |
| PUT | /api/user/password | 修改密码 |

#### 节点相关

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/user/nodes | 可用节点列表（公开） |
| GET | /api/user/available-nodes | VIP 可用节点（需登录） |

#### 隧道相关

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/user/tunnels | 隧道列表 |
| POST | /api/user/tunnels | 创建隧道 |
| PUT | /api/user/tunnels/:id | 编辑隧道 |
| DELETE | /api/user/tunnels/:id | 删除隧道 |
| PUT | /api/user/tunnels/:id/toggle | 开启/关闭隧道 |
| GET | /api/user/tunnels/:id/frpc-config | 下载 frpc 配置 |

#### 域名相关

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/user/subdomains | 我的域名列表 |
| POST | /api/user/subdomains | 绑定域名（Pro+） |
| DELETE | /api/user/subdomains/:id | 解绑域名 |

#### VIP 相关

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/user/vip/plans | 套餐列表 |
| GET | /api/user/vip/info | VIP 信息 |
| GET | /api/user/vip/orders | 我的订单 |

### 3.2 管理员 API

#### 仪表盘

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/admin/dashboard | 统计数据 |

#### 用户管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/admin/users | 用户列表 |
| POST | /api/admin/users | 手动添加用户 |
| PUT | /api/admin/users/:id/vip | 设置 VIP |
| PUT | /api/admin/users/:id/ban | 封禁/解封用户 |
| PUT | /api/admin/users/:id/password | 重置密码 |
| DELETE | /api/admin/users/:id | 删除用户 |

#### 节点管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/admin/nodes | 节点列表 |
| POST | /api/admin/nodes | 创建节点 |
| PUT | /api/admin/nodes/:id | 更新节点 |
| DELETE | /api/admin/nodes/:id | 删除节点 |
| GET | /api/admin/nodes/:id/install-cmd | 获取安装命令 |

#### 隧道管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/admin/tunnels | 隧道列表 |
| DELETE | /api/admin/tunnels/:id | 删除隧道 |

#### 域名管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/admin/subdomains | 域名列表 |
| POST | /api/admin/subdomains | 手动添加域名 |
| PUT | /api/admin/subdomains/:id/approve | 审批域名 |
| DELETE | /api/admin/subdomains/:id | 删除域名 |

#### VIP 订单

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/admin/orders | 订单列表 |
| POST | /api/admin/vip/grant | 手动开通 VIP |

#### 系统设置

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/admin/settings | 获取设置 |
| POST | /api/admin/settings/smtp | 保存 SMTP 配置 |
| POST | /api/admin/settings/smtp/test | 测试邮件 |
| POST | /api/admin/settings/site | 保存站点配置 |

### 3.3 Agent API

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/agent/register | Agent 注册 |
| POST | /api/agent/heartbeat | 心跳上报 |

---

## 四、前端页面路由

### 4.1 用户端路由

| 路径 | 页面 | 说明 |
|------|------|------|
| / | HomeView | 首页 |
| /login | LoginView | 登录 |
| /register | RegisterView | 注册 |
| /dashboard | DashboardView | 控制台 |
| /tunnels | TunnelsView | 隧道管理 |
| /subdomains | SubdomainsView | 域名管理 |
| /vip | VIPView | VIP 中心 |

### 4.2 管理员路由

| 路径 | 页面 | 说明 |
|------|------|------|
| /admin | - | 管理后台入口 |
| /admin/dashboard | DashboardView | 仪表盘 |
| /admin/users | UsersView | 用户管理 |
| /admin/nodes | NodesView | 节点管理 |
| /admin/tunnels | TunnelsView | 隧道管理 |
| /admin/subdomains | SubdomainsView | 域名管理 |
| /admin/orders | OrdersView | VIP 订单 |
| /admin/settings | SettingsView | 系统设置 |

---

## 五、部署文档

### 5.1 Fly.io 部署

#### 1. 创建应用
```bash
fly auth login
fly apps create jumpfrp
```

#### 2. 创建存储卷
```bash
fly volumes create jumpfrp_data --size 1 --region sin --app jumpfrp
```

#### 3. 设置环境变量
```bash
fly secrets set JWT_SECRET="$(openssl rand -base64 48)" --app jumpfrp
fly secrets set GIN_MODE="release" --app jumpfrp
```

#### 4. 部署
```bash
fly deploy --app jumpfrp
```

### 5.2 节点安装

在管理后台 → 节点管理 → 添加节点后，点击「安装」复制命令：

```bash
bash <(wget -qO- https://api.jumpfrp.top/install.sh) \
  --node-id <slug> \
  --token <token> \
  --master-url https://api.jumpfrp.top \
  --frps-port 7000 \
  --agent-port 7500
```

### 5.3 环境变量

| 变量 | 说明 | 示例 |
|------|------|------|
| JWT_SECRET | JWT 签名密钥 | base64 随机字符串 |
| GIN_MODE | 运行模式 | release |
| DATABASE_URL | 数据库连接（可选） | file:/data/jumpfrp.db |

---

## 六、开发指南

### 6.1 本地开发

```bash
# 克隆项目
git clone https://github.com/AKEXZ/JumpFrp.git
cd JumpFrp

# 启动后端
cd master
go mod tidy
go run ./cmd/server/main.go

# 启动前端（新终端）
cd frontend
npm install
npm run dev
```

### 6.2 API 文档（生产环境）

- 后端地址：`https://api.jumpfrp.top`
- 前端地址：`https://jumpfrp.top`
- 管理后台：`https://jumpfrp.top/admin`

### 6.3 默认账户

- 管理员：`admin` / `admin123456`
- 邮箱：`admin@jumpfrp.top`

⚠️ **生产环境必须修改默认密码！**

### 6.4 frpc 配置说明

每个隧道的 frpc 配置文件由后端自动生成，包含：

```toml
[common]
server_addr = "节点IP"
server_port = 7000
auth.method = "token"
auth.token = "用户APIToken"  # 自动生成，用于认证
pool_count = 10
transport.tcp_mux = true
transport.protocol = "tcp"

[[proxies]]
name = "隧道名称"
type = "tcp"
local_ip = "127.0.0.1"
local_port = 8080
remote_port = 12345
bandwidth_limit = "5MB"
```

**重要说明：**
- ✅ 每个隧道的配置文件只包含该隧道的 `[[proxies]]` 段
- ❌ 用户不应手动修改配置文件或添加多个 `[[proxies]]` 段
- ✅ 如需多个隧道，应在平台上创建多个隧道，分别下载配置
- ✅ 修改隧道配置后，需重新下载 frpc 配置文件并重启 frpc 客户端

### 6.5 APIToken 自动生成

- 新用户注册时自动生成 32 位随机 token
- 旧用户首次启动时自动生成 token（如果为空）
- Token 用于 frpc 客户端认证，不可修改或重置

---

## 七、常见问题

### Q1: frpc 连接失败 - "token in login doesn't match"

**原因**：用户的 APIToken 为空或不正确

**解决**：
1. 重启主控服务（会自动为所有用户生成 token）
2. 重新下载 frpc 配置文件
3. 重启 frpc 客户端

### Q2: 如何修改隧道配置？

**步骤**：
1. 进入"我的隧道"页面
2. 点击隧道卡片的"编辑"按钮
3. 修改节点、协议、本地 IP、本地端口
4. 点击"保存"
5. 重新下载 frpc 配置文件
6. 重启 frpc 客户端

### Q3: VIP 过期后隧道会怎样？

**行为**：
- 所有隧道自动关闭（enabled = false）
- 带宽限制降为 1 Mbps
- 用户无法创建新隧道（超过 Free 限制）
- 续费后可手动开启隧道

### Q4: Free 用户最多能开启几条隧道？

**限制**：
- Free 用户最多 1 条启用的隧道
- 可创建多条隧道，但同时只能启用 1 条
- 需要启用其他隧道时，先关闭当前隧道

---

## 八、更新日志

### v2.1 (2026-03-29)

**新增功能**：
- ✅ 隧道编辑功能（修改节点、协议、本地 IP、本地端口）
- ✅ 隧道开关功能（启用/禁用隧道）
- ✅ APIToken 自动生成（新用户注册时生成，旧用户启动时补齐）
- ✅ VIP 过期自动处理（关闭隧道、降低带宽）

**修复**：
- 🔧 修复 frpc 配置中 token 不匹配的问题
- 🔧 修复用户节点列表查询问题
- 🔧 修复前端 TypeScript 类型错误

### v2.0 (2026-03-27)

**完整功能**：
- 用户系统（注册、登录、VIP 分级）
- 隧道管理（创建、删除、下载配置）
- 域名管理（申请、审批、绑定）
- 节点管理（添加、编辑、删除）
- 带宽限制（按用户 VIP 等级限制）
- 邮件通知（验证码、VIP 到期提醒）

---

## 七、配置文件说明

### 7.1 后端配置 (config/config.go)

```go
type Config struct {
    JWT_SECRET string // JWT 签名密钥
    Port      string // 监听端口，默认 8080
    Database  string // 数据库路径
}
```

### 7.2 前端环境变量

**开发环境 (.env)**
```env
VITE_API_URL=http://localhost:8080/api
```

**生产环境 (.env.production)**
```env
VITE_API_URL=/api
```

### 7.3 Dockerfile 多阶段构建

1. **阶段1：构建前端** - Node 20 Alpine
2. **阶段2：构建后端** - Go 1.24 Alpine
3. **阶段3：运行** - Alpine Latest

---

## 八、关键业务逻辑

### 8.1 用户注册流程

1. 用户提交邮箱 → 发送验证码到邮箱
2. 用户填写用户名、密码、验证码 → 注册
3. 验证码存储在 `system_configs` 表，5 分钟有效
4. 注册成功后自动登录

### 8.2 隧道创建流程

1. 用户选择节点、协议、填写本地信息 → 创建隧道
2. 系统检查用户 VIP 等级和隧道数量限制
3. 自动分配远程端口（根据节点端口池）
4. 生成 frpc.toml 配置文件供用户下载

### 8.3 Agent 心跳机制

1. Agent 每 30 秒上报心跳到主控
2. 主控更新节点最后心跳时间
3. 超过 2 分钟无心跳，节点标记为 offline
4. Agent 会上报 CPU、内存、连接数等信息

### 8.4 域名绑定规则

1. 只有 Pro+ 用户（VIPLevel >= 2）可以绑定域名
2. 只支持 HTTP/HTTPS 协议的隧道
3. 子域名全局唯一
4. 用户申请的域名自动批准

---

## 九、已知问题与解决方案

| 问题 | 原因 | 解决 |
|------|------|------|
| SQLite CGO 报错 | 默认驱动需要 CGO | 使用 `glebarez/sqlite` 纯 Go 驱动 |
| 前端 404 | 镜像无前端文件 | Dockerfile 多阶段构建 |
| API 请求 localhost | 生产用了开发地址 | `.env.production` 用相对路径 `/api` |
| hkg 区域不可用 | Fly.io 已废弃 | 使用 `sin`（新加坡）区域 |
| go.mod 版本冲突 | 本地 Go 版本高于镜像 | Dockerfile 用 `golang:1.24-alpine` |
| fly.toml 格式错误 | `[[http_service]]` 写法错误 | 改为 `[http_service]` |
| frps 启动失败 | frp 0.61.0 改用 TOML 格式 | 生成 `frps.toml`，不再用 INI |
| Agent Exec format error | 下载到 HTML 而非二进制 | 从 GitHub Releases 下载 |
| GitHub 下载卡住 | 服务器无法访问 GitHub | 优先用 `gitproxy.ake.cx` 代理 |
| 节点查询返回空 | Select 字段名不匹配 | 移除 Select 子句，直接查询全部字段 |

---

## 十、GitHub Releases 发布

每次更新 Agent 代码后需重新编译并上传到 GitHub Releases：

```bash
cd agent

# 编译 amd64
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o jumpfrp-agent-linux-amd64 ./cmd/main.go

# 编译 arm64
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o jumpfrp-agent-linux-arm64 ./cmd/main.go

# 压缩
gzip -k jumpfrp-agent-linux-amd64
gzip -k jumpfrp-agent-linux-arm64

# 上传到 GitHub Releases
# https://github.com/AKEXZ/JumpFrp/releases
```

---

## 十一、许可证

MIT License © 2026 JumpFrp

---

## 附录 A：VIP 套餐配置

| 套餐 | ID | 价格 | 天数 | 隧道数 | 端口数 | 带宽 | 协议 | 子域名 |
|------|-----|------|------|--------|--------|------|------|--------|
| Basic | 1 | ¥9.9 | 30 | 5 | 10 | 5 Mbps | TCP/UDP | ✗ |
| Pro | 2 | ¥29.9 | 30 | 20 | 50 | 20 Mbps | 全协议 | ✓ |
| Ultimate | 3 | ¥99 | 30 | ∞ | 200 | 100 Mbps | 全协议 | ✓ |

## 附录 B：SMTP 配置模板

| 服务商 | SMTP 地址 | 端口 | SSL |
|--------|-----------|------|-----|
| QQ 邮箱 | smtp.qq.com | 465 | ✓ |
| 163 邮箱 | smtp.163.com | 465 | ✓ |
| Gmail | smtp.gmail.com | 587 | ✗ |
| Outlook | smtp.office365.com | 587 | ✗ |

## 附录 C：frp 0.61.0 配置格式

frps.toml 示例：
```toml
bindPort = 7000
auth.token = "AGENT_TOKEN"
transport.maxPoolCount = 100
transport.poolCount = 10

[[vhost.httpRoutes]]
customDomains = ["*.jumpfrp.top"]
handlerRegistries = []

[[vhost.httpsRoutes]]
customDomains = ["*.jumpfrp.top"]
```

---

*本文档最后更新于 2026-03-27*
