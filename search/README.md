# Search RPC Service

基于 Elasticsearch 的搜索微服务，提供帖子和用户的搜索功能。

## 功能特性

- ✅ 帖子全文搜索（标题、内容、标签）
- ✅ 用户搜索（用户名、昵称）
- ✅ 使用 ik_smart 分词器（宽泛匹配）
- ✅ MySQL 数据同步到 ES（全量/增量）
- ✅ 搜索结果高亮显示
- ✅ 多种排序方式（相关性、最新、最热）
- ✅ 过滤条件（社区、作者、时间范围、标签）
- ✅ 健康检查接口

## 目录结构

```
search/
├── api/
│   └── proto/
│       └── search.proto          # Proto 定义
├── etc/
│   └── search.yaml               # 配置文件
├── internal/
│   ├── config/
│   │   └── config.go             # 配置结构体
│   ├── es/
│   │   ├── client.go             # ES 客户端封装
│   │   ├── client_test.go        # 连接测试
│   │   └── mapping.go            # 索引 Mapping 定义
│   ├── logic/
│   │   ├── searchlogic.go        # 搜索逻辑
│   │   └── healthlogic.go        # 健康检查和同步逻辑
│   ├── server/
│   │   └── searchserver.go       # RPC 服务器
│   ├── svc/
│   │   └── servicecontext.go     # 服务上下文
│   └── sync/
│       └── syncer.go             # 数据同步逻辑
├── go.mod
└── search.go                     # 主程序入口
```

## 快速开始

### 1. 环境要求

- Go 1.25+
- Elasticsearch 8.x（需安装 ik 分词器）
- MySQL 5.7+

### 2. 配置 Elasticsearch

确保 ES 安装了 ik 分词器：

```bash
# 进入 ES 容器
docker exec -it <es_container_id> bash

# 检查 ik 分词器
bin/elasticsearch-plugin list

# 如果没有，安装 ik 分词器
bin/elasticsearch-plugin install https://github.com/medcl/elasticsearch-analysis-ik/releases/download/v8.x.x/elasticsearch-analysis-ik-8.x.x.zip
```

### 3. 配置文件

编辑 `etc/search.yaml`:

```yaml
Name: search-rpc
ListenOn: 0.0.0.0:9093

Elasticsearch:
  Addresses:
    - http://elasticsearch:9200  # Docker 环境
    # - http://localhost:9200    # 本地环境
  Username: elastic
  Password: bobo123
  Timeout: 5

MySQL:
  Dsn: "root:bobo123@tcp(mysql:3306)/bobo_db?charset=utf8mb4&parseTime=true&loc=Local"
  MaxIdleConns: 10
  MaxOpenConns: 50

Log:
  Mode: console
  Level: info
```

### 4. 运行测试

#### 测试 ES 连接

```bash
cd search

# 本地环境
go test -v ./internal/es -run TestConnectionOnly

# 指定 ES 地址
ES_HOST=192.168.1.100 go test -v ./internal/es -run TestConnectionOnly

# 完整健康检查
go test -v ./internal/es -run TestHealthCheck
```

#### 测试索引创建

```bash
go test -v ./internal/es -run TestCreateIndices
```

### 5. 启动服务

```bash
# 本地开发
go run search.go -f etc/search.yaml

# Docker 环境
docker run -d \
  --name search-rpc \
  -v $(pwd)/etc/search.yaml:/app/etc/search.yaml \
  search-rpc:latest
```

## API 接口

### 1. Search - 搜索接口

**请求**:
```protobuf
message SearchRequest {
  string keyword = 1;      // 搜索关键词
  string type = 2;         // "post" 或 "user"
  int32 page = 3;          // 页码，默认 1
  int32 size = 4;          // 每页数量，默认 10
  optional int64 community_id = 5;
  optional int64 author_id = 6;
  optional string tags = 7;
  optional int64 start_time = 8;
  optional int64 end_time = 9;
  string sort_by = 10;     // "relevance", "latest", "hottest"
}
```

**响应**:
```protobuf
message SearchResponse {
  int64 total = 1;
  repeated SearchResult list = 2;
  int32 page = 3;
  int32 size = 4;
}
```

**示例**:
```bash
# 搜索帖子
grpcurl -plaintext -d '{"keyword":"Python","type":"post","page":1,"size":10}' \
  localhost:9093 search.SearchService/Search

# 搜索用户
grpcurl -plaintext -d '{"keyword":"张三","type":"user"}' \
  localhost:9093 search.SearchService/Search
```

### 2. SyncData - 数据同步接口

**请求**:
```protobuf
message SyncRequest {
  string sync_type = 1;    // "full" 或 "incremental"
}
```

