# JumpFrp 部署文档

## 架构说明

```
用户浏览器
    │
    ▼
Fly.io (主控服务 + 前端静态文件)
    │ API 通信
    ▼
节点服务器 (Ubuntu VPS)
  frps + jumpfrp-agent
    ▲
    │ frpc 连接
用户设备 (官方 frpc)
```

---

## 一、主控服务部署（Fly.io）

### 1. 安装 flyctl

```bash
curl -L https://fly.io/install.sh | sh
fly auth login
```

### 2. 构建前端

```bash
cd frontend
npm install && npm run build
# 将 dist/ 复制到 master/web/
cp -r dist/ ../master/web/
```

### 3. 主控嵌入前端静态文件

在 `master/cmd/server/main.go` 中添加静态文件服务：

```go
// 前端静态文件
r.Static("/assets", "./web/assets")
r.StaticFile("/", "./web/index.html")
r.NoRoute(func(c *gin.Context) {
    c.File("./web/index.html") // SPA fallback
})
```

### 4. 创建 Dockerfile

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY master/ .
RUN go build -o jumpfrp-master ./cmd/server

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/jumpfrp-master .
COPY master/web/ ./web/
COPY scripts/ ./scripts/
RUN mkdir -p /data
EXPOSE 8080
CMD ["./jumpfrp-master"]
```

### 5. 创建 fly.toml

```toml
app = "jumpfrp-master"
primary_region = "hkg"

[build]

[env]
  APP_MODE = "release"
  SERVER_ADDR = ":8080"
  DB_PATH = "/data/jumpfrp.db"

[mounts]
  source = "jumpfrp_data"
  destination = "/data"

[[services]]
  protocol = "tcp"
  internal_port = 8080

  [[services.ports]]
    port = 80
    handlers = ["http"]

  [[services.ports]]
    port = 443
    handlers = ["tls", "http"]
```

### 6. 部署

```bash
cd master
fly launch --no-deploy
fly volumes create jumpfrp_data --size 1
fly secrets set JWT_SECRET="your-strong-secret-here"
fly secrets set SMTP_HOST="smtp.example.com"
fly secrets set SMTP_USER="noreply@jumpfrp.top"
fly secrets set SMTP_PASS="your-smtp-password"
fly deploy
```

### 7. 绑定自定义域名

```bash
fly certs add jumpfrp.top
fly certs add api.jumpfrp.top
fly certs add "*.jumpfrp.top"
```

---

## 二、节点服务器部署（Ubuntu VPS）

### 前提条件
- Ubuntu 20.04 或更高版本
- Root 权限
- 开放端口：frps 端口（默认 7000）、Agent 端口（默认 7500）

### 一键安装

在管理员后台 → 节点管理 → 点击"安装命令"，复制命令后在服务器上执行：

```bash
bash <(wget -qO- https://api.jumpfrp.top/install.sh) \
  --node-id sh-01 \
  --token xxxxxxxxxxxxxxxx \
  --master-url https://api.jumpfrp.top \
  --frps-port 7000 \
  --agent-port 7500
```

### 服务管理

```bash
# 查看状态
systemctl status frps
systemctl status jumpfrp-agent

# 查看日志
journalctl -u frps -f
journalctl -u jumpfrp-agent -f

# 重启服务
systemctl restart frps jumpfrp-agent

# 卸载
bash <(wget -qO- https://api.jumpfrp.top/uninstall.sh)
```

---

## 三、域名 DNS 配置

| 记录类型 | 主机名 | 值 |
|---------|--------|-----|
| A | `@` | Fly.io IP |
| A | `api` | Fly.io IP |
| A | `*` | 节点服务器 IP（子域名穿透用） |
| CNAME | `www` | `jumpfrp.top` |

> 通配符 `*.jumpfrp.top` 指向节点服务器，用于 HTTP/HTTPS 子域名穿透。

---

## 四、环境变量说明

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `APP_MODE` | 运行模式 (debug/release) | `debug` |
| `SERVER_ADDR` | 监听地址 | `:8080` |
| `DB_PATH` | SQLite 数据库路径 | `./data/jumpfrp.db` |
| `JWT_SECRET` | JWT 签名密钥（**必须修改**） | 默认值不安全 |
| `SMTP_HOST` | SMTP 服务器地址 | 空（不发邮件） |
| `SMTP_USER` | SMTP 用户名 | 空 |
| `SMTP_PASS` | SMTP 密码 | 空 |
| `SMTP_FROM` | 发件人地址 | `noreply@jumpfrp.top` |

---

## 五、首次使用

1. 访问 `https://jumpfrp.top`
2. 进入 `https://jumpfrp.top/admin` 用 `admin / admin123456` 登录
3. **立即修改管理员密码！**
4. 在节点管理中添加节点，复制安装命令到服务器执行
5. 等待节点上线后，用户即可创建隧道

---

## 六、常见问题

**Q: 节点一直显示离线？**
- 检查节点服务器防火墙是否开放 Agent 端口（默认 7500）
- 检查主控地址是否正确：`journalctl -u jumpfrp-agent -f`

**Q: 用户连接 frpc 失败？**
- 检查 frps 端口（默认 7000）是否开放
- 检查 frpc.ini 中的 token 是否与 frps 配置一致

**Q: 邮件发送失败？**
- 检查 SMTP 配置是否正确
- 部分邮件服务商需要开启"应用专用密码"

**Q: SQLite 数据备份？**
```bash
# Fly.io 上备份
fly ssh console -C "cp /data/jumpfrp.db /data/jumpfrp.db.bak"
fly sftp get /data/jumpfrp.db ./backup/
```
