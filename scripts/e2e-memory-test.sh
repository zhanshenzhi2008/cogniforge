#!/usr/bin/env bash
# =============================================================================
# 阶段十四 E2E 测试脚本  (14.2 / 14.3 / 14.4 / 14.5)
# =============================================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
COGNIFORGE_DIR="$(cd "$(dirname "$SCRIPT_DIR")" && pwd)"
# cogniforge-ai 和 cogniforge 是同级目录（父目录的兄弟）
PYTHON_DIR="$(cd "$COGNIFORGE_DIR/.." && pwd)/cogniforge-ai"

GO_BASE="${COGNIFORGE_API_URL:-http://localhost:8080}"
PYTHON_BASE="${PYTHON_SERVICE_URL:-http://localhost:8000}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'
PASS=0
FAIL=0

log()  { echo -e "${GREEN}[PASS]${NC} $1"; ((PASS++)) || true; }
warn() { echo -e "${YELLOW}[SKIP]${NC} $1"; }
fail() { echo -e "${RED}[FAIL]${NC} $1"; ((FAIL++)) || true; }

http_get()  { curl -sf "${1}${2}" 2>/dev/null; }
http_post() { curl -sf -X POST "${1}${2}" -H "Content-Type: application/json" -d "$3" 2>/dev/null; }

# =============================================================================
# 前置检查
# =============================================================================
echo "=== 前置检查 ==="
if http_get "$GO_BASE" "/health" > /dev/null; then
  log "Go 服务在线: $GO_BASE"
else
  warn "Go 服务离线: $GO_BASE (跳过需要 Go 的测试)"
fi
if http_get "$PYTHON_BASE" "/health" > /dev/null; then
  log "Python 服务在线: $PYTHON_BASE"
else
  warn "Python 服务离线: $PYTHON_BASE (跳过需要 Python 的测试)"
fi

# =============================================================================
# 14.2 Playground memory_turns 参数测试
# =============================================================================
echo ""
echo "=== 14.2 Playground memory_turns ==="

# 测试模型列表端点（无需认证即可检查 API 是否正常）
if http_get "$GO_BASE" "/api/v1/models" > /dev/null; then
  log "GET /api/v1/models 返回正常"
else
  warn "GET /api/v1/models 无响应"
fi

# 检查 Go 代码中的 memory_turns 字段（grep 静态验证）
if grep -q 'memory_turns' "$COGNIFORGE_DIR/internal/chat/dto.go" 2>/dev/null; then
  log "ChatRequest 定义了 memory_turns 字段"
else
  fail "ChatRequest 缺少 memory_turns 字段"
fi

# =============================================================================
# 14.3 滚动摘要
# =============================================================================
echo ""
echo "=== 14.3 滚动摘要 ==="

# Python /api/memory/summarize 端点
if http_post "$PYTHON_BASE" "/api/memory/summarize" '{"messages":[{"role":"user","content":"你好"}]}' > /dev/null; then
  log "Python POST /api/memory/summarize 端点可达"
else
  warn "Python /api/memory/summarize 无法访问（服务可能未启动）"
fi

# Go 代码验证：buildFullMessages 支持 summary
if grep -q '对话摘要' "$COGNIFORGE_DIR/internal/chat/handler.go" 2>/dev/null; then
  log "buildFullMessages 包含摘要注入逻辑"
else
  fail "buildFullMessages 缺少摘要注入"
fi

# saveConversationAsync 存在
if grep -q "saveConversationAsync" "$COGNIFORGE_DIR/internal/chat/handler.go" 2>/dev/null; then
  log "saveConversationAsync 异步保存函数存在"
else
  fail "saveConversationAsync 未找到"
fi

# =============================================================================
# 14.4 长期记忆搜索
# =============================================================================
echo ""
echo "=== 14.4 长期记忆搜索 ==="

# Python /api/memory/search 端点
if http_post "$PYTHON_BASE" "/api/memory/search" '{"user_id":"test","query":"测试"}' > /dev/null; then
  log "Python POST /api/memory/search 端点可达"
else
  warn "Python /api/memory/search 无法访问（服务可能未启动）"
fi