**示例**:
```bash
# 全量同步
grpcurl -plaintext -d '{"sync_type":"full"}' \
  localhost:9093 search.SearchService/SyncData
```

### 3. HealthCheck - 健康检查接口

**示例**:
```bash
grpcurl -plaintext localhost:9093 search.SearchService/HealthCheck
```

## 数据同步策略

### 全量同步

```bash
# 调用 SyncData API
curl -X POST http://localhost:9093/sync -d '{"sync_type":"full"}'
```

全量同步会：
1. 从 MySQL 读取所有已发布的帖子（status=2）
2. 从 MySQL 读取所有正常状态的用户（status=1）
3. 批量索引到 ES

### 增量同步（推荐方案）

实际生产中建议使用以下方案：

**方案 A: 消息队列**
```
MySQL 变更 → 发送消息 → Kafka/RabbitMQ → Search 服务消费 → 更新 ES
```

**方案 B: Canal 监听 Binlog**
```
Canal 监听 MySQL Binlog → 解析变更 → 发送到 ES
```

**方案 C: 定时任务**
```
定时任务（每分钟）→ 查询 updated_at > last_sync_time → 增量更新 ES
```

## 索引设计

### posts 索引

```json
{
  "mappings": {
    "properties": {
      "id": {"type": "keyword"},
      "user_id": {"type": "keyword"},
      "community_id": {"type": "keyword"},
      "username": {"type": "keyword"},
      "title": {
        "type": "text",
        "analyzer": "ik_smart"
      },
      "content": {
        "type": "text",
        "analyzer": "ik_smart"
      },
      "tags": {
        "type": "text",
        "analyzer": "ik_smart"
      },
      "like_count": {"type": "integer"},
      "comment_count": {"type": "integer"},
      "view_count": {"type": "integer"},
      "status": {"type": "integer"},
      "created_at": {"type": "date"},
      "updated_at": {"type": "date"}
    }
  }
}
```

### users 索引

```json
{
  "mappings": {
    "properties": {
      "id": {"type": "keyword"},
      "username": {
        "type": "text",
        "analyzer": "ik_smart"
      },
      "nickname": {
        "type": "text",
        "analyzer": "ik_smart"
      },
      "email": {"type": "keyword"},
      "avatar": {"type": "keyword"},
      "gender": {"type": "byte"},
      "status": {"type": "integer"},
      "created_at": {"type": "date"}
    }
  }
}
```

## 搜索示例

### 1. 基础搜索

```json
{
  "keyword": "Python 教程",
  "type": "post",
  "page": 1,
  "size": 10
}
```

### 2. 带过滤的搜索

```json
{
  "keyword": "Golang",
  "type": "post",
  "community_id": 1,
  "author_id": 123,
  "tags": "backend,golang",
  "start_time": 1704067200,
  "end_time": 1735689599,
  "sort_by": "hottest"
}
```

### 3. 搜索结果

```json
{
  "total": 100,
  "list": [
    {
      "type": "post",
      "post_id": 1,
      "user_id": 123,
      "username": "张三",
      "title": "Python 入门教程",
      "content": "本文介绍 Python 基础知识...",
      "like_count": 50,
      "highlight": {
        "title": "<em style='color: red;'>Python</em> 入门教程",
        "content": "本文介绍 <em style='color: red;'>Python</em> 基础知识..."
      }
    }
  ],
  "page": 1,
  "size": 10
}
```

## 故障排查

### 1. ES 连接失败

```bash
# 检查 ES 是否运行
curl http://localhost:9200/_cluster/health

# 检查网络连通性
telnet localhost 9200

# 查看 ES 日志
docker logs <es_container_id>
```

### 2. ik 分词器未安装

```
错误：analyzer [ik_smart] not found
解决：安装 ik 分词器并重启 ES
```

### 3. 索引创建失败

```bash
# 手动删除索引后重试
curl -X DELETE "localhost:9200/posts"
curl -X DELETE "localhost:9200/users"

# 重新运行测试
go test -v ./internal/es -run TestCreateIndices
```

## 性能优化

1. **批量索引**: 使用 Bulk API 批量写入数据
2. **合理分片**: 根据数据量设置分片数
3. **刷新间隔**: 调整 refresh_interval 提高写入性能
4. **查询缓存**: 使用 Filter 上下文启用缓存
5. **避免深度分页**: 使用 search_after 替代 from/size

## 监控指标

- ES 集群健康状态
- 索引/查询 QPS
- 查询延迟（P95, P99）
- 数据同步延迟
- ES 节点资源使用率

## 下一步计划

- [ ] 实现搜索建议（Suggester）
- [ ] 实现拼写纠错
- [ ] 添加搜索热度统计
- [ ] 实现个性化推荐
- [ ] 集成 Canal 实时同步

## 许可证

MIT License
