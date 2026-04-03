# Search 微服务使用指南

## 📋 概述

本搜索微服务基于 Elasticsearch 8.x 实现，提供帖子和用户的全文搜索功能，使用 ik_smart 分词器进行中文分词。

## 🎯 已完成的功能

### 1. 核心功能
- ✅ **Elasticsearch 连接测试** - 包含健康检查和 Ping 测试
- ✅ **索引自动创建** - 启动时自动创建 posts 和 users 索引
- ✅ **帖子搜索** - 支持标题、内容、标签的全文检索
- ✅ **用户搜索** - 支持用户名、昵称搜索
- ✅ **数据同步** - MySQL → ES 全量/增量同步
- ✅ **结果高亮** - 搜索关键词高亮显示
- ✅ **多种排序** - 相关性、最新、最热
- ✅ **过滤条件** - 社区、作者、时间范围、标签过滤

### 2. 技术特性
- ✅ 使用 ik_smart 分词器（宽泛匹配）
- ✅ go-zero RPC 框架
- ✅ GORM ORM
- ✅ Protobuf API 定义
- ✅ Docker 容器化支持

## 🚀 快速开始

### 步骤 1: 启动 Elasticsearch

```bash
# 使用 docker-compose 启动 ES 和 MySQL
docker-compose -f docker-compose.search.yml up -d elasticsearch mysql

# 查看日志
docker logs -f elasticsearch

# 验证 ES 是否启动成功
curl http://localhost:9200/_cluster/health
```

**预期输出**:
```json
{
  "cluster_name": "es-cluster",
  "status": "green",
  "number_of_nodes": 1,
  "number_of_data_nodes": 1,
  "active_primary_shards": 0,
  "active_shards": 0
}
```

### 步骤 2: 安装 ik 分词器（重要！）

Elasticsearch 8.x 默认不包含 ik 分词器，需要手动安装：

```bash
# 进入 ES 容器
docker exec -it elasticsearch bash

# 安装 ik 分词器（版本号需与 ES 版本匹配）
bin/elasticsearch-plugin install https://github.com/medcl/elasticsearch-analysis-ik/releases/download/v8.11.0/elasticsearch-analysis-ik-8.11.0.zip

# 退出并重启 ES
exit
docker restart elasticsearch

# 验证插件安装
docker exec -it elasticsearch bin/elasticsearch-plugin list
```

**预期输出**:
```
analysis-ik
```

### 步骤 3: 测试 ES 连接

```bash
cd search

# 运行连接测试
go test -v ./internal/es -run TestConnectionOnly
```

**预期输出**:
```
=== RUN   TestConnectionOnly
✓ Elasticsearch connection successful!
  Cluster: {"name":"es-cluster",...}
--- PASS: TestConnectionOnly (0.12s)
PASS
```

### 步骤 4: 创建索引

```bash
# 测试索引创建
go test -v ./internal/es -run TestCreateIndices
```

**预期输出**:
```
=== RUN   TestCreateIndices
Index posts created successfully
Index users created successfully
All indices created successfully
  Index posts: Exists ✓
  Index users: Exists ✓
--- PASS: TestCreateIndices (0.45s)
PASS
```

### 步骤 5: 同步数据到 ES

```bash
# 启动 search-rpc 服务
cd search
go run search.go -f etc/search.yaml
```

**预期输出**:
```
ES Health Check: Elasticsearch connected successfully
MySQL connected successfully
Starting search-rpc at 0.0.0.0:9093...
```

在另一个终端调用同步 API：

```bash
# 使用 grpcurl 调用 SyncData API
grpcurl -plaintext \
  -d '{"sync_type":"full"}' \
  localhost:9093 \
  search.SearchService/SyncData
```

**预期输出**:
```json
{
  "success": true,
  "message": "Data synced successfully",
  "syncedCount": "50"
}
```

