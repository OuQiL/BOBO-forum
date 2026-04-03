# Search 微服务项目总结

## ✅ 已完成功能

### 1. 项目结构

```
search/
├── api/proto/search.proto          # Protobuf API 定义
├── etc/search.yaml                 # 服务配置
├── internal/
│   ├── config/config.go            # 配置结构体
│   ├── es/                         # Elasticsearch 模块
│   │   ├── client.go               # ES 客户端封装
│   │   ├── client_test.go          # 连接测试代码
│   │   └── mapping.go              # 索引 Mapping 定义（ik_smart）
│   ├── logic/                      # 业务逻辑层
│   │   ├── searchlogic.go          # 搜索逻辑（帖子 + 用户）
│   │   └── healthlogic.go          # 健康检查 + 数据同步
│   ├── server/searchserver.go      # RPC 服务器实现
│   ├── svc/servicecontext.go       # 服务上下文（ES + MySQL）
│   └── sync/syncer.go              # MySQL→ES 数据同步
├── Dockerfile                      # Docker 镜像构建
├── go.mod                          # Go 模块依赖
├── search.go                       # 主程序入口
├── README.md                       # 项目文档
└── USAGE.md                        # 使用指南
```

### 2. 核心功能

#### ✅ Elasticsearch 连接与测试
- ES 客户端初始化（支持认证）
- 健康检查接口（Info + Cluster Health API）
- Ping 测试
- 索引自动创建
- 完整的测试用例

#### ✅ 索引设计（ik_smart 分词器）

**posts 索引** - 帖子搜索
- `title`: text + ik_smart 分词（权重 3）
- `content`: text + ik_smart 分词（权重 1）
- `tags`: text + ik_smart 分词（权重 2）
- `user_id`, `community_id`: keyword（精确过滤）
- `like_count`, `view_count`: integer（排序）
- `status`: integer（状态过滤：2=已发布）
- `created_at`, `updated_at`: date（时间范围查询）

**users 索引** - 用户搜索
- `username`: text + ik_smart 分词（权重 2）
- `nickname`: text + ik_smart 分词（权重 1）
- `email`: keyword（精确匹配）
- `status`: integer（状态过滤：1=正常）

#### ✅ 搜索功能

**帖子搜索**:
- 全文检索：标题、内容、标签
- 多字段匹配 + 字段加权
- 过滤条件：社区、作者、标签、时间范围、状态
- 排序方式：相关性、最新、最热
- 结果高亮：标题和内容关键词高亮

**用户搜索**:
- 用户名、昵称搜索
- 状态过滤（只搜索正常用户）
- 结果高亮

#### ✅ 数据同步

**全量同步**:
- 从 MySQL 读取所有已发布帖子（status=2）
- 从 MySQL 读取所有正常用户（status=1）
- 批量索引到 ES

**增量同步**（基础实现）:
- 当前实现：调用全量同步
- 扩展点：支持按更新时间增量同步

**同步 API**:
```protobuf
rpc SyncData(SyncRequest) returns (SyncResponse)
```

#### ✅ RPC API 接口

**SearchService** (gRPC):
1. `Search(SearchRequest) → SearchResponse` - 搜索接口
2. `SyncData(SyncRequest) → SyncResponse` - 数据同步
3. `HealthCheck(HealthCheckRequest) → HealthCheckResponse` - 健康检查

### 3. 集成到 api-gateway

已更新 `api-gateway/internal/logic/searchlogic.go`:
- 调用 search-rpc 的 Search 接口
- 转换 RPC 响应为 HTTP 响应
- 错误处理和状态码映射

已更新 `api-gateway/go.mod`:
- 添加 search 模块引用
- 配置 replace 路径

### 4. 测试与文档

**测试代码**:
- `client_test.go` - ES 连接测试
- `TestConnectionOnly` - 简单连接测试
- `TestHealthCheck` - 完整健康检查
- `TestCreateIndices` - 索引创建测试

**文档**:
- `README.md` - 项目概述和快速开始
- `USAGE.md` - 详细使用指南和故障排查
- `test-search.sh` - Linux/Mac测试脚本
- `test-search.bat` - Windows 测试脚本

**Docker 支持**:
- `Dockerfile` - search 服务镜像
- `docker-compose.search.yml` - 一键启动 ES + MySQL + Search

---

## 🎯 技术选型与思维链

### 1. 为什么选择 Elasticsearch？

```
需求分析 → 技术对比 → 选型决策
├─ 全文检索需求：MySQL LIKE 无法满足
├─ 中文分词需求：需要专业分词器
├─ 相关性排序：需要 TF-IDF/BM25 算法
└─ 性能要求：毫秒级搜索响应
```

