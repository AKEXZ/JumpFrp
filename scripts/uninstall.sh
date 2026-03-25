#!/bin/bash
# JumpFrp 节点卸载脚本

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

if [[ $EUID -ne 0 ]]; then
  echo -e "${RED}请使用 root 用户运行${NC}"
  exit 1
fi

echo -e "${GREEN}=== JumpFrp 节点卸载 ===${NC}"
read -p "确定要卸载 JumpFrp 节点吗? [y/N] " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
  exit 0
fi

systemctl stop frps jumpfrp-agent 2>/dev/null || true
systemctl disable frps jumpfrp-agent 2>/dev/null || true
rm -f /etc/systemd/system/frps.service
rm -f /etc/systemd/system/jumpfrp-agent.service
systemctl daemon-reload
rm -rf /opt/jumpfrp

echo -e "${GREEN}卸载完成${NC}"