### 步骤 6: 测试搜索

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
```

### 步骤 7: 健康检查

```bash
grpcurl -plaintext localhost:9093 search.SearchService/HealthCheck
```

**预期输出**:
```json
{
  "healthy": true,
  "message": "All services healthy",
  "elasticsearchConnected": true
}
```

## 📖 API 使用示例

### 1. 搜索帖子（基础）

**请求**:
```json
{
  "keyword": "Python 教程",
  "type": "post",
  "page": 1,
  "size": 10
}
```

**响应**:
```json
{
  "total": 25,
  "list": [
    {
      "type": "post",
      "post_id": "1",
      "user_id": "123",
      "username": "技术达人",
      "title": "Python 入门教程",
      "content": "本文详细介绍 Python 基础知识...",
      "like_count": 50,
      "comment_count": 10,
      "view_count": 500,
      "created_at": 1704067200,
      "highlight_title": "<em style='color: red;'>Python</em> 入门教程",
      "highlight_content": "本文详细介绍 <em style='color: red;'>Python</em> 基础知识..."
    }
  ],
  "page": 1,
  "size": 10
}
```

### 2. 搜索帖子（带过滤和排序）

**请求**:
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

**说明**:
- `community_id`: 只搜索指定社区的帖子
- `author_id`: 只搜索指定作者的帖子
- `tags`: 逗号分隔的标签列表
- `start_time/end_time`: Unix 时间戳
- `sort_by`: 
  - `relevance` - 按相关性排序（默认）
  - `latest` - 按发布时间排序
  - `hottest` - 按热度排序（点赞数）

### 3. 搜索用户

**请求**:
```json
{
  "keyword": "张三",
  "type": "user",
  "page": 1,
  "size": 10
}
```

**响应**:
```json
{
  "total": 5,
  "list": [
    {
      "type": "user",
      "user_id": "123",
      "username": "张三",
      "nickname": "张三丰",
      "email": "zhangsan@example.com",
      "avatar": "http://example.com/avatar/123.jpg",
      "gender": 1,
      "user_created_at": 1672531200,
      "highlight_title": "<em style='color: red;'>张三</em>",
      "highlight_content": "<em style='color: red;'>张三</em>丰"
    }
  ]
}
```

## 🔧 配置说明

### search.yaml 配置文件

```yaml
Name: search-rpc
ListenOn: 0.0.0.0:9093  # RPC 服务监听地址

Elasticsearch:
  Addresses:
    - http://elasticsearch:9200  # Docker 环境
    # - http://localhost:9200    # 本地环境
  Username: elastic              # ES 用户名（如启用安全）
  Password: bobo123              # ES 密码
  Timeout: 5                     # 超时时间（秒）

MySQL:
  Dsn: "root:bobo123@tcp(mysql:3306)/bobo_db?charset=utf8mb4&parseTime=true&loc=Local"
  MaxIdleConns: 10               # 最大空闲连接数
  MaxOpenConns: 50               # 最大打开连接数

Log:
  Mode: console                  # 日志模式：console/file
  Level: info                    # 日志级别：debug/info/warn/error
```

## 🧪 测试命令

### 单元测试

```bash
# 测试 ES 连接
go test -v ./internal/es -run TestConnectionOnly

# 测试健康检查
go test -v ./internal/es -run TestHealthCheck

# 测试索引创建
go test -v ./internal/es -run TestCreateIndices

# 所有测试
go test -v ./...
```

### 集成测试

```bash
# 使用 docker-compose 启动所有服务
docker-compose -f docker-compose.search.yml up -d

# 等待服务启动后，测试搜索功能
grpcurl -plaintext -d '{"keyword":"test"}' localhost:9093 search.SearchService/Search
```

## 🐛 故障排查

### 问题 1: ES 连接失败

**错误信息**:
```
Failed to create elasticsearch client: connection refused
```

**解决方案**:
```bash
# 1. 检查 ES 是否运行
docker ps | grep elasticsearch

# 2. 查看 ES 日志
docker logs elasticsearch

# 3. 测试网络连通性
curl http://localhost:9200/_cluster/health

# 4. 如果是 Docker 环境，检查网络
docker network ls
docker network inspect bobo-network
```

### 问题 2: ik 分词器未找到

**错误信息**:
```
analyzer [ik_smart] not found
```

**解决方案**:
```bash
# 1. 检查插件是否安装
docker exec -it elasticsearch bin/elasticsearch-plugin list

