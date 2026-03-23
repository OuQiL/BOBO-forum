# BOBO Forum - API Gateway

论坛项目的API网关，统一入口部署在8080端口。

## 项目架构

本项目采用go-zero微服务架构，包含以下微服务：

- **auth** (port:8001) - 用户注册登录服务
- **post** (port:8002) - 发帖删帖服务
- **search** (port:8003) - 搜索帖子和博主服务
- **interaction** (port:8004) - 互动服务（点赞评论帖子，关注博主）
- **api-gateway** (port:8080) - API网关统一入口

## API Gateway 功能

### 已实现功能

- [x] 基础网关框架搭建
- [x] 路由注册和请求分发
- [x] 配置文件管理
- [x] 服务上下文初始化
- [x] Handler处理器
- [x] Logic业务逻辑层（Demo版本）

### API接口

#### 认证相关
- `POST /api/v1/auth/register` - 用户注册
- `POST /api/v1/auth/login` - 用户登录

#### 帖子相关
- `POST /api/v1/posts` - 创建帖子
- `DELETE /api/v1/posts/:id` - 删除帖子

#### 搜索相关
- `GET /api/v1/search` - 搜索帖子或用户

#### 互动相关
- `POST /api/v1/interaction/like` - 点赞帖子
- `POST /api/v1/interaction/comment` - 评论帖子
- `POST /api/v1/interaction/follow` - 关注用户

## TODO 待完成事项

### 高优先级

1. **JWT认证中间件**
   - 实现JWT token生成和验证
   - 在需要认证的路由上添加中间件
   - Token刷新机制

2. **RPC服务调用**
   - 实现与auth.rpc的通信
   - 实现与post.rpc的通信
   - 实现与search.rpc的通信
   - 实现与interaction.rpc的通信

3. **错误处理**
   - 统一错误码定义
   - 优雅的错误响应格式
   - 日志记录和监控

### 中优先级

4. **请求限流**
   - 基于IP的限流
   - 基于用户的限流
   - 防止DDoS攻击

5. **链路追踪**
   - 集成Jaeger或Zipkin
   - 请求链路可视化

6. **缓存优化**
   - Redis缓存热点数据
   - 缓存失效策略

7. **日志中间件**
   - 请求日志记录
   - 响应时间统计
   - 错误日志告警

### 低优先级

8. **API文档**
   - Swagger文档生成
   - 接口测试页面

9. **监控告警**
   - Prometheus指标采集
   - Grafana监控面板
   - 告警规则配置

10. **性能优化**
    - 连接池优化
    - 并发控制
    - 响应压缩

## Logic层待实现细节

### AuthLogic
- [ ] Register: 调用auth.rpc进行用户注册
- [ ] Login: 调用auth.rpc进行用户登录并生成JWT token
- [ ] Logout: 用户登出逻辑
- [ ] RefreshToken: Token刷新逻辑

### PostLogic
- [ ] CreatePost: 从context获取当前用户ID，调用post.rpc创建帖子
- [ ] DeletePost: 验证帖子所有权，调用post.rpc删除帖子
- [ ] GetPost: 获取帖子详情，支持缓存
- [ ] ListPosts: 分页获取帖子列表

### SearchLogic
- [ ] Search: 根据type区分搜索帖子或用户
- [ ] SearchPosts: 实现帖子全文搜索
- [ ] SearchUsers: 实现用户搜索
- [ ] 集成Elasticsearch

### InteractionLogic
- [ ] Like: 点赞帖子，检查幂等性，更新计数
- [ ] Unlike: 取消点赞
- [ ] Comment: 评论帖子，敏感词过滤，通知作者
- [ ] Follow: 关注用户，检查幂等性
- [ ] Unfollow: 取消关注
- [ ] GetFollowers: 获取粉丝列表
- [ ] GetFollowing: 获取关注列表

## 运行方式

```bash
# 安装依赖
go mod tidy

# 启动网关
go run gateway.go -f etc/gateway.yaml
```

## 配置说明

配置文件位于 `etc/gateway.yaml`，包含：

- 服务端口配置
- 各微服务的RPC客户端配置
- Etcd服务发现配置
- 日志配置

## 项目结构

```
api-gateway/
├── etc/
│   └── gateway.yaml          # 配置文件
├── internal/
│   ├── config/
│   │   └── config.go         # 配置结构定义
│   ├── handler/
│   │   ├── handler.go        # HTTP处理器
│   │   └── routes.go         # 路由注册
│   ├── logic/
│   │   ├── authlogic.go      # 认证业务逻辑
│   │   ├── postlogic.go      # 帖子业务逻辑
│   │   ├── searchlogic.go    # 搜索业务逻辑
│   │   └── interactionlogic.go # 互动业务逻辑
│   ├── svc/
│   │   └── servicecontext.go # 服务上下文
│   └── types/
│       └── types.go          # 类型定义
├── gateway.go                # 主入口
└── go.mod                    # Go模块文件
```
