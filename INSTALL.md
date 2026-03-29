# JumpFrp 节点安装指南 v1.1.0

## 快速安装

在你的节点服务器上运行以下命令（需要 root 权限）：

```bash
bash <(wget -qO- https://raw.githubusercontent.com/AKEXZ/JumpFrp/main/scripts/install.sh) \
  --node-id <节点标识> \
  --token <节点Token>
```

### 参数说明

| 参数 | 说明 | 示例 |
|------|------|------|
| `--node-id` | 节点标识（从管理后台获取） | `sh-01` |
| `--token` | 节点 Token（从管理后台获取） | `abc123...` |
| `--master-url` | 主控地址（可选，默认 https://api.jumpfrp.top） | `https://api.jumpfrp.top` |
| `--frps-port` | frps 端口（可选，默认 7000） | `7000` |
| `--agent-port` | Agent 端口（可选，默认 7500） | `7500` |
| `--use-proxy` | 使用代理下载（可选） | - |
| `--no-proxy` | 不使用代理下载（可选，默认） | - |

### 完整示例

```bash
# 基础安装
bash <(wget -qO- https://raw.githubusercontent.com/AKEXZ/JumpFrp/main/scripts/install.sh) \
  --node-id sh-01 \
  --token 2ff919368e325325830a67322a2700471e8f1dd657a496e0b352339921bd9c7b

# 使用代理下载（网络不好时）
bash <(wget -qO- https://raw.githubusercontent.com/AKEXZ/JumpFrp/main/scripts/install.sh) \
  --node-id sh-01 \
  --token 2ff919368e325325830a67322a2700471e8f1dd657a496e0b352339921bd9c7b \
  --use-proxy

# 自定义端口
bash <(wget -qO- https://raw.githubusercontent.com/AKEXZ/JumpFrp/main/scripts/install.sh) \
  --node-id sh-01 \
  --token 2ff919368e325325830a67322a2700471e8f1dd657a496e0b352339921bd9c7b \
  --frps-port 8000 \
  --agent-port 8500
```

## 安装步骤

脚本会自动执行以下步骤：

1. ✅ **检查系统依赖** - 检查并安装必要的工具（tc, iptables, wget, curl 等）
2. ✅ **下载 frps** - 从 GitHub 下载 frp 0.61.0 服务端
3. ✅ **下载 Agent** - 从 GitHub Releases 下载 JumpFrp Agent v1.1.0
4. ✅ **创建配置文件** - 生成 Agent 配置和 frps 基础配置
5. ✅ **创建系统服务** - 注册 systemd 服务，开机自启
6. ✅ **启动服务** - 启动 Agent 并向主控注册

## 验证安装

### 1. 检查服务状态

```bash
systemctl status jumpfrp-agent
```

应该看到 `active (running)` 状态。

### 2. 查看实时日志

```bash
journalctl -u jumpfrp-agent -f
```

应该看到类似的输出：
```
[配置] 检测到新的 frps 配置，正在更新...
[配置] frps 配置已更新至版本 1
```

### 3. 检查管理后台

访问 https://jumpfrp.top，进入管理后台：
- 进入 **节点管理**
- 查看你的节点状态是否为 **在线**（绿色）

## 常见问题

### Q: 安装后节点显示离线？

A: 检查以下几点：

```bash
# 1. 查看 Agent 日志
journalctl -u jumpfrp-agent -n 50

# 2. 检查防火墙
# 确保以下端口已开放：
# - frps 端口（默认 7000）
# - Agent 端口（默认 7500）

# 3. 检查网络连接
ping api.jumpfrp.top

# 4. 重启 Agent
systemctl restart jumpfrp-agent
```

### Q: 如何更新 Agent？

A: 运行更新脚本：

```bash
curl -O https://raw.githubusercontent.com/AKEXZ/JumpFrp/main/scripts/update-agent.sh
chmod +x update-agent.sh
sudo ./update-agent.sh
```

### Q: 如何卸载？

A: 

```bash
# 停止服务
systemctl stop jumpfrp-agent
systemctl disable jumpfrp-agent

# 删除文件
rm -rf /opt/jumpfrp
rm /etc/systemd/system/jumpfrp-agent.service

# 重新加载 systemd
systemctl daemon-reload
```

### Q: 如何修改端口？

A: 编辑配置文件并重启：

```bash
# 编辑配置
nano /opt/jumpfrp/agent.env

# 修改 FRPS_PORT 和 AGENT_PORT

# 重启服务
systemctl restart jumpfrp-agent
```

## 系统要求

- **操作系统**：Ubuntu 20.04+ 或 Debian 10+
- **权限**：root 用户
- **网络**：能访问 GitHub 和主控地址
- **依赖**：tc, iptables, wget, curl, tar, gzip

## 文件位置

| 文件 | 位置 |
|------|------|
| Agent 二进制 | `/opt/jumpfrp/jumpfrp-agent` |
| frps 二进制 | `/opt/jumpfrp/frps` |
| Agent 配置 | `/opt/jumpfrp/agent.env` |
| frps 配置 | `/opt/jumpfrp/frps.toml` |
| frps 日志 | `/var/log/frps.log` |
| systemd 服务 | `/etc/systemd/system/jumpfrp-agent.service` |

## 支持

如有问题，请：

1. 查看日志：`journalctl -u jumpfrp-agent -n 100`
2. 检查网络：`ping api.jumpfrp.top`
3. 联系管理员

## 更新日志

### v1.1.0 - 2026-03-29

- ✨ Agent 启动时立即获取 frps 配置（2s 延迟）
- ✨ 支持每用户独立 Token 认证
- ✨ 动态 frps 配置更新（版本控制）
- ✨ 配置变化时自动重启 frps
- 🐛 修复 frpc 连接失败（token 不匹配）
- 📈 改进配置版本跟踪和自动同步