# 2. 如果没有 analysis-ik，重新安装
docker exec -it elasticsearch \
  bin/elasticsearch-plugin install \
  https://github.com/medcl/elasticsearch-analysis-ik/releases/download/v8.11.0/elasticsearch-analysis-ik-8.11.0.zip

# 3. 重启 ES
docker restart elasticsearch
```

### 问题 3: 索引创建失败

**错误信息**:
```
failed to create index: index_already_exists_exception
```

**解决方案**:
```bash
# 1. 删除现有索引
curl -X DELETE "localhost:9200/posts"
curl -X DELETE "localhost:9200/users"

# 2. 重新运行服务或测试
go test -v ./internal/es -run TestCreateIndices
```

### 问题 4: 数据同步失败

**错误信息**:
```
failed to query posts: table 'bobo_db.posts' doesn't exist
```

**解决方案**:
```bash
# 1. 检查 MySQL 中是否有表
docker exec -it mysql mysql -uroot -pbobo123 bobo_db -e "SHOW TABLES;"

# 2. 如果没有表，导入 SQL 文件
docker exec -i mysql mysql -uroot -pbobo123 bobo_db < ../sql/post.sql
docker exec -i mysql mysql -uroot -pbobo123 bobo_db < ../sql/user.sql

# 3. 检查数据
docker exec -it mysql mysql -uroot -pbobo123 bobo_db -e "SELECT COUNT(*) FROM posts;"
```

## 📊 性能优化建议

### 1. 批量索引优化

对于大量数据，使用 Bulk API：

```go
// 在 syncer.go 中实现批量索引
func (s *Syncer) BulkIndexPosts(ctx context.Context, posts []Post) error {
    var buf bytes.Buffer
    for _, post := range posts {
        // 构建 bulk 请求
        // ...
    }
    // 执行 bulk 请求
    // ...
}
```

### 2. 查询优化

```json
// 使用 filter 上下文（可缓存）
{
  "query": {
    "bool": {
      "filter": [
        { "term": { "status": 2 }},
        { "range": { "created_at": { "gte": "now-7d" }}}
      ]
    }
  }
}
```

### 3. 避免深度分页

```json
// 使用 search_after 替代 from/size
{
  "size": 10,
  "search_after": [1704067200],
  "sort": [{"created_at": "desc"}]
}
```

## 🔄 数据同步策略

### 当前实现：手动全量同步

```bash
# 调用 API 触发全量同步
grpcurl -plaintext -d '{"sync_type":"full"}' \
  localhost:9093 search.SearchService/SyncData
```

### 推荐方案：定时任务 + 增量同步

```go
// 在 main.go 中添加定时任务
ticker := time.NewTicker(1 * time.Minute)
go func() {
    for range ticker.C {
        // 查询最近 1 分钟更新的数据
        // 增量同步到 ES
    }
}()
```

### 最佳实践：Canal 实时同步

```
MySQL (Master) → Canal (Slave) → Kafka → Search Service → ES
```

## 📝 下一步计划

- [ ] 实现搜索建议（Completion Suggester）
- [ ] 实现拼写纠错（Term Suggester）
- [ ] 添加搜索热度统计
- [ ] 实现 Canal 实时数据同步
- [ ] 添加搜索日志和 analytics
- [ ] 实现个性化搜索排序

## 💡 常见问题

**Q: 为什么选择 ik_smart 而不是 ik_max_word？**

A: ik_smart 采用宽泛匹配，分词粒度较粗，适合搜索场景，能召回更多结果。ik_max_word 分词更细，适合精确搜索。

**Q: 搜索延迟是多少？**

A: 正常情况搜索延迟在 50-200ms，取决于数据量和查询复杂度。

**Q: 如何保证数据一致性？**

A: 当前采用手动同步，建议生产环境使用 Canal 实时同步 + 每天全量重建。

**Q: ES 集群如何扩展？**

A: 通过增加节点自动分片，建议 3 节点起步，数据量大时使用 ILM 按时间索引。

## 📚 参考资料

- [Elasticsearch 官方文档](https://www.elastic.co/guide/en/elasticsearch/reference/current/index.html)
- [ik 分词器文档](https://github.com/medcl/elasticsearch-analysis-ik)
- [go-zero 文档](https://go-zero.dev/)
- [Protobuf 文档](https://protobuf.dev/)
