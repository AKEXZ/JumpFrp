#!/bin/bash
# JumpFrp 节点一键安装脚本
# 用法: bash <(wget -qO- https://api.jumpfrp.top/install.sh) --node-id xxx --token xxx

set -e

# 颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 默认参数
NODE_ID=""
TOKEN=""
MASTER_URL="https://api.jumpfrp.top"
FRPS_PORT=7000
AGENT_PORT=7500

# 解析参数
while [[ $# -gt 0 ]]; do
  case $1 in
    --node-id) NODE_ID="$2"; shift 2 ;;
    --token) TOKEN="$2"; shift 2 ;;
    --master-url) MASTER_URL="$2"; shift 2 ;;
    --frps-port) FRPS_PORT="$2"; shift 2 ;;
    --agent-port) AGENT_PORT="$2"; shift 2 ;;
    *) echo "未知参数: $1"; exit 1 ;;
  esac
done

# 检查必需参数
if [[ -z "$NODE_ID" || -z "$TOKEN" ]]; then
  echo -e "${RED}错误: 必须指定 --node-id 和 --token${NC}"
  echo "用法: bash install.sh --node-id <节点标识> --token <Token>"
  exit 1
fi

echo -e "${GREEN}=== JumpFrp 节点安装程序 ===${NC}"
echo "节点标识: $NODE_ID"
echo "主控地址: $MASTER_URL"
echo "frps 端口: $FRPS_PORT"
echo "Agent 端口: $AGENT_PORT"
echo ""
echo -e "${YELLOW}安装步骤:${NC}"
echo "  [1/5] 下载 frps 服务端"
echo "  [2/5] 下载 JumpFrp Agent"
echo "  [3/5] 创建配置文件"
echo "  [4/5] 创建系统服务"
echo "  [5/5] 启动服务并注册"
echo ""

# 检查系统
if [[ ! -f /etc/os-release ]]; then
  echo -e "${RED}错误: 无法识别操作系统${NC}"
  exit 1
fi

source /etc/os-release
if [[ "$ID" != "ubuntu" && "$ID" != "debian" ]]; then
  echo -e "${YELLOW}警告: 当前系统 $ID 可能不受支持，建议 Ubuntu 20.04+${NC}"
  read -p "是否继续? [y/N] " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    exit 1
  fi
fi

# 检查 root
if [[ $EUID -ne 0 ]]; then
  echo -e "${RED}错误: 请使用 root 用户运行此脚本${NC}"
  exit 1
fi

# 安装目录
INSTALL_DIR="/opt/jumpfrp"
mkdir -p $INSTALL_DIR

# 下载 frps（从 GitHub 或国内镜像）
echo -e "${GREEN}[1/5] 下载 frps...${NC}"
FRPS_VERSION="0.61.0"
ARCH=$(uname -m)
if [[ "$ARCH" == "x86_64" ]]; then
  FRPS_ARCH="amd64"
elif [[ "$ARCH" == "aarch64" ]]; then
  FRPS_ARCH="arm64"
else
  echo -e "${RED}不支持的架构: $ARCH${NC}"
  exit 1
fi

FRPS_URL="https://github.com/fatedier/frp/releases/download/v${FRPS_VERSION}/frp_${FRPS_VERSION}_linux_${FRPS_ARCH}.tar.gz"
FRPS_MIRROR="https://gitproxy.ake.cx/${FRPS_URL}"

# 尝试下载（优先使用代理）
echo "下载地址: $FRPS_MIRROR"
echo -e "${YELLOW}正在连接...${NC}"
if wget --progress=bar:force --timeout=60 -O /tmp/frp.tar.gz "$FRPS_MIRROR" 2>&1; then
  echo -e "${GREEN}frps 下载完成${NC}"
else
  echo -e "${YELLOW}代理下载失败，尝试直连 GitHub...${NC}"
  if wget --progress=bar:force --timeout=60 -O /tmp/frp.tar.gz "$FRPS_URL" 2>&1; then
    echo -e "${GREEN}frps 下载完成${NC}"
  else
    echo -e "${RED}下载 frps 失败，请检查网络连接${NC}"
    echo "提示：如果在中国大陆，请确保代理可用或手动下载后上传到服务器"
    exit 1
  fi
fi

tar -xzf /tmp/frp.tar.gz -C /tmp/
cp /tmp/frp_*/frps $INSTALL_DIR/frps
chmod +x $INSTALL_DIR/frps
rm -rf /tmp/frp_*
echo -e "${GREEN}frps 已安装 (v${FRPS_VERSION})${NC}"

# 下载 Agent（从主控下载压缩文件）
echo -e "${GREEN}[2/5] 下载 JumpFrp Agent...${NC}"
AGENT_URL="${MASTER_URL}/download/agent-linux-${FRPS_ARCH}.gz"