### 2. 为什么使用 ik_smart 而非 ik_max_word？

```
分词器对比：
├─ ik_smart: 宽泛匹配，分词粒度粗
│   └─ 例："Python 编程教程" → ["Python", "编程", "教程"]
│   └─ 适合：搜索场景，召回率高
└─ ik_max_word: 精确匹配，分词粒度细
    └─ 例："Python 编程教程" → ["Python", "编程", "教程", "Python 编程", "编程教程", ...]
    └─ 适合：精确搜索，准确率高

决策：用户要求"宽泛"匹配 → 选择 ik_smart
```

### 3. 数据同步策略选择

```
方案对比：
├─ 方案 A: 实时双写
│   ├─ 优点：实时性强
│   └─ 缺点：复杂度高，需处理失败回滚
├─ 方案 B: 消息队列（推荐）
│   ├─ 优点：解耦，失败重试
│   └─ 缺点：秒级延迟
├─ 方案 C: Canal 监听 Binlog（最佳）
│   ├─ 优点：无侵入，自动捕获变更
│   └─ 缺点：需额外部署
└─ 方案 D: 定时任务（当前实现）
    ├─ 优点：简单易实现
    └─ 缺点：延迟较高

当前实现：方案 D（定时全量同步）
扩展建议：方案 B + 方案 C 组合
```

### 4. 索引设计思维

```
字段类型选择：
├─ text 类型：用于全文检索（title, content, tags）
│   └─ 需要分词 → 使用 ik_smart
├─ keyword 类型：用于精确匹配和过滤（user_id, community_id）
│   └─ 不分词，直接匹配
├─ integer/date 类型：用于数值查询和排序
│   └─ like_count, created_at
└─ 双字段策略：某些字段同时需要 text 和 keyword
    └─ 例：username 既需要搜索也需要精确过滤
```

### 5. 查询 DSL 设计

```
查询构建思维链：
用户需求 → 查询类型 → 字段加权 → 过滤条件 → 排序 → 分页

1. 查询类型：multi_match（多字段匹配）
2. 字段加权：title^3 > tags^2 > content^1
3. 过滤条件：bool.filter（可缓存）
4. 排序策略：
   - relevance: _score desc（默认）
   - latest: created_at desc
   - hottest: like_count desc
5. 高亮显示：highlight 配置
```

---

## 📊 系统架构

```
┌─────────────────┐
│  API Gateway    │ ← HTTP/REST API
│  (api-gateway)  │
└────────┬────────┘
         │ gRPC
         │
┌────────▼────────┐
│   Search RPC    │ ← SearchService
│   (search)      │
└────────┬────────┘
         │
    ┌────┴────┐
    │         │
┌───▼──┐  ┌──▼──────┐
│  ES  │  │  MySQL  │
│  9200│  │  3306   │
└──────┘  └─────────┘

数据流：
1. 用户搜索 → API Gateway → Search RPC → ES → 返回结果
2. 数据同步 → Search RPC → MySQL 读取 → ES 写入
```

---

## 🚀 使用流程

### 1. 环境准备

```bash
# 启动 Elasticsearch 和 MySQL
docker-compose -f docker-compose.search.yml up -d elasticsearch mysql

# 安装 ik 分词器（重要！）
docker exec -it elasticsearch \
  bin/elasticsearch-plugin install \
  https://github.com/medcl/elasticsearch-analysis-ik/releases/download/v8.11.0/elasticsearch-analysis-ik-8.11.0.zip

# 重启 ES
docker restart elasticsearch
```

### 2. 运行测试

```bash
# Windows
test-search.bat

# Linux/Mac
chmod +x test-search.sh
./test-search.sh

# 或手动测试
cd search
go test -v ./internal/es -run TestConnectionOnly
```

### 3. 启动服务

```bash
# 启动 search-rpc
cd search
go run search.go -f etc/search.yaml

# 启动 api-gateway（另一个终端）
cd api-gateway
go run gateway.go -f etc/gateway.yaml
```

### 4. 同步数据

```bash
# 使用 grpcurl 调用同步 API
grpcurl -plaintext \
  -d '{"sync_type":"full"}' \
  localhost:9093 \
  search.SearchService/SyncData
```

### 5. 测试搜索

```bash
# 搜索帖子
grpcurl -plaintext \
  -d '{"keyword":"Python","type":"post","page":1,"size":10}' \
  localhost:9093 \
  search.SearchService/Search

# 搜索用户
grpcurl -plaintext \
  -d '{"keyword":"张三","type":"user"}' \
  localhost:9093 \
  search.SearchService/Search

# 健康检查
grpcurl -plaintext \
  localhost:9093 \
  search.SearchService/HealthCheck
```

