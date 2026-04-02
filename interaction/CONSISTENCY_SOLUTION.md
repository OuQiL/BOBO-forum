# 点赞数据一致性保障方案

## 📊 问题背景

在 Kafka 消息队列方案中，存在以下数据不一致风险：

```
用户点赞 ──▶ MySQL(likes 表)
              │
              ├──▶ Redis(计数缓存) ──▶ 实时查询
              │
              └──▶ Kafka(消息队列) ──▶ 批量同步 ──▶ MySQL(计数)
```

### ⚠️ 不一致场景

| 场景 | 原因 | 影响 | 概率 |
|------|------|------|------|
| MySQL 成功，Redis 失败 | Redis 网络故障 | 点赞记录存在，计数未增加 | 低 |
| Redis 成功，MySQL 失败 | MySQL 连接池满 | 计数虚高 | 中 |
| Kafka 消息丢失 | Broker 故障 | 计数无法同步到 MySQL | 低 |
| 消费者重复消费 | Kafka Rebalance | 计数重复增加 | 中 |
| Redis 与 MySQL 延迟 | 异步同步 | 短暂不一致（秒级） | 高 |

---

## ✅ 解决方案

### 方案一：增强版双写 + 定期校验（已实现）

#### 核心思路

1. **双写策略**
   - 写入 MySQL（likes 表）
   - 更新 Redis（计数缓存）
   - 发送 Kafka（异步同步）

2. **一致性校验**
   - 定期抽查（每 5 分钟）
   - 随机抽取 100 个目标
   - 对比 MySQL vs Redis vs Relation Count

3. **自动修复**
   - 以 Redis Relation Count 为准（Set 成员数最可靠）
   - 自动修复 Redis Count 和 MySQL Count

#### 实现代码

```go
// internal/sync/consistency_checker.go

type LikeConsistencyChecker struct {
    redis       *redis.Redis
    db          sqlx.SqlConn
    likeCache   *cache.LikeCache
    config      ConsistencyCheckerConfig
}

type ConsistencyCheckerConfig struct {
    CheckInterval   time.Duration // 校验间隔（默认 5 分钟）
    BatchSize       int           // 每次校验数量（默认 100）
    AutoFix         bool          // 是否自动修复（默认 true）
    EnableAlarm     bool          // 是否告警（默认 true）
    AlarmThreshold  int           // 告警阈值（默认 10）
}

// 启动校验器
checker := sync.NewLikeConsistencyChecker(redis, db, likeCache, ConsistencyCheckerConfig{
    CheckInterval:  5 * time.Minute,
    BatchSize:      100,
    AutoFix:        true,
    EnableAlarm:    true,
    AlarmThreshold: 10,
})
checker.Start()

// 手动触发校验（可选）
checker.AddCheckTask(&CheckTask{
    TargetType: 1,
    TargetId:   12345,
    Priority:   1, // 高优先级
})
```

#### 校验逻辑

```go
func (c *LikeConsistencyChecker) checkSingleTarget(ctx context.Context, targetType int32, targetId int64) (*CheckResult, error) {
    // 1. MySQL 计数（从 likes 表统计）
    mysqlCount := SELECT COUNT(*) FROM likes WHERE target_id = ? AND status = 1
    
    // 2. Redis 计数（String 类型）
    redisCount := GET like:count:1:12345
    
    // 3. Redis 关系计数（Set 成员数，最可靠）
    relationCount := SCARD like:relation:1:12345
    
    // 4. 判断一致性
    if mysqlCount == redisCount && redisCount == relationCount {
        return &CheckResult{IsConsistent: true}
    }
    
    // 5. 自动修复（以 relationCount 为准）
    SET like:count:1:12345 relationCount
    UPDATE posts SET like_count = relationCount WHERE id = 12345
}
```

---

### 方案二：本地消息表（更强一致性）

#### 核心思路

在 `likes` 表中增加同步状态字段，通过定时任务保证消息最终被发送到 Kafka。

```sql
ALTER TABLE likes ADD COLUMN sync_status TINYINT DEFAULT 0 COMMENT '0-待同步 1-已同步';
ALTER TABLE likes ADD COLUMN retry_count INT DEFAULT 0 COMMENT '重试次数';
```

