#!/bin/bash
# JumpFrp 开发环境启动脚本

echo "🚀 启动 JumpFrp 开发环境..."

# 创建数据目录
mkdir -p master/data

# 启动后端
echo "📦 启动后端服务 (port 8080)..."
cd master
export PATH="/opt/homebrew/bin:$PATH"
go run cmd/server/main.go &
BACKEND_PID=$!
cd ..

# 等待后端启动
sleep 2

# 启动前端
echo "🎨 启动前端服务 (port 5173)..."
cd frontend
npm run dev &
FRONTEND_PID=$!
cd ..

echo ""
echo "✅ 启动完成！"
echo "   前台页面: http://localhost:5173"
echo "   管理后台: http://localhost:5173/admin"
echo "   API接口:  http://localhost:8080/api"
echo "   默认管理员: admin / admin123456"
echo ""
echo "按 Ctrl+C 停止所有服务"

# 等待退出信号
trap "kill $BACKEND_PID $FRONTEND_PID 2>/dev/null; echo '已停止'" INT
wait