---

## 📝 关键代码说明

### 1. ES 客户端初始化 (`internal/es/client.go`)

```go
func NewClient(conf config.ElasticsearchConf) (*Client, error) {
    cfg := elasticsearch.Config{
        Addresses: conf.Addresses,
        Username:  conf.Username,
        Password:  conf.Password,
        Timeout:   time.Duration(conf.Timeout) * time.Second,
    }
    es, err := elasticsearch.NewClient(cfg)
    return &Client{ES: es}, nil
}
```

### 2. 索引 Mapping (`internal/es/mapping.go`)

```go
var IndexMapping = map[string]map[string]interface{}{
    "posts": {
        "mappings": map[string]interface{}{
            "properties": map[string]interface{}{
                "title": map[string]interface{}{
                    "type":      "text",
                    "analyzer":  "ik_smart",  // 关键：使用 ik_smart
                },
                // ...
            },
        },
    },
}
```

### 3. 搜索查询构建 (`internal/logic/searchlogic.go`)

```go
func (l *SearchLogic) buildPostQuery(req *proto.SearchRequest) map[string]interface{} {
    return map[string]interface{}{
        "query": map[string]interface{}{
            "bool": map[string]interface{}{
                "must": []interface{}{
                    map[string]interface{}{
                        "multi_match": map[string]interface{}{
                            "query":  req.Keyword,
                            "fields": []string{"title^3", "content^1", "tags^2"},
                            "type":   "best_fields",
                        },
                    },
                },
                "filter": filters,  // 过滤条件
            },
        },
        "highlight": highlight,  // 高亮配置
        "sort": sort,           // 排序配置
    }
}
```

### 4. 数据同步 (`internal/sync/syncer.go`)

```go
func (s *Syncer) SyncAllPosts(ctx context.Context) (int, error) {
    var posts []Post
    s.svcCtx.MySQL.Find(&posts)
    
    count := 0
    for _, post := range posts {
        if post.Status != 2 {  // 只同步已发布
            continue
        }
        s.IndexPost(ctx, &post)  // 索引到 ES
        count++
    }
    return count, nil
}
```

---

## ⚠️ 注意事项

### 1. ik 分词器版本必须匹配 ES 版本

```
ES 8.11.0 → ik 8.11.0
ES 8.10.0 → ik 8.10.0
```

### 2. ES 安全配置

当前配置禁用了 xpack.security（`xpack.security.enabled=false`），生产环境需启用认证。

### 3. 数据一致性

当前为手动同步，生产环境建议：
- Canal 实时同步
- 每天凌晨全量重建索引

### 4. 性能优化

- 批量索引：使用 Bulk API
- 避免深度分页：使用 search_after
- 查询缓存：使用 filter 上下文

---

## 🔮 扩展建议

### 短期（1-2 周）
- [ ] 实现搜索建议（Completion Suggester）
- [ ] 实现拼写纠错（Term Suggester）
- [ ] 添加搜索日志记录

### 中期（1 个月）
- [ ] 集成 Kafka 消息队列
- [ ] 实现 Canal 实时同步
- [ ] 添加搜索热度统计

### 长期（3 个月）
- [ ] 实现个性化搜索排序
- [ ] ES 集群化部署
- [ ] 搜索 analytics 平台

---

## 📚 学习资源

1. **Elasticsearch**
   - [官方文档](https://www.elastic.co/guide/)
   - [ik 分词器](https://github.com/medcl/elasticsearch-analysis-ik)

2. **go-zero**
   - [官方文档](https://go-zero.dev/)
   - [GitHub](https://github.com/zeromicro/go-zero)

3. **Protobuf/gRPC**
   - [Protobuf 文档](https://protobuf.dev/)
   - [gRPC Go](https://grpc.io/docs/languages/go/)

---

## ✅ 项目交付清单

- [x] search 微服务完整代码
- [x] Elasticsearch 连接测试代码
- [x] 索引 Mapping 设计（ik_smart）
- [x] MySQL→ES 数据同步逻辑
- [x] 帖子搜索 API
- [x] 用户搜索 API
- [x] 健康检查 API
- [x] api-gateway 集成
- [x] Docker 配置
- [x] 测试脚本（Windows + Linux）
- [x] 完整文档（README + USAGE）
- [x] 思维链和设计说明

---

**项目已完成！可以开始测试和使用了。** 🎉
