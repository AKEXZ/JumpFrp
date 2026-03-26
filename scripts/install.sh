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

# 尝试下载
if ! wget -q --timeout=30 -O /tmp/frp.tar.gz "$FRPS_URL" 2>/dev/null; then
  echo "尝试镜像下载..."
  wget -q --timeout=60 -O /tmp/frp.tar.gz "$FRPS_MIRROR" || {
    echo -e "${RED}下载 frps 失败${NC}"
    exit 1
  }
fi

tar -xzf /tmp/frp.tar.gz -C /tmp/
cp /tmp/frp_*/frps $INSTALL_DIR/frps
chmod +x $INSTALL_DIR/frps
rm -rf /tmp/frp_*
echo -e "${GREEN}frps 已安装${NC}"

# 下载 Agent
echo -e "${GREEN}[2/5] 下载 JumpFrp Agent...${NC}"
AGENT_URL="${MASTER_URL}/download/agent-linux-${FRPS_ARCH}"
AGENT_MIRROR="https://ghproxy.com/${AGENT_URL}"

if ! wget -q --timeout=30 -O $INSTALL_DIR/agent "$AGENT_URL" 2>/dev/null; then
  echo "尝试镜像下载..."
  wget -q --timeout=60 -O $INSTALL_DIR/agent "$AGENT_MIRROR" || {
    echo -e "${RED}下载 Agent 失败${NC}"
    exit 1
  }
fi
chmod +x $INSTALL_DIR/agent
echo -e "${GREEN}Agent 已安装${NC}"

# 创建配置文件
echo -e "${GREEN}[3/5] 创建配置文件...${NC}"
cat > $INSTALL_DIR/agent.env << EOF
AGENT_NODE_ID=${NODE_ID}
AGENT_TOKEN=${TOKEN}
AGENT_MASTER_URL=${MASTER_URL}
AGENT_PORT=${AGENT_PORT}
FRPS_PORT=${FRPS_PORT}
EOF

# 创建 frps 配置
cat > $INSTALL_DIR/frps.ini << EOF
[common]
bind_port = ${FRPS_PORT}
token = ${TOKEN}
dashboard_port = $((FRPS_PORT + 1))
dashboard_user = admin
dashboard_pwd = jumpfrp-dashboard
log_file = /var/log/frps.log
log_level = info
log_max_days = 3
EOF

# 创建 systemd 服务
echo -e "${GREEN}[4/5] 创建系统服务...${NC}"

# frps 服务
cat > /etc/systemd/system/frps.service << 'EOF'
[Unit]
Description=JumpFrp Server
After=network.target

[Service]
Type=simple
ExecStart=/opt/jumpfrp/frps -c /opt/jumpfrp/frps.ini
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
systemctl enable frps jumpfrp-agent
systemctl start frps jumpfrp-agent

# 检查状态
sleep 2
if systemctl is-active --quiet frps; then
  echo -e "${GREEN}frps 服务运行正常${NC}"
else
  echo -e "${RED}frps 服务启动失败，请检查日志: journalctl -u frps${NC}"
fi

if systemctl is-active --quiet jumpfrp-agent; then
  echo -e "${GREEN}Agent 服务运行正常${NC}"
else
  echo -e "${RED}Agent 服务启动失败，请检查日志: journalctl -u jumpfrp-agent${NC}"
fi

echo ""
echo -e "${GREEN}=== 安装完成 ===${NC}"
echo "安装目录: $INSTALL_DIR"
echo "frps 端口: $FRPS_PORT"
echo "Agent 端口: $AGENT_PORT"
echo ""
echo "常用命令:"
echo "  查看 frps 状态: systemctl status frps"
echo "  查看 Agent 状态: systemctl status jumpfrp-agent"
echo "  查看日志: journalctl -u frps -f"
echo ""
echo -e "${YELLOW}请前往管理后台确认节点已上线${NC}"