#### 实现代码

```go
func (l *LikeLogic) Like(in *proto.LikeRequest) (*proto.LikeResponse, error) {
    err := l.svcCtx.LikeRepo.Transact(l.ctx, func(ctx context.Context, repo repository.LikeRepository) error {
        // 1. 写入 likes 表（sync_status=0）
        like := &model.Like{
            Type:       int8(in.TargetType),
            TargetId:   in.TargetId,
            UserId:     in.UserId,
            Status:     1,
            SyncStatus: 0, // 待同步
            CreatedAt:  time.Now(),
        }
        return repo.Create(ctx, like)
    })
    
    if err != nil {
        return nil, err
    }
    
    // 2. 异步发送 Kafka（失败可重试）
    go l.sendKafkaMessageWithRetry(in.TargetType, in.TargetId, in.UserId)
    
    // 3. 更新 Redis 缓存
    l.svcCtx.LikeCache.IncrLikeCount(ctx, in.TargetType, in.TargetId)
    
    return &proto.LikeResponse{Success: true}, nil
}

// 定时任务：扫描未同步记录
func (s *Syncer) ScanUnsyncedRecords() {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        // 扫描 1000 条未同步记录
        records, _ := s.db.Query("SELECT * FROM likes WHERE sync_status = 0 LIMIT 1000")
        
        for _, record := range records {
            msg := &kafka.LikeMessage{
                TargetType: record.Type,
                TargetId:   record.TargetId,
                UserId:     record.UserId,
                Action:     kafka.ActionLike,
            }
            
            err := s.kafkaProducer.SendMessage(ctx, msg)
            if err != nil {
                // 失败增加重试次数
                s.db.Exec("UPDATE likes SET retry_count = retry_count + 1 WHERE id = ?", record.Id)
                continue
            }
            
            // 成功更新状态
            s.db.Exec("UPDATE likes SET sync_status = 1 WHERE id = ?", record.Id)
        }
    }
}
```

**优点：**
- ✅ 消息不丢失（MySQL 事务保证）
- ✅ 支持失败重试
- ✅ 实现相对简单

**缺点：**
- ⚠️ 需要修改表结构
- ⚠️ 增加数据库查询压力

---

### 方案三：事务消息（RocketMQ，最强一致性）

#### 核心思路

使用 RocketMQ 的事务消息，保证本地事务和消息发送的原子性。

```go
func (l *LikeLogic) Like(in *proto.LikeRequest) (*proto.LikeResponse, error) {
    // 1. 发送半消息（Half Message）
    txMsg := l.rocketMQProducer.PrepareMessage(&kafka.LikeMessage{
        TargetType: in.TargetType,
        TargetId:   in.TargetId,
        UserId:     in.UserId,
        Action:     kafka.ActionLike,
    })
    
    // 2. 执行本地事务
    err := l.svcCtx.LikeRepo.Create(ctx, like)
    
    // 3. 根据本地事务结果提交/回滚消息
    if err == nil {
        l.rocketMQProducer.CommitMessage(txMsg)
    } else {
        l.rocketMQProducer.RollbackMessage(txMsg)
    }
    
    return &proto.LikeResponse{Success: true}, nil
}
```

**优点：**
- ✅ 最终一致性 100% 保证
- ✅ 消息不丢失、不重复

**缺点：**
- ⚠️ 需要 RocketMQ（Kafka 不支持事务消息）
- ⚠️ 实现复杂

---

## 🎯 推荐配置

### 生产环境配置

```yaml
# interaction.yaml

LikeSync:
  IntervalSeconds: 5          # Kafka 批量同步间隔
  BatchSize: 100              # 批量大小

ConsistencyCheck:
  IntervalMinutes: 5          # 一致性校验间隔
  BatchSize: 100              # 每次校验数量
  AutoFix: true               # 自动修复
  EnableAlarm: true           # 开启告警
  AlarmThreshold: 10          # 不一致数量超过 10 个告警
```

### 监控指标

```go
// Prometheus 指标
like_consistency_check_total{status="consistent"}    # 一致的数量
like_consistency_check_total{status="inconsistent"}  # 不一致的数量
like_consistency_fix_total                           # 修复次数
like_consistency_alarm_total                         # 告警次数
```