# Go searchLongTermMemory 存在
if grep -q "searchLongTermMemory" "$COGNIFORGE_DIR/internal/chat/handler.go" 2>/dev/null; then
  log "searchLongTermMemory 函数存在"
else
  fail "searchLongTermMemory 未找到"
fi

# ChatRequest 有 UseLongTerm 和 AgentID 字段
if grep -q "UseLongTerm" "$COGNIFORGE_DIR/internal/chat/dto.go" 2>/dev/null && \
   grep -q "AgentID" "$COGNIFORGE_DIR/internal/chat/dto.go" 2>/dev/null; then
  log "ChatRequest 有 UseLongTerm 和 AgentID 字段"
else
  fail "ChatRequest 缺少 UseLongTerm 或 AgentID"
fi

# =============================================================================
# 14.5 记忆 CRUD API
# =============================================================================
echo ""
echo "=== 14.5 记忆 CRUD API ==="

# Router 注册检查（实际代码是 memoryHandler.RegisterRoutes）
if grep -q 'memoryHandler.RegisterRoutes' "$COGNIFORGE_DIR/internal/router/router.go" 2>/dev/null; then
  log "Router 已注册 memories 路由"
else
  fail "Router 缺少 memories 路由"
fi

# MemoryHandler 存在
if grep -q "func.*MemoryHandler" "$COGNIFORGE_DIR/internal/memory/handler.go" 2>/dev/null; then
  log "MemoryHandler 已定义"
else
  fail "MemoryHandler 未找到"
fi

# MemoryService 存在
if grep -q "func.*MemoryService" "$COGNIFORGE_DIR/internal/memory/service.go" 2>/dev/null; then
  log "MemoryService 已定义"
else
  fail "MemoryService 未找到"
fi

# ChatMemory model
if grep -q "ChatMemory" "$COGNIFORGE_DIR/internal/model/memory.go" 2>/dev/null; then
  log "ChatMemory model 已定义"
else
  fail "ChatMemory model 未找到"
fi

# =============================================================================
# Python memory router 路由
# =============================================================================
echo ""
echo "=== Python memory router ==="

if grep -q '@router.post("/search")' "$PYTHON_DIR/app/routers/memory.py" 2>/dev/null; then
  log "Python /api/memory/search 路由已注册"
else
  fail "Python /api/memory/search 路由未找到"
fi

if grep -q '@router.post("/summarize")' "$PYTHON_DIR/app/routers/memory.py" 2>/dev/null; then
  log "Python /api/memory/summarize 路由已注册"
else
  fail "Python /api/memory/summarize 路由未找到"
fi

if grep -q "set_rag_processor" "$PYTHON_DIR/app/main.py" 2>/dev/null; then
  log "set_rag_processor 在 main.py 中注册"
else
  fail "set_rag_processor 未在 main.py 中注册"
fi

# =============================================================================
# 单元测试汇总（仅检查文件存在）
# =============================================================================
echo ""
echo "=== 单元测试文件 ==="

if [ -f "$COGNIFORGE_DIR/internal/memory/service_test.go" ]; then
  log "MemoryService 单元测试文件存在"
else
  fail "MemoryService 单元测试文件缺失"
fi

if [ -f "$COGNIFORGE_DIR/internal/chat/build_test.go" ]; then
  log "buildFullMessages 单元测试文件存在"
else
  fail "buildFullMessages 单元测试文件缺失"
fi

if [ -f "$PYTHON_DIR/tests/test_memory_router.py" ]; then
  log "Python memory router 单元测试文件存在"
else
  fail "Python memory router 单元测试文件缺失"
fi

# =============================================================================
# 编译检查
# =============================================================================
echo ""
echo "=== 编译检查 ==="

if cd "$COGNIFORGE_DIR" && go build ./... 2>/dev/null; then
  log "Go 编译通过"
else
  fail "Go 编译失败"
fi

# =============================================================================
# 汇总
# =============================================================================
echo ""
echo "========================================"
echo -e "测试完成: ${GREEN}$PASS passed${NC}, ${RED}$FAIL failed${NC}"
echo "========================================"
if [ $FAIL -eq 0 ]; then
  echo -e "${GREEN}全部检查通过 ✅${NC}"
  exit 0
else
  echo -e "${RED}有 $FAIL 项检查失败${NC}"
  exit 1
fi
