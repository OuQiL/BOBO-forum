#!/bin/bash

# Search 微服务快速测试脚本

set -e

echo "========================================"
echo "  Search 微服务测试脚本"
echo "========================================"
echo ""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查 ES_HOST 环境变量
ES_HOST=${ES_HOST:-"localhost"}
ES_PORT=${ES_PORT:-"9200"}

echo -e "${YELLOW}Elasticsearch 地址：http://${ES_HOST}:${ES_PORT}${NC}"
echo ""

# 1. 测试 ES 连接
echo "========================================"
echo "1. 测试 Elasticsearch 连接"
echo "========================================"

if curl -s -o /dev/null -w "%{http_code}" "http://${ES_HOST}:${ES_PORT}/_cluster/health" | grep -q "200\|503"; then
    echo -e "${GREEN}✓ Elasticsearch 可达${NC}"
    
    # 获取集群信息
    echo ""
    echo "集群信息:"
    curl -s "http://${ES_HOST}:${ES_PORT}/" | head -20
    echo ""
else
    echo -e "${RED}✗ Elasticsearch 不可达${NC}"
    echo "请检查 ES 是否运行：docker ps | grep elasticsearch"
    exit 1
fi

# 2. 检查 ik 分词器
echo ""
echo "========================================"
echo "2. 检查 ik 分词器"
echo "========================================"

# 测试 ik 分词器
TEST_ANALYSIS=$(curl -s -X POST "http://${ES_HOST}:${ES_PORT}/_analyze" \
  -H 'Content-Type: application/json' \
  -d '{
    "analyzer": "ik_smart",
    "text": "Python 编程教程"
  }')

if echo "$TEST_ANALYSIS" | grep -q "tokens"; then
    echo -e "${GREEN}✓ ik_smart 分词器可用${NC}"
    echo ""
    echo "分词测试结果:"
    echo "$TEST_ANALYSIS" | python3 -m json.tool 2>/dev/null || echo "$TEST_ANALYSIS"
else
    echo -e "${RED}✗ ik_smart 分词器不可用${NC}"
    echo ""
    echo "请安装 ik 分词器:"
    echo "docker exec -it elasticsearch bin/elasticsearch-plugin install https://github.com/medcl/elasticsearch-analysis-ik/releases/download/v8.11.0/elasticsearch-analysis-ik-8.11.0.zip"
    exit 1
fi

# 3. 检查索引
echo ""
echo "========================================"
echo "3. 检查索引状态"
echo "========================================"

echo "现有索引:"
curl -s "http://${ES_HOST}:${ES_PORT}/_cat/indices?v" | grep -E "posts|users" || echo "暂无搜索索引"

# 4. 运行 Go 测试
echo ""
echo "========================================"
echo "4. 运行 Go 连接测试"
echo "========================================"

cd "$(dirname "$0")/search"

if [ -f "go.mod" ]; then
    echo "运行 ES 连接测试..."
    echo ""
    
    # 运行测试
    if go test -v ./internal/es -run TestConnectionOnly; then
        echo -e "${GREEN}✓ 连接测试通过${NC}"
    else
        echo -e "${RED}✗ 连接测试失败${NC}"
        echo "请检查 go.mod 和依赖是否正确安装"
        exit 1
    fi
    
    echo ""
    echo "运行健康检查测试..."
    echo ""
    
    if go test -v ./internal/es -run TestHealthCheck; then
        echo -e "${GREEN}✓ 健康检查测试通过${NC}"
    else
        echo -e "${YELLOW}⚠ 健康检查测试失败（可能是认证问题）${NC}"
    fi
else
    echo -e "${RED}✗ go.mod 文件不存在${NC}"
    echo "请确保在正确的目录运行脚本"
    exit 1
fi

# 5. 总结
echo ""
echo "========================================"
echo "测试总结"
echo "========================================"
echo ""
echo -e "${GREEN}✓ 所有测试完成${NC}"
echo ""
echo "下一步操作:"
echo "1. 启动 search-rpc 服务：cd search && go run search.go -f etc/search.yaml"
echo "2. 同步数据到 ES: grpcurl -plaintext -d '{\"sync_type\":\"full\"}' localhost:9093 search.SearchService/SyncData"
echo "3. 测试搜索功能：grpcurl -plaintext -d '{\"keyword\":\"Python\",\"type\":\"post\"}' localhost:9093 search.SearchService/Search"
echo ""