---

## 📋 实施步骤

### 1. 启用一致性校验器

```go
// cmd/server/main.go

ctx := svc.NewServiceContext(c)

// 启动 Kafka 消费者
ctx.KafkaConsumer.StartBatchConsumer(...)

// 启动一致性校验器
checker := sync.NewLikeConsistencyChecker(
    ctx.Redis,
    ctx.DB,
    ctx.LikeCache,
    sync.ConsistencyCheckerConfig{
        CheckInterval:  5 * time.Minute,
        BatchSize:      100,
        AutoFix:        true,
        EnableAlarm:    true,
        AlarmThreshold: 10,
    },
)
checker.Start()
defer checker.Stop()
```

### 2. 配置告警回调

```go
// 设置告警函数
checker.SetAlarmFunc(func(alarmType string, data map[string]interface{}) {
    // 发送钉钉/企业微信告警
    sendDingTalkAlarm(alarmType, data)
    
    // 或记录到监控系统
    metrics.Alarm(alarmType, data)
})
```

### 3. 手动触发校验（可选）

```go
// API 接口：手动触发校验
POST /api/v1/consistency/check
{
    "target_type": 1,
    "target_id": 12345
}

// 实现
func (h *Handler) CheckConsistency(w http.ResponseWriter, r *http.Request) {
    req := &CheckRequest{}
    json.NewDecoder(r.Body).Decode(req)
    
    h.svcCtx.LikeChecker.AddCheckTask(&CheckTask{
        TargetType: req.TargetType,
        TargetId:   req.TargetId,
        Priority:   1,
    })
    
    json.NewEncoder(w).Encode(map[string]bool{"success": true})
}
```

---

## 🔍 故障排查

### 场景一：Redis 与 MySQL 计数不一致

```bash
# 1. 查看 Redis 计数
redis-cli GET like:count:1:12345
redis-cli SCARD like:relation:1:12345

# 2. 查看 MySQL 计数
mysql> SELECT COUNT(*) FROM likes WHERE target_id = 12345 AND status = 1;
mysql> SELECT like_count FROM posts WHERE id = 12345;

# 3. 手动触发校验
curl -X POST http://localhost:8003/api/v1/consistency/check \
  -H "Content-Type: application/json" \
  -d '{"target_type":1,"target_id":12345}'
```

### 场景二：Kafka 消息积压

```bash
# 查看消费积压
kafka-consumer-groups.sh --bootstrap-server 127.0.0.1:9092 \
  --describe --group interaction-group

# 如果积压严重，增加消费者实例
# 或临时增加批次大小
```

### 场景三：校验器频繁告警

```go
// 1. 检查日志
tail -f interaction.log | grep "inconsistent"

// 2. 分析原因
- 网络问题？
- Redis 故障？
- Kafka 故障？

// 3. 调整配置
ConsistencyCheck:
  IntervalMinutes: 1      # 缩短校验间隔
  BatchSize: 500          # 增加校验数量
  AlarmThreshold: 5       # 降低告警阈值
```

---

## 📊 性能影响

| 操作 | 频率 | 耗时 | 影响 |
|------|------|------|------|
| 定期校验 | 每 5 分钟 | ~500ms | 低 |
| 随机抽查 | 100 个目标 | ~200ms | 低 |
| 自动修复 | 偶尔 | ~50ms/次 | 低 |
| 手动校验 | 按需 | ~10ms/次 | 低 |

---

## ✨ 最佳实践

1. **开发环境**
   - 关闭自动修复
   - 仅记录日志

2. **测试环境**
   - 开启自动修复
   - 关闭告警

3. **生产环境**
   - 开启自动修复
   - 开启告警
   - 设置合理的告警阈值

4. **大促期间**
   - 缩短校验间隔（1 分钟）
   - 增加校验数量（500 个）
   - 降低告警阈值（5 个）

---

## 📚 参考资料

- [分布式系统一致性理论](https://en.wikipedia.org/wiki/Consistency_model)
- [最终一致性模式](https://microservices.io/patterns/data/eventual-consistency.html)
- [Kafka 消息可靠性保证](https://kafka.apache.org/documentation/#design_durability)