echo "下载地址: $AGENT_URL"
echo -e "${YELLOW}正在连接主控...${NC}"
if wget --progress=bar:force --timeout=60 -O /tmp/agent.gz "$AGENT_URL" 2>&1; then
  # 解压
  echo -e "${YELLOW}正在解压...${NC}"
  if gunzip -f /tmp/agent.gz && mv /tmp/agent $INSTALL_DIR/agent; then
    chmod +x $INSTALL_DIR/agent
    # 验证是否为有效的 ELF 可执行文件
    if file $INSTALL_DIR/agent | grep -q "ELF"; then
      echo -e "${GREEN}Agent 安装完成${NC}"
    else
      echo -e "${RED}Agent 解压失败：文件格式不正确${NC}"
      rm -f $INSTALL_DIR/agent
      exit 1
    fi
  else
    echo -e "${RED}Agent 解压失败${NC}"
    exit 1
  fi
else
  echo -e "${RED}下载 Agent 失败${NC}"
  echo "提示：请检查主控服务是否正常运行"
  exit 1
fi

# 创建配置文件
echo -e "${GREEN}[3/5] 创建配置文件...${NC}"
cat > $INSTALL_DIR/agent.env << EOF
AGENT_NODE_ID=${NODE_ID}
AGENT_TOKEN=${TOKEN}
AGENT_MASTER_URL=${MASTER_URL}
AGENT_PORT=${AGENT_PORT}
FRPS_PORT=${FRPS_PORT}
EOF

# 创建 frps.toml 配置 (TOML格式，frp 0.61.0+)
cat > $INSTALL_DIR/frps.toml << EOF
# frps 服务端配置
bindPort = ${FRPS_PORT}
auth.method = "token"
auth.token = "${TOKEN}"

# Web 面板（可选）
webServer.addr = "0.0.0.0"
webServer.port = $((FRPS_PORT + 1))
webServer.user = "admin"
webServer.password = "jumpfrp-dashboard"

# 日志
log.to = "/var/log/frps.log"
log.level = "info"
log.maxDays = 3

# 传输层安全（建议生产环境启用）
# transport.tls.force = true
EOF

echo -e "${GREEN}配置文件已创建 (TOML格式)${NC}"

# 创建 systemd 服务
echo -e "${GREEN}[4/5] 创建系统服务...${NC}"

# frps 服务
cat > /etc/systemd/system/frps.service << 'EOF'
[Unit]
Description=JumpFrp Server
After=network.target

[Service]
Type=simple
ExecStart=/opt/jumpfrp/frps -c /opt/jumpfrp/frps.toml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# Agent 服务
cat > /etc/systemd/system/jumpfrp-agent.service << EOF
[Unit]
Description=JumpFrp Agent
After=network.target frps.service
Requires=frps.service

[Service]
Type=simple
EnvironmentFile=/opt/jumpfrp/agent.env
ExecStart=/opt/jumpfrp/agent --node-id \${AGENT_NODE_ID} --token \${AGENT_TOKEN} --master-url \${AGENT_MASTER_URL} --frps-port \${FRPS_PORT} --agent-port \${AGENT_PORT}
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# 启动服务
echo -e "${GREEN}[5/5] 启动服务...${NC}"
systemctl daemon-reload
systemctl enable frps jumpfrp-agent 2>/dev/null

echo -e "${YELLOW}正在启动 frps...${NC}"
if systemctl start frps; then
  sleep 2
  if systemctl is-active --quiet frps; then
    echo -e "${GREEN}✓ frps 服务运行正常 (端口 $FRPS_PORT)${NC}"
  else
    echo -e "${RED}✗ frps 服务启动失败${NC}"
    echo "  查看错误: journalctl -u frps -n 20 --no-pager"
  fi
else
  echo -e "${RED}✗ frps 启动命令执行失败${NC}"
fi

echo -e "${YELLOW}正在启动 Agent...${NC}"
if systemctl start jumpfrp-agent; then
  sleep 2
  if systemctl is-active --quiet jumpfrp-agent; then
    echo -e "${GREEN}✓ Agent 服务运行正常${NC}"
    echo -e "${YELLOW}  正在向主控注册节点，请稍候...${NC}"
    sleep 3
  else
    echo -e "${RED}✗ Agent 服务启动失败${NC}"
    echo "  查看错误: journalctl -u jumpfrp-agent -n 20 --no-pager"
  fi
else
  echo -e "${RED}✗ Agent 启动命令执行失败${NC}"
fi

echo ""
echo -e "${GREEN}══════════════════════════════════════${NC}"
echo -e "${GREEN}        安装完成${NC}"
echo -e "${GREEN}══════════════════════════════════════${NC}"
echo ""
echo "安装信息:"
echo "  ├─ 安装目录: $INSTALL_DIR"
echo "  ├─ frps 端口: $FRPS_PORT"
echo "  ├─ Agent 端口: $AGENT_PORT"
echo "  └─ 节点标识: $NODE_ID"
echo ""
echo "常用命令:"
echo "  ├─ 查看状态: systemctl status frps jumpfrp-agent"
echo "  ├─ 查看日志: journalctl -u frps -u jumpfrp-agent -f"
echo "  └─ 重启服务: systemctl restart frps jumpfrp-agent"
echo ""
echo -e "${YELLOW}提示: 请前往管理后台确认节点状态已变为「在线」${NC}"
echo -e "${YELLOW}如果显示离线，请检查: ${NC}"
echo "  1. 服务器防火墙是否开放端口 $FRPS_PORT 和 $AGENT_PORT"
echo "  2. 运行 'journalctl -u jumpfrp-agent -n 50' 查看详细错误"
