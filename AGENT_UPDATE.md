# Agent 更新指南 v1.1.0

## 快速更新（推荐）

在节点服务器上运行：

```bash
# 下载更新脚本
curl -O https://raw.githubusercontent.com/AKEXZ/JumpFrp/main/scripts/update-agent.sh
chmod +x update-agent.sh

# 运行更新（需要 sudo）
sudo ./update-agent.sh
```

脚本会自动：
1. ✅ 检查依赖（git, Go）
2. ✅ 克隆最新代码
3. ✅ 编译新 Agent
4. ✅ 备份旧版本
5. ✅ 安装新版本
6. ✅ 重启 Agent 服务

## 手动更新

如果脚本不可用，可以手动更新：

```bash
# 1. 停止 Agent
sudo systemctl stop jumpfrp-agent

# 2. 克隆/更新代码
cd /tmp
git clone https://github.com/AKEXZ/JumpFrp.git
cd JumpFrp/agent

# 3. 编译
go build -o jumpfrp-agent ./cmd/main.go

# 4. 安装
sudo cp jumpfrp-agent /opt/jumpfrp/jumpfrp-agent
sudo chmod +x /opt/jumpfrp/jumpfrp-agent

# 5. 启动
sudo systemctl start jumpfrp-agent

# 6. 验证
sudo systemctl status jumpfrp-agent
```

## 验证更新

更新后，检查 Agent 日志：

```bash
# 查看最近日志
sudo tail -f /var/log/jumpfrp-agent.log

# 应该看到类似的输出：
# [配置] 检测到新的 frps 配置，正在更新...
# [配置] frps 配置已更新至版本 X
```

## 更新内容

### v1.1.0 - 2026-03-29

**新功能**：
- ✨ Agent 启动时立即获取 frps 配置（2s 延迟）
- ✨ 支持每用户独立 Token 认证
- ✨ 动态 frps 配置更新（版本控制）
- ✨ 配置变化时自动重启 frps

**修复**：
- 🐛 修复 frpc 连接失败（token 不匹配）
- 🐛 修复新用户注册后 frps 不知道新 token 的问题

**改进**：
- 📈 配置版本跟踪
- 📈 自动配置同步
- 📈 更好的错误日志

## 常见问题

### Q: 更新后 frps 还是用旧配置？
A: Agent 启动后 2 秒会自动从主控获取新配置。如果还是不行，检查：
```bash
# 查看 frps 配置
cat /opt/jumpfrp/frps.toml

# 查看 Agent 日志
sudo journalctl -u jumpfrp-agent -n 50
```

### Q: 更新失败了怎么办？
A: 脚本会自动恢复旧版本。如果需要手动恢复：
```bash
sudo cp /opt/jumpfrp/jumpfrp-agent.bak /opt/jumpfrp/jumpfrp-agent
sudo systemctl restart jumpfrp-agent
```

### Q: 如何回滚到旧版本？
A: 
```bash
# 恢复备份
sudo cp /opt/jumpfrp/jumpfrp-agent.bak /opt/jumpfrp/jumpfrp-agent

# 重启
sudo systemctl restart jumpfrp-agent
```

## 支持

如有问题，请检查：
1. Agent 日志：`sudo journalctl -u jumpfrp-agent -n 100`
2. frps 日志：`sudo tail -f /var/log/frps.log`
3. 主控日志：查看 Fly.io 控制面板

或联系管理员。
