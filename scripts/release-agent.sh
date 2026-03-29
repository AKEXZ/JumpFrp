#!/bin/bash
# 发布 Agent 到 GitHub Releases
# 用法: ./release-agent.sh <github-token>

set -e

if [ -z "$1" ]; then
  echo "用法: $0 <github-token>"
  echo ""
  echo "获取 token:"
  echo "1. 访问 https://github.com/settings/tokens"
  echo "2. 创建新 token (repo 权限)"
  echo "3. 复制 token 并运行此脚本"
  exit 1
fi

GITHUB_TOKEN="$1"
REPO="AKEXZ/JumpFrp"
TAG="v1.1.0"
AGENT_BIN="agent/jumpfrp-agent"
RELEASE_NAME="v1.1.0 - Agent Config Auto-Update"

echo "📦 发布 Agent 到 GitHub Releases"
echo "================================"
echo "仓库: $REPO"
echo "版本: $TAG"
echo "二进制: $AGENT_BIN"
echo ""

# 检查二进制是否存在
if [ ! -f "$AGENT_BIN" ]; then
  echo "❌ 未找到 $AGENT_BIN"
  echo "请先运行: cd agent && go build -o jumpfrp-agent ./cmd/main.go"
  exit 1
fi

echo "✅ 二进制文件存在"

# 检查 tag 是否已存在
echo "🔍 检查 tag..."
if git rev-parse "$TAG" >/dev/null 2>&1; then
  echo "⚠️  Tag $TAG 已存在，删除旧 tag..."
  git tag -d "$TAG"
  git push origin ":refs/tags/$TAG" 2>/dev/null || true
fi

# 创建新 tag
echo "🏷️  创建 tag..."
git tag "$TAG"
git push origin "$TAG"

# 创建 Release
echo "📤 创建 Release..."
RELEASE_JSON=$(cat <<EOF
{
  "tag_name": "$TAG",
  "name": "$RELEASE_NAME",
  "body": "## Changes\n- Agent now fetches frps config immediately on startup (2s delay)\n- Supports per-user token authentication\n- Dynamic frps config update with version control\n- Automatic frps restart when config changes\n\n## Features\n- Multi-token support in frps configuration\n- Config version tracking and auto-update\n- Immediate config fetch on Agent startup\n- Graceful frps restart\n\n## Installation\n\`\`\`bash\nwget https://github.com/AKEXZ/JumpFrp/releases/download/v1.1.0/jumpfrp-agent\nchmod +x jumpfrp-agent\nsudo ./jumpfrp-agent\n\`\`\`\n\n## Update\n\nOn your node server:\n\n\`\`\`bash\ncurl -O https://raw.githubusercontent.com/AKEXZ/JumpFrp/main/scripts/update-agent.sh\nchmod +x update-agent.sh\nsudo ./update-agent.sh\n\`\`\`",
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
  --data-binary @"$AGENT_BIN" \
  "https://api.github.com/repos/$REPO/releases/$RELEASE_ID/assets?name=jumpfrp-agent" \
  > /dev/null

echo "✅ 二进制已上传"

echo ""
echo "================================"
echo "✅ 发布完成！"
echo ""
echo "📥 下载链接:"
echo "https://github.com/$REPO/releases/download/$TAG/jumpfrp-agent"
echo ""
echo "🚀 一键安装:"
echo "curl -O https://raw.githubusercontent.com/AKEXZ/JumpFrp/main/scripts/update-agent.sh"
echo "chmod +x update-agent.sh"
echo "sudo ./update-agent.sh"
