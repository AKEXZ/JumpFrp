# JumpFrp — 内网穿透托管平台 项目完整文档

> 版本：v2.0
> 日期：2026-03-28
> 状态：✅ 功能完成

---

## 一、项目概述

### 1.1 项目简介

**JumpFrp** 是一款基于 [frp (fast reverse proxy)](https://github.com/fatedier/frp) 内网穿透服务二次开发的多节点内网穿透托管平台。

用户可以通过平台注册账号、选择节点服务器、创建隧道配置，使用官方 frpc 客户端连接即可将本地服务暴露到公网。平台支持 VIP 会员制度，管理员可通过后台管理服务器节点、用户和权限。

### 1.2 项目特色

- 🌐 **多节点管理** - 支持添加多台节点服务器，一键安装/卸载，实时监控
- 👤 **用户系统** - 邮箱注册（验证码）、用户名/邮箱登录、JWT 认证
- 🎫 **VIP 制度** - Free / Basic / Pro / Ultimate 四档，差异化权限
- 🔌 **隧道管理** - TCP/UDP/HTTP/HTTPS 全协议，自动分配端口
- 🔗 **域名管理** - Pro+ 用户可绑定自定义子域名，管理员统一管理
- 📊 **实时监控** - 节点 CPU/内存/连接数 30s 自动上报
- 🚀 **带宽限制** - **服务端强制限速**（Linux TC + iptables）
- 📧 **邮件通知** - 注册验证码、VIP 到期提醒，SMTP 后台可视化配置
- 🔐 **安全加固** - 全局限流、登录防暴力破解、JWT 认证、安全响应头
- ⚙️ **系统设置** - SMTP / 站点配置全部后台管理
- 🚀 **免费部署** - 主控支持 Fly.io 免费托管，节点一键安装

### 1.3 技术栈

| 层级 | 技术选型 | 版本 | 说明 |
|------|----------|------|------|
| **后端** | Go + Gin | 1.24 | HTTP 框架，高性能 |
| **数据库** | SQLite | - | `github.com/glebarez/sqlite` 纯 Go，无 CGO |
| **前端** | Vue 3 + Vite | 3.x / 5.x | 组合式 API + TypeScript |
| **UI 框架** | Element Plus | 2.x | Vue 3 组件库 |
| **状态管理** | Pinia | 2.x | Vue 3 官方推荐 |
| **路由** | Vue Router | 4.x | SPA 路由 |
| **HTTP 客户端** | Axios | 1.x | API 请求 |
| **frp 版本** | 0.61.0 | - | 使用 TOML 配置格式 |

### 1.4 项目地址

| 类型 | 地址 |
|------|------|
| GitHub 仓库 | https://github.com/AKEXZ/JumpFrp |
| 主控前台 | https://jumpfrp.top |
| API 地址 | https://api.jumpfrp.top |
| 管理后台 | https://jumpfrp.top/admin |
| Fly.io App | jumpfrp (sin 区域) |

---

## 二、系统架构

### 2.1 架构图

```
┌─────────────────────────────────────────────────────────────────────┐
│                           用户浏览器                                  │
│                  ┌──────────────────────┐                          │
│                  │   前台 (Vue 3 SPA)    │                          │
│                  │   /dashboard         │                          │
│                  │   /tunnels          │                          │
│                  │   /subdomains       │                          │
│                  │   /vip              │                          │
│                  └──────────────────────┘                          │
│                  ┌──────────────────────┐                          │
│                  │   管理后台 (Vue 3)    │                          │
│                  │   /admin/dashboard   │                          │
│                  │   /admin/users      │                          │
│                  │   /admin/nodes      │                          │
│                  │   /admin/subdomains  │                          │
│                  └──────────────────────┘                          │
└────────────────────────────┬───────────────────────────────────────┘
                             │ HTTPS (443)
                             ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        主控服务 (Master)                              │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                      Go + Gin 后端                            │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │   │
│  │  │ 用户认证    │  │ 节点管理    │  │ 隧道管理    │         │   │
│  │  │ JWT/Session │  │ 心跳监控   │  │ 端口分配    │         │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘         │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │   │
│  │  │ VIP 系统   │  │ 域名管理    │  │ SMTP 邮件  │         │   │
│  │  │ 套餐/订单  │  │ 子域名绑定  │  │ 发送通知   │         │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘         │   │
│  └─────────────────────────────────────────────────────────────┘   │
│  ┌─────────────┐  ┌─────────────────────────────────────┐      │
│  │  SQLite    │  │        Vue 3 前端 (静态文件)            │      │
│  │  数据库    │  │   /assets/* (JS/CSS/图片)              │      │
│  └─────────────┘  └─────────────────────────────────────┘      │
│                                                                        │
│  📌 SQLite 持久化: /data/jumpfrp.db                                   │
│  📌 JWT_SECRET: 环境变量配置                                         │
└────────────────────────────┬───────────────────────────────────────┘
                             │ HTTPS (API)
         ┌───────────────────┼───────────────────┐
         │                   │                   │
         ▼                   ▼                   ▼
┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
│   节点 A        │ │   节点 B        │ │   节点 C        │
│  8.130.118.231 │ │                 │ │                 │
│ ┌─────────────┐│ │                 │ │                 │
│ │ jumpfrp-agent│ │                 │ │                 │
│ │  - 心跳上报 ││ │                 │ │                 │
│ │  - frps 管理│ │                 │ │                 │
│ │  - TC 限速 ││ │                 │ │                 │
│ └──────┬──────┘│ │                 │ │                 │
│        │        │ │                 │ │                 │
│ ┌──────▼──────┐│ │                 │ │                 │
│ │   frps      ││ │                 │ │                 │
│ │  0.61.0     ││ │                 │ │                 │
│ │  TOML 配置   ││ │                 │ │                 │
│ │  带宽限制    ││ │                 │ │                 │
│ └──────┬──────┘│ │                 │ │                 │
│        │        │ │                 │ │                 │
│        ▼        │                 │ │
│   ┌────┴────┐   │                 │ │
│   │ frpc 连接│   │                 │ │
│   │ 用户设备 │   │                 │ │
│   └─────────┘   │                 │ │
└─────────────────┘ └─────────────────┘ └─────────────────┘
```

### 2.2 数据流向

```
用户创建隧道流程：
1. 用户在浏览器创建隧道 → 前端 POST /api/user/tunnels
2. 主控验证 VIP 权限 → 分配远程端口 → 写入数据库
3. 用户下载 frpc.toml → 配置本地 frpc
4. frpc 连接节点 frps → 隧道建立

节点心跳流程：
1. Agent 每 30s 上报心跳 → POST /api/agent/heartbeat
2. 主控更新节点状态 → 检查离线超时
3. Agent 获取 frps.toml 配置 → 按需热更新

带宽限制流程：
1. frps 记录连接日志 (auth token)
2. Agent 解析日志 → 获取用户 token
3. Agent 查询主控获取用户 VIP 等级
4. Agent 通过 iptables + tc 设置带宽限制
```

---

## 三、数据库设计

### 3.1 数据库概览

- **数据库类型**：SQLite
- **驱动**：`github.com/glebarez/sqlite`（纯 Go，无 CGO）
- **ORM**：GORM v2
- **数据库文件**：`/data/jumpfrp.db`（Fly.io Volume）
- **迁移方式**：AutoMigrate + 启动时检查缺失列

### 3.2 数据模型

#### 3.2.1 用户表 (users)

```go
type User struct {
    ID           uint           `gorm:"primarykey" json:"id"`
    Username     string         `gorm:"uniqueIndex;size:50" json:"username"`
    Email        string         `gorm:"uniqueIndex;size:100" json:"email"`
    PasswordHash string         `gorm:"size:255" json:"-"`
    VIPLevel     int            `gorm:"column:vip_level;default:0" json:"vip_level"`
    VIPExpireAt  *time.Time    `gorm:"column:vip_expire_at" json:"vip_expire_at"`
    APIToken     string         `gorm:"uniqueIndex;size:64" json:"api_token"`
    Status       string         `gorm:"size:20;default:'active'" json:"status"`
    EmailVerified bool          `gorm:"column:email_verified;default:false" json:"email_verified"`
    VerifyCode   string         `gorm:"column:verify_code;size:10" json:"-"`
    VerifyExpire *time.Time    `gorm:"column:verify_expire" json:"-"`
    ResetToken   string         `gorm:"column:reset_token;size:64" json:"-"`
    ResetExpire  *time.Time    `gorm:"column:reset_expire" json:"-"`
    CreatedAt    time.Time      `json:"created_at"`
    UpdatedAt    time.Time      `json:"updated_at"`
    DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
```

#### 3.2.2 节点表 (nodes)

```go
type Node struct {
    ID              uint           `gorm:"primarykey" json:"id"`
    Name            string         `gorm:"size:100" json:"name"`
    Slug            string         `gorm:"uniqueIndex;size:50" json:"slug"`
    IP              string         `gorm:"size:50" json:"ip"`
    Region          string         `gorm:"size:100" json:"region"`
    FrpsPort        int            `gorm:"default:7000" json:"frps_port"`
    AgentPort       int            `gorm:"default:7500" json:"agent_port"`
    AgentToken      string         `gorm:"size:64" json:"-"`
    PortRangeStart   int            `json:"port_range_start"`
    PortRangeEnd     int            `json:"port_range_end"`
    PortExcludes     string         `gorm:"size:500" json:"port_excludes"`
    MinVIPLevel      int            `gorm:"column:min_vip_level;default:0" json:"min_vip_level"`
    BandwidthLimit   int            `gorm:"column:bandwidth_limit;default:0" json:"bandwidth_limit"`
    MaxConnections   int            `json:"max_connections"`
    Status          string         `gorm:"size:20;default:'offline'" json:"status"`
    Version         string         `gorm:"size:20" json:"version"`
    LastHeartbeat    *time.Time     `json:"last_heartbeat"`
    CPUUsage         float64        `json:"cpu_usage"`
    MemoryUsage      float64        `json:"memory_usage"`
    CurrentConns     int            `json:"current_conns"`
    Installed        bool           `json:"installed"`
    Remark          string         `gorm:"size:500" json:"remark"`
    CreatedAt        time.Time       `json:"created_at"`
    UpdatedAt        time.Time       `json:"updated_at"`
    DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}
```

#### 3.2.3 隧道表 (tunnels)

```go
type Tunnel struct {
    ID            uint           `gorm:"primarykey" json:"id"`
    UserID        uint           `gorm:"index" json:"user_id"`
    NodeID        uint           `gorm:"index" json:"node_id"`
    Name          string         `gorm:"size:100" json:"name"`
    Protocol      string         `gorm:"size:10" json:"protocol"`
    LocalIP       string         `gorm:"size:50" json:"local_ip"`
    LocalPort     int            `json:"local_port"`
    RemotePort    int            `json:"remote_port"`
    Subdomain     string         `gorm:"size:100" json:"subdomain"`
    BandwidthLimit int           `gorm:"column:bandwidth_limit" json:"bandwidth_limit"`
    Status        string         `gorm:"size:20;default:'inactive'" json:"status"`
    CreatedAt     time.Time       `json:"created_at"`
    UpdatedAt     time.Time       `json:"updated_at"`
    DeletedAt     gorm.DeletedAt  `gorm:"index" json:"-"`
    User          User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
    Node          Node           `gorm:"foreignKey:NodeID" json:"node,omitempty"`
}
```

#### 3.2.4 域名表 (subdomains)

```go
type Subdomain struct {
    ID         uint           `gorm:"primarykey" json:"id"`
    UserID     uint           `gorm:"index" json:"user_id"`
    TunnelID   uint           `gorm:"index" json:"tunnel_id"`
    Subdomain  string         `gorm:"uniqueIndex;size:50" json:"subdomain"`
    Status     string         `gorm:"size:20;default:'pending'" json:"status"`
    CreatedAt  time.Time      `json:"created_at"`
    UpdatedAt  time.Time      `json:"updated_at"`
    DeletedAt  gorm.DeletedAt  `gorm:"index" json:"-"`
    User      User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
```

#### 3.2.5 VIP 订单表 (vip_orders)

```go
type VIPOrder struct {
    ID          uint           `gorm:"primarykey" json:"id"`
    UserID      uint           `gorm:"index" json:"user_id"`
    VIPLevel    int            `gorm:"column:vip_level" json:"vip_level"`
    DurationDays int           `json:"duration_days"`
    Price       float64        `json:"price"`
    Status      string         `gorm:"size:20" json:"status"`
    ExpireAt    *time.Time     `json:"expire_at"`
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
    DeletedAt   gorm.DeletedAt  `gorm:"index" json:"-"`
}
```

#### 3.2.6 系统配置表 (system_configs)

```go
type SystemConfig struct {
    ID        uint      `gorm:"primarykey" json:"id"`
    Key       string    `gorm:"uniqueIndex;size:100" json:"key"`
    Value     string    `gorm:"type:text" json:"value"`
    Remark    string    `gorm:"size:255" json:"remark"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

### 3.3 数据库兼容性处理

启动时自动检查并添加缺失的列：

```go
func fixMissingColumns(db *gorm.DB) {
    // Users 表
    userColumns := []struct {
        column  string
        colType string
    }{
        {"vip_level", "INTEGER DEFAULT 0"},
        {"vip_expire_at", "TEXT"},
        {"email_verified", "INTEGER DEFAULT 0"},
        {"api_token", "TEXT"},
        {"bandwidth_limit", "INTEGER DEFAULT 0"},
    }
    for _, col := range userColumns {
        var count int64
        db.Raw("SELECT COUNT(*) FROM pragma_table_info('users') WHERE name = ?", col.column).Scan(&count)
        if count == 0 {
            db.Exec("ALTER TABLE users ADD COLUMN " + col.column + " " + col.colType)
        }
    }
}
```

---

## 四、VIP 等级设计

### 4.1 等级定义

| 等级 | 值 | 隧道数 | 端口数 | 带宽 | 协议 | 子域名 | 节点限制 |
|------|-----|--------|--------|------|------|--------|---------|
| **Free** | 0 | 1 | 3 | 1 Mbps | TCP | ✗ | 免费节点 |
| **Basic** | 1 | 5 | 10 | 5 Mbps | TCP/UDP | ✗ | 免费节点 |
| **Pro** | 2 | 20 | 50 | 20 Mbps | 全协议 | ✓ | 付费节点 |
| **Ultimate** | 3 | ∞ | 200 | 100 Mbps | 全协议 | ✓ | 付费节点 |

### 4.2 VIP 权益代码

```go
type VIPQuota struct {
    MaxTunnels    int
    MaxPorts      int
    MaxBandwidth  int  // Mbps
    Protocols     []string
    CanSubdomain  bool
}

var VIPQuotas = map[int]VIPQuota{
    0: {1, 3, 1, []string{"tcp"}, false},
    1: {5, 10, 5, []string{"tcp", "udp"}, false},
    2: {20, 50, 20, []string{"tcp", "udp", "http", "https"}, true},
    3: {9999, 200, 100, []string{"tcp", "udp", "http", "https"}, true},
}
```

### 4.3 VIP 升级自动调整

当管理员升级用户 VIP 时，系统自动调整用户所有现有隧道的带宽限制：

```go
func upgradeTunnelBandwidth(db *gorm.DB, userID interface{}, newLevel int) {
    quota := getVIPQuota(newLevel)
    db.Model(&model.Tunnel{}).Where("user_id = ?", userID).Updates(map[string]interface{}{
        "BandwidthLimit": quota.MaxBandwidth,
    })
}
```

---

## 五、带宽限制系统

### 5.1 设计方案

采用 **Linux TC (Traffic Control) + iptables** 实现服务端强制限速，用户无法绕过。

### 5.2 技术原理

```
┌─────────────────────────────────────────────────────────────────────┐
│                      节点服务器 (Linux)                                │
│                                                                      │
│  用户 frpc ──→ frps ──→ 内网服务                                     │
│                    │                                                  │
│                    ▼                                                  │
│             ┌──────────────┐                                         │
│             │ frps 日志    │                                         │
│             │ login from IP│                                         │
│             └──────┬───────┘                                         │
│                    │                                                  │
│                    ▼                                                  │
│             ┌──────────────┐                                         │
│             │ Agent        │  解析日志                               │
│             │ (Go)         │  获取 auth token                        │
│             └──────┬───────┘                                         │
│                    │                                                  │
│                    ▼                                                  │
│             ┌──────────────┐                                         │
│             │ 查询 VIP 等级 │  GET /api/agent/get-user-vip         │
│             │ (HTTP)      │                                         │
│             └──────┬───────┘                                         │
│                    │                                                  │
│                    ▼                                                  │
│  ┌─────────────────────────────────────────────────────────────┐     │
│  │                   iptables + tc                             │     │
│  │                                                               │     │
│  │  iptables -t mangle -A TC_MARK -s 用户IP -j MARK --set-mark │     │
│  │                                                               │     │
│  │  tc class add dev eth0 parent 1: classid 1:21 htb rate 20mbit│     │
│  │                                                               │     │
│  │  tc filter add dev eth0 parent 1: protocol ip fw             │     │
│  │               handle MARK classid 1:21                        │     │
│  └─────────────────────────────────────────────────────────────┘     │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### 5.3 VIP 带宽等级

| VIP 等级 | TC Class ID | 带宽限制 (rate) | 突发 (ceil) |
|----------|------------|----------------|-------------|
| Free | 1:1 | 1 mbit | 2 mbit |
| Basic | 1:11 | 5 mbit | 10 mbit |
| Pro | 1:21 | 20 mbit | 40 mbit |
| Ultimate | 1:31 | 100 mbit | 200 mbit |

### 5.4 Agent TC 模块

```go
// agent/internal/tc/traffic.go
type TrafficControl struct {
    iface       string
    connections map[string]*ConnectionInfo
    vipLimits   map[int]int  // VIP等级 → 带宽(Mbps)
}

func (tc *TrafficControl) Init() error {
    // 1. 清理旧规则
    tc.cleanup()
    // 2. 创建根 HTB 队列
    exec.Command("tc", "qdisc", "add", "dev", tc.iface, "root", "handle", "1:", "htb", "default", "9999")
    // 3. 创建 VIP 分类
    for vip, rate := range tc.vipLimits {
        classID := fmt.Sprintf("1:%d", vip*10+1)
        exec.Command("tc", "class", "add", "dev", tc.iface, "parent", "1:",
            "classid", classID, "htb", "rate", fmt.Sprintf("%dmbit", rate))
    }
    return nil
}

func (tc *TrafficControl) AddConnection(token, ip string, vipLevel int) error {
    mark := allocateMark()
    classID := fmt.Sprintf("1:%d", vipLevel*10+1)
    
    // 添加 iptables MARK 规则
    exec.Command("iptables", "-t", "mangle", "-A", "TC_MARK", "-s", ip,
        "-j", "MARK", "--set-mark", strconv.Itoa(mark))
    
    // 添加 tc filter 规则
    exec.Command("tc", "filter", "add", "dev", tc.iface, "parent", "1:",
        "protocol", "ip", "fw", "handle", strconv.Itoa(mark), "classid", classID)
    
    return nil
}
```

---

## 六、API 接口文档

### 6.1 用户端 API

#### 认证相关

| 方法 | 路径 | 说明 | 参数 |
|------|------|------|------|
| POST | `/api/user/auth/send-code` | 发送验证码 | `{email}` |
| POST | `/api/user/auth/register` | 用户注册 | `{username, email, password, code}` |
| POST | `/api/user/auth/login` | 用户登录 | `{email, password}` |
| POST | `/api/user/auth/reset-password` | 重置密码 | `{email, code, password}` |

#### 用户信息

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/user/profile` | 获取个人信息 |
| PUT | `/api/user/password` | 修改密码 |

#### 节点相关

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/user/nodes` | 公开节点列表（所有节点） |
| GET | `/api/user/available-nodes` | 用户可用节点（按 VIP 过滤） |

#### 隧道相关

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/user/tunnels` | 隧道列表 |
| POST | `/api/user/tunnels` | 创建隧道 |
| DELETE | `/api/user/tunnels/:id` | 删除隧道 |
| GET | `/api/user/tunnels/:id/frpc-config` | 下载 frpc.toml 配置 |

**创建隧道请求：**
```json
{
    "node_id": 1,
    "name": "我的Web服务",
    "protocol": "tcp",
    "local_ip": "127.0.0.1",
    "local_port": 8080
}
```

**frpc.toml 响应：**
```toml
[common]
server_addr = "8.130.118.231"
server_port = 7000
auth.method = "token"
auth.token = "user-api-token"
pool_count = 10
transport.tcp_mux = true
transport.protocol = "tcp"

[[proxies]]
name = "我的Web服务"
type = "tcp"
local_ip = "127.0.0.1"
local_port = 8080
remote_port = 12345
```

#### 域名相关

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/user/subdomains` | 我的域名列表 |
| POST | `/api/user/subdomains` | 绑定域名（Pro+） |
| DELETE | `/api/user/subdomains/:id` | 解绑域名 |

#### VIP 相关

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/user/vip/plans` | 套餐列表 |
| GET | `/api/user/vip/info` | VIP 信息 |
| GET | `/api/user/vip/orders` | 我的订单 |

### 6.2 管理员 API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/admin/dashboard` | 仪表盘统计 |
| GET/POST | `/api/admin/users` | 用户列表/添加 |
| PUT | `/api/admin/users/:id/vip` | 设置 VIP |
| PUT | `/api/admin/users/:id/ban` | 封禁用户 |
| PUT | `/api/admin/users/:id/password` | 重置密码 |
| DELETE | `/api/admin/users/:id` | 删除用户 |
| GET/POST/PUT/DELETE | `/api/admin/nodes/*` | 节点管理 |
| GET/POST/PUT/DELETE | `/api/admin/tunnels/*` | 隧道管理 |
| GET/POST/PUT/DELETE | `/api/admin/subdomains/*` | 域名管理 |
| GET/POST | `/api/admin/orders` | 订单管理 |
| POST | `/api/admin/vip/grant` | 手动开通 VIP |
| GET/POST | `/api/admin/settings/*` | 系统设置 |

### 6.3 Agent API

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/agent/register` | Agent 注册，返回 frps.toml |
| POST | `/api/agent/heartbeat` | 心跳上报，返回配置更新 |
| POST | `/api/agent/get-user-vip` | 获取用户 VIP 等级 |

---

## 七、前端页面路由

### 7.1 用户端

| 路径 | 页面 | 说明 |
|------|------|------|
| `/` | HomeView | 首页 |
| `/login` | LoginView | 登录 |
| `/register` | RegisterView | 注册 |
| `/dashboard` | DashboardView | 控制台 |
| `/tunnels` | TunnelsView | 隧道管理 |
| `/subdomains` | SubdomainsView | 域名管理 |
| `/vip` | VIPView | VIP 中心 |

### 7.2 管理员端

| 路径 | 页面 | 说明 |
|------|------|------|
| `/admin/dashboard` | DashboardView | 仪表盘 |
| `/admin/users` | UsersView | 用户管理 |
| `/admin/nodes` | NodesView | 节点管理 |
| `/admin/tunnels` | TunnelsView | 隧道管理 |
| `/admin/subdomains` | SubdomainsView | 域名管理 |
| `/admin/orders` | OrdersView | VIP 订单 |
| `/admin/settings` | SettingsView | 系统设置 |

---

## 八、项目目录结构

```
JumpFrp/
├── master/                          # 主控服务 (Go + Gin)
│   ├── cmd/server/
│   │   └── main.go                 # 程序入口，数据库初始化
│   ├── internal/
│   │   ├── api/
│   │   │   ├── admin/              # 管理员 API
│   │   │   │   ├── handlers.go     # 用户/节点/隧道/VIP 处理
│   │   │   │   ├── routes.go       # 路由注册
│   │   │   │   ├── agent.go        # Agent 注册/心跳/get-user-vip
│   │   │   │   ├── settings.go     # SMTP/站点配置
│   │   │   │   ├── tunnel.go       # 管理员隧道管理
│   │   │   │   └── subdomain.go     # 管理员域名管理
│   │   │   ├── auth/              # 认证 API
│   │   │   │   ├── register.go     # 用户注册
│   │   │   │   ├── login.go        # 用户登录
│   │   │   │   └── password.go     # 密码重置
│   │   │   └── user/              # 用户 API
│   │   │       ├── routes.go       # 路由注册
│   │   │       ├── tunnel.go       # 隧道管理
│   │   │       ├── vip.go         # VIP 相关
│   │   │       └── profile.go     # 用户信息
│   │   ├── middleware/            # 中间件
│   │   │   ├── auth.go           # JWT 认证
│   │   │   ├── ratelimit.go      # 限流
│   │   │   ├── cors.go           # 跨域
│   │   │   └── security.go       # 安全响应头
│   │   ├── model/                 # 数据模型
│   │   │   ├── models.go         # Node/Tunnel/Subdomain/VIPOrder
│   │   │   └── user.go           # User 模型 + VIP 权益
│   │   ├── service/              # 业务逻辑
│   │   │   ├── auth.go           # 认证服务
│   │   │   ├── mail.go          # 邮件服务
│   │   │   ├── tunnel.go        # 隧道服务
│   │   │   └── system.go        # 系统配置 + frps.toml 生成
│   │   ├── scheduler/            # 定时任务
│   │   │   └── scheduler.go      # VIP 到期检查
│   │   └── config/               # 配置
│   │       └── config.go         # JWT/Database 配置
│   └── go.mod
│
├── agent/                           # 节点 Agent (Go)
│   ├── cmd/main.go               # 程序入口
│   └── internal/
│       ├── agent/                 # Agent 核心
│       │   └── agent.go          # 心跳/frps 管理/TC 限速
│       ├── frps/                 # frps 管理
│       │   └── manager.go       # 启动/停止/重启/热更新
│       ├── monitor/              # 系统监控
│       │   └── monitor.go       # CPU/内存采集
│       └── tc/                   # 流量控制
│           └── traffic.go       # iptables + tc 限速
│
├── frontend/                      # 前端 (Vue 3)
│   ├── src/
│   │   ├── api/
│   │   │   └── index.ts        # Axios 封装 + API 定义
│   │   ├── stores/
│   │   │   └── auth.ts        # Pinia 认证状态
│   │   ├── router/
│   │   │   └── index.ts       # 路由配置
│   │   └── views/
│   │       ├── user/           # 用户页面
│   │       │   ├── HomeView.vue
│   │       │   ├── LoginView.vue
│   │       │   ├── RegisterView.vue
│   │       │   ├── UserLayout.vue
│   │       │   ├── DashboardView.vue
│   │       │   ├── TunnelsView.vue
│   │       │   ├── SubdomainsView.vue
│   │       │   └── VIPView.vue
│   │       └── admin/         # 管理员页面
│   │           ├── AdminLayout.vue
│   │           ├── DashboardView.vue
│   │           ├── UsersView.vue
│   │           ├── NodesView.vue
│   │           ├── TunnelsView.vue
│   │           ├── SubdomainsView.vue
│   │           ├── OrdersView.vue
│   │           └── SettingsView.vue
│   ├── index.html
│   ├── vite.config.ts
│   └── package.json
│
├── scripts/                        # Shell 脚本
│   ├── install.sh                # 节点一键安装
│   │                            #   - 检查依赖 (tc/iptables)
│   │                            #   - 下载 frps + Agent
│   │                            #   - 生成配置
│   │                            #   - 注册 systemd 服务
│   │                            #   - 启动服务
│   └── uninstall.sh             # 节点卸载
│
├── Dockerfile                     # 多阶段构建
├── fly.toml                      # Fly.io 配置
├── .dockerignore                 # Docker 构建忽略
├── .env                          # 开发环境变量
├── .env.production              # 生产环境变量
├── PROJECT.md                    # 项目文档
├── README.md                     # 项目说明
└── DEVELOPMENT.md                # 完整开发文档
```

---

## 九、部署方案

### 9.1 Fly.io 主控部署

```bash
# 1. 登录并创建应用
fly auth login
fly apps create jumpfrp

# 2. 创建存储卷
fly volumes create jumpfrp_data --size 1 --region sin --app jumpfrp

# 3. 设置环境变量
fly secrets set JWT_SECRET="$(openssl rand -base64 48)" --app jumpfrp
fly secrets set GIN_MODE="release" --app jumpfrp

# 4. 部署
fly deploy --app jumpfrp

# 5. 验证
curl https://api.jumpfrp.top/health
```

### 9.2 节点服务器安装

```bash
# 从管理后台复制安装命令，格式如下：
bash <(wget -qO- https://api.jumpfrp.top/install.sh) \
  --node-id <节点标识> \
  --token <Agent Token> \
  --master-url https://api.jumpfrp.top \
  --frps-port 7000 \
  --agent-port 7500

# 可选参数：
#   --proxy true    # 使用代理下载（适用于中国大陆服务器）
#   --no-proxy      # 禁用代理（默认）
```

### 9.3 Agent 发布流程

```bash
cd agent

# 编译 amd64
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o jumpfrp-agent-linux-amd64 ./cmd/main.go

# 编译 arm64
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o jumpfrp-agent-linux-arm64 ./cmd/main.go

# 压缩（减小体积）
gzip -k jumpfrp-agent-linux-amd64
gzip -k jumpfrp-agent-linux-arm64

# 上传到 GitHub Releases:
# https://github.com/AKEXZ/JumpFrp/releases
```

---

## 十、已解决的技术问题

### 10.1 数据库问题

| 问题 | 原因 | 解决方案 |
|------|------|---------|
| SQLite CGO 报错 | `gorm.io/driver/sqlite` 需要 CGO | 使用 `github.com/glebarez/sqlite` |
| 列名不匹配 | GORM 列名映射问题 | 添加显式 `column:` tag + 启动时自动添加缺失列 |
| `min_vip_level` 列不存在 | 旧数据库缺少列 | `fixMissingColumns()` 自动修复 |
| `vip_level` 列不存在 | 同上 | 同上 |

### 10.2 前端问题

| 问题 | 原因 | 解决方案 |
|------|------|---------|
| 前端 404 | 未复制前端产物 | Dockerfile 多阶段构建 |
| API 请求 localhost | 生产用了开发地址 | `.env.production` 设置 `/api` |
| TypeScript 错误 | axios 响应类型 | 添加 `: any` 类型断言 |

### 10.3 frp 版本问题

| 问题 | 原因 | 解决方案 |
|------|------|---------|
| frps 启动失败 | frp 0.61.0 改用
### 10.3 frp 版本问题

| 问题 | 原因 | 解决方案 |
|------|------|---------|
| frps 启动失败 | frp 0.61.0 改用 TOML | 生成 `frps.toml` 配置文件 |
| frpc 配置 INI 格式 | 旧版本生成 INI | 改为生成 TOML 格式 |
| 配置文件下载失败 | 被 axios 拦截器拦截 | 使用 fetch 直接下载 |
| 配置文件名错误 | 命名为 `frpc.ini` | 改为 `frpc.toml` |

### 10.4 Fly.io 部署问题

| 问题 | 原因 | 解决方案 |
|------|------|---------|
| 健康检查超时 | `[checks]` 配置太严格 | 删除 check 配置 |
| hkg 区域废弃 | Fly.io 已弃用香港区域 | 改用 `sin`（新加坡） |
| fly.toml 格式错误 | `[[http_service]]` 应为 `[http_service]` | 使用单括号 |
| 构建上下文太大 | node_modules 等未排除 | 添加 `.dockerignore` |
| go.mod 版本冲突 | 本地 Go 1.25+ | Dockerfile 使用 `golang:1.24-alpine` |

### 10.5 其他问题

| 问题 | 原因 | 解决方案 |
|------|------|---------|
| SMTP 未配置却提示成功 | 未检查 SMTP 配置 | 添加 SMTP 配置检查 |
| 节点列表为空 | Select 列名不匹配 | 移除 Select 子句 |
| Token 无效仍显示登录 | 未验证 token | 启动时验证 token 有效性 |
| 用户删除后仍显示登录 | 前端只检查 localStorage | 启动时调用 /user/profile 验证 |
| 本地端口冲突 | 未检查重复 | 创建隧道时检查 local_ip:local_port 组合 |

---

## 十一、业务规则

### 11.1 隧道创建规则

1. **检查用户 VIP 等级与隧道数量限制**
2. **检查协议权限**（Free 只能用 TCP）
3. **检查节点可用性**（在线状态、VIP 等级要求）
4. **检查隧道名称唯一性**（同用户下唯一）
5. **检查本地端口冲突**（同用户下同一 local_ip:local_port 不能重复）
6. **分配远程端口**（从节点端口池随机分配）
7. **设置带宽限制**（根据用户 VIP 等级）
8. **HTTP/HTTPS 协议可选绑定子域名**

### 11.2 域名绑定规则

1. **VIP 等级要求**：Pro（VIPLevel >= 2）才能绑定域名
2. **协议限制**：仅 HTTP/HTTPS 隧道支持绑定域名
3. **子域名唯一性**：全局唯一，不能重复
4. **自动批准**：用户申请的域名自动批准，无需管理员审核

### 11.3 VIP 升级规则

1. **管理员手动开通**：通过后台设置 VIP 等级和到期天数
2. **自动延期**：用户已有 VIP 时，在现有基础上延期
3. **带宽自动调整**：升级时所有现有隧道的带宽限制自动更新到新等级上限

### 11.4 节点管理规则

1. **节点标识自动生成**：创建时如未填写，自动生成 `node-xxx` 格式
2. **节点标识必填**：IP 地址、地区、节点名称为必填项
3. **Agent Token 自动生成**：32 位随机字符串
4. **心跳超时检测**：2 分钟无心跳标记为离线
5. **安装命令动态生成**：包含节点标识和 Token

---

## 十二、安全机制

### 12.1 认证与授权

| 机制 | 实现 |
|------|------|
| JWT Token | RS256 算法，24 小时有效期 |
| Token 刷新 | 登录时自动刷新 |
| 管理员识别 | username == "admin" |
| API Token | 用户 frpc 认证使用独立 Token |

### 12.2 限流

| 接口 | 限制 |
|------|------|
| 全局 | 120 次/分钟/IP |
| 登录/注册 | 10 次/分钟/IP |
| 发送验证码 | 5 次/小时/邮箱 |

### 12.3 安全响应头

```go
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
X-Content-Type-Options: nosniff
Referrer-Policy: strict-origin-when-cross-origin
Content-Security-Policy: default-src 'self'
```

---

## 十三、frps.toml 配置模板

### 13.1 主控生成配置

由 `SystemService.GenerateFrpsConfig()` 生成：

```toml
# frps.toml - JumpFrp 服务端配置
# 由主控自动生成，请勿手动修改

bindPort = 7000
auth.method = "token"
auth.token = "节点AgentToken"

[transport]
max_pool_count = 100
pool_count = 10
tcp_mux = true
transport.tcp_mux = true

[log]
to = "/var/log/frps.log"
level = "info"
max_days = 3

[[vhost.httpRoutes]]
custom_domains = ["*.jumpfrp.top"]

[[vhost.httpsRoutes]]
custom_domains = ["*.jumpfrp.top"]
```

### 13.2 带宽限制（可选）

在 transport 中添加：

```toml
[transport]
# 每客户端最大带宽（服务端强制）
transport.max_bandwidth_per_client = "20MB"
```

---

## 十四、开发指南

### 14.1 本地开发

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

# 访问
# 前端: http://localhost:5173
# 后端: http://localhost:8080
```

### 14.2 默认账户

| 角色 | 用户名 | 密码 | 邮箱 |
|------|--------|------|------|
| 管理员 | admin | admin123456 | admin@jumpfrp.top |

⚠️ **生产环境必须修改默认密码！**

### 14.3 环境变量

| 变量 | 说明 | 示例 |
|------|------|------|
| JWT_SECRET | JWT 签名密钥 | `openssl rand -base64 48` |
| GIN_MODE | 运行模式 | `release` |
| DATABASE_PATH | 数据库路径 | `/data/jumpfrp.db` |

### 14.4 数据库调试

```bash
# 连接 Fly.io 容器
fly ssh console -a jumpfrp

# 容器内无 sqlite3，使用 Go 调试
# 查看日志
fly logs -a jumpfrp -n

# 重启
fly machines restart -a jumpfrp
```

---

## 十五、未来计划

- [ ] **支付接口集成**：对接 Stripe/PayPal/支付宝/微信支付
- [ ] **流量统计增强**：按日/按隧道详细流量报表
- [ ] **隧道在线状态**：实时检测隧道连通性
- [ ] **WebSocket 推送**：节点状态变化实时通知
- [ ] **API 密钥管理**：用户自主管理 API 密钥
- [ ] **多语言支持**：国际化（i18n）
- [ ] **黑暗模式**：主题切换
- [ ] **隧道分组**：按项目分组管理隧道
- [ ] **团队协作**：多人协作管理隧道

---

## 附录 A：VIP 套餐配置

| 套餐 | ID | 价格 (CNY) | 天数 | 隧道数 | 端口数 | 带宽 | 协议 | 子域名 |
|------|-----|-----------|------|--------|--------|------|------|--------|
| Free | 0 | 免费 | 永久 | 1 | 3 | 1 Mbps | TCP | ✗ |
| Basic | 1 | ¥9.9 | 30 | 5 | 10 | 5 Mbps | TCP/UDP | ✗ |
| Pro | 2 | ¥29.9 | 30 | 20 | 50 | 20 Mbps | 全协议 | ✓ |
| Ultimate | 3 | ¥99 | 30 | ∞ | 200 | 100 Mbps | 全协议 | ✓ |

## 附录 B：端口池分配规则

1. 每个节点有独立的端口池（默认 10000-20000）
2. 端口分配时随机选择可用端口
3. 可通过 `port_excludes` 排除特定端口
4. 端口分配后与用户绑定，不回收

## 附录 C：心跳机制

```
Agent → 主控 (每 30 秒)
├── POST /api/agent/heartbeat
├── Body: {node_id, token, cpu_usage, memory_usage, current_conns, version}
└── Response: {code, msg, frps_config, config_version}

主控 → Agent (配置变更时)
├── 心跳响应包含新的 frps.toml 配置
└── Agent 保存配置并热重启 frps
```

## 附录 D：日志文件位置

| 组件 | 路径 |
|------|------|
| frps 日志 | `/var/log/frps.log` |
| Agent 日志 | systemd journal |
| 主控日志 | stdout (Fly.io) |

---

*文档版本：v2.0*
*最后更新：2026-03-28*
*维护者：JumpFrp Team*
