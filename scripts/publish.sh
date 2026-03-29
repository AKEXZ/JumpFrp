#!/bin/bash
# JumpFrp Agent 完整发布流程
# 一键编译、提交、发布到 GitHub

set -e

echo "🚀 JumpFrp Agent 完整发布流程"
echo "================================"
echo ""

# 第 1 步：编译
echo "📦 第 1 步：编译 Agent..."
cd agent
rm -f jumpfrp-agent
go build -o jumpfrp-agent ./cmd/main.go
AGENT_SIZE=$(ls -lh jumpfrp-agent | awk '{print $5}')
echo "✅ 编译成功 ($AGENT_SIZE)"
cd ..

# 第 2 步：提交代码
echo ""
echo "📝 第 2 步：提交代码..."
git add -A
git commit -m "build: compile Agent v1.1.0" || echo "⚠️  没有新改动"
git push
echo "✅ 代码已推送"

# 第 3 步：创建 tag
echo ""
echo "🏷️  第 3 步：创建 tag..."
TAG="v1.1.0"
if git rev-parse "$TAG" >/dev/null 2>&1; then
  echo "⚠️  Tag $TAG 已存在，删除旧 tag..."
  git tag -d "$TAG"
  git push origin ":refs/tags/$TAG" 2>/dev/null || true
fi
git tag "$TAG"
git push origin "$TAG"
echo "✅ Tag 已创建并推送"

# 第 4 步：发布 Release
echo ""
echo "📤 第 4 步：发布 Release..."
echo ""
echo "需要 GitHub token 来发布 Release"
echo "获取方法："
echo "1. 访问 https://github.com/settings/tokens"
echo "2. 创建新 token (repo 权限)"
echo "3. 复制 token"
echo ""
read -p "请输入你的 GitHub token: " GITHUB_TOKEN

if [ -z "$GITHUB_TOKEN" ]; then
  echo "❌ Token 为空，取消发布"
  exit 1
fi

# 发布 Release
REPO="AKEXZ/JumpFrp"
RELEASE_NAME="v1.1.0 - Agent Config Auto-Update"

RELEASE_JSON=$(cat <<'EOF'
{
  "tag_name": "v1.1.0",
  "name": "v1.1.0 - Agent Config Auto-Update",
  "body": "## Changes\n- Agent now fetches frps config immediately on startup (2s delay)\n- Supports per-user token authentication\n- Dynamic frps config update with version control\n- Automatic frps restart when config changes\n\n## Features\n- Multi-token support in frps configuration\n- Config version tracking and auto-update\n- Immediate config fetch on Agent startup\n- Graceful frps restart\n\n## Installation\n\n```bash\nwget https://github.com/AKEXZ/JumpFrp/releases/download/v1.1.0/jumpfrp-agent\nchmod +x jumpfrp-agent\nsudo ./jumpfrp-agent\n```\n\n## Update\n\nOn your node server:\n\n```bash\ncurl -O https://raw.githubusercontent.com/AKEXZ/JumpFrp/main/scripts/update-agent.sh\nchmod +x update-agent.sh\nsudo ./update-agent.sh\n```",
  "draft": false,
  "prerelease": false
}
EOF
)

RELEASE_RESPONSE=$(curl -s -X POST \
  -H "Authorization: token $GITHUB_TOKEN" \
  -H "Accept: application/vnd.github.v3+json" \
  "https://api.github.com/repos/$REPO/releases" \
  -d "$RELEASE_JSON")

RELEASE_ID=$(echo "$RELEASE_RESPONSE" | grep -o '"id": [0-9]*' | head -1 | grep -o '[0-9]*')

if [ -z "$RELEASE_ID" ]; then
  echo "❌ 创建 Release 失败"
  echo "$RELEASE_RESPONSE"
  exit 1
fi

echo "✅ Release 已创建 (ID: $RELEASE_ID)"

# 上传二进制
echo "📤 上传二进制..."
curl -s -X POST \
  -H "Authorization: token $GITHUB_TOKEN" \
  -H "Accept: application/vnd.github.v3+json" \
  -H "Content-Type: application/octet-stream" \
  --data-binary @"agent/jumpfrp-agent" \
  "https://api.github.com/repos/$REPO/releases/$RELEASE_ID/assets?name=jumpfrp-agent" \
  > /dev/null

echo "✅ 二进制已上传"

# 完成
echo ""
echo "================================"
echo "✅ 发布完成！"
echo ""
echo "📥 下载链接:"
echo "https://github.com/$REPO/releases/download/v1.1.0/jumpfrp-agent"
echo ""
echo "🚀 在节点上一键安装:"
echo "curl -O https://raw.githubusercontent.com/AKEXZ/JumpFrp/main/scripts/update-agent.sh"
echo "chmod +x update-agent.sh"
echo "sudo ./update-agent.sh"
echo ""
