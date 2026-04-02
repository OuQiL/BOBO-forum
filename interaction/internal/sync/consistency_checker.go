package sync

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"sync"
	"time"

	"interaction/internal/cache"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ConsistencyCheckerConfig struct {
	CheckInterval   time.Duration // 校验间隔
	BatchSize       int           // 每次校验数量
	AutoFix         bool          // 是否自动修复
	EnableAlarm     bool          // 是否告警
	AlarmThreshold  int           // 告警阈值
}

type LikeConsistencyChecker struct {
	redis       *redis.Redis
	db          sqlx.SqlConn
	likeCache   *cache.LikeCache
	config      ConsistencyCheckerConfig
	stopChan    chan struct{}
	wg          sync.WaitGroup
	checkQueue  chan *CheckTask
	alarmFunc   func(string, map[string]interface{})
}

type CheckTask struct {
	TargetType int32
	TargetId   int64
	Priority   int // 1-高 2-中 3-低
}

func NewLikeConsistencyChecker(
	redis *redis.Redis, 
	db sqlx.SqlConn, 
	likeCache *cache.LikeCache,
	config ConsistencyCheckerConfig,
) *LikeConsistencyChecker {
	checker := &LikeConsistencyChecker{
		redis:      redis,
		db:         db,
		likeCache:  likeCache,
		config:     config,
		stopChan:   make(chan struct{}),
		checkQueue: make(chan *CheckTask, 1000),
	}
	
	return checker
}

func (c *LikeConsistencyChecker) SetAlarmFunc(alarmFunc func(string, map[string]interface{})) {
	c.alarmFunc = alarmFunc
}

func (c *LikeConsistencyChecker) Start() {
	c.wg.Add(2)
	
	// 1. 启动定时校验
	go func() {
		defer c.wg.Done()
		ticker := time.NewTicker(c.config.CheckInterval)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				c.periodicCheck()
			case <-c.stopChan:
				logx.Info("Periodic checker stopping...")
				return
			}
		}
	}()
	
	// 2. 启动校验队列消费者
	go func() {
		defer c.wg.Done()
		for {
			select {
			case task := <-c.checkQueue:
				c.checkSingleTask(task)
			case <-c.stopChan:
				logx.Info("Check queue consumer stopping...")
				return
			}
		}
	}()
	
	logx.Infof("LikeConsistencyChecker started with interval %v", c.config.CheckInterval)
}

func (c *LikeConsistencyChecker) Stop() {
	close(c.stopChan)
	c.wg.Wait()
	logx.Info("LikeConsistencyChecker stopped")
}

func (c *LikeConsistencyChecker) AddCheckTask(task *CheckTask) {
	select {
	case c.checkQueue <- task:
		logx.Debugf("Added check task: type=%d, id=%d", task.TargetType, task.TargetId)
	default:
		logx.Warnf("Check queue is full, dropping task: type=%d, id=%d", task.TargetType, task.TargetId)
	}
}

func (c *LikeConsistencyChecker) periodicCheck() {
	ctx := context.Background()
	
	// 1. 随机抽查 100 个目标
	targets, err := c.getRandomTargets(ctx, c.config.BatchSize)
	if err != nil {
		logx.Errorf("Failed to get random targets: %v", err)
		return
	}
	
	logx.Infof("Starting periodic check for %d targets", len(targets))
	
	inconsistentCount := 0
	for _, target := range targets {
		result, err := c.checkSingleTarget(ctx, target.TargetType, target.TargetId)
		if err != nil {
			logx.Errorf("Check target failed: type=%d, id=%d, err=%v", 
				target.TargetType, target.TargetId, err)
			continue
		}
		
		if !result.IsConsistent {
			inconsistentCount++
			
			if c.config.AutoFix {
				c.fixInconsistency(ctx, target.TargetType, target.TargetId, result)
			}
		}
	}
	
	logx.Infof("Periodic check completed: total=%d, inconsistent=%d", len(targets), inconsistentCount)
	
	// 3. 告警
	if inconsistentCount > c.config.AlarmThreshold && c.config.EnableAlarm && c.alarmFunc != nil {
		c.alarmFunc("like_consistency_check", map[string]interface{}{
			"inconsistent_count": inconsistentCount,
			"threshold":          c.config.AlarmThreshold,
			"check_time":         time.Now(),
		})
	}
}

func (c *LikeConsistencyChecker) checkSingleTask(task *CheckTask) {
	ctx := context.Background()
	logx.Infof("Processing check task: type=%d, id=%d, priority=%d", 
		task.TargetType, task.TargetId, task.Priority)
	
	result, err := c.checkSingleTarget(ctx, task.TargetType, task.TargetId)
	if err != nil {
		logx.Errorf("Check task failed: type=%d, id=%d, err=%v", 
			task.TargetType, task.TargetId, err)
		return
	}
	
	if !result.IsConsistent && c.config.AutoFix {
		c.fixInconsistency(ctx, task.TargetType, task.TargetId, result)
	}
}

type CheckResult struct {
	IsConsistent   bool
	MySQLCount     int64
	RedisCount     int64
	RelationCount  int64
}

func (c *LikeConsistencyChecker) checkSingleTarget(ctx context.Context, targetType int32, targetId int64) (*CheckResult, error) {
	result := &CheckResult{}
	
	// 1. 获取 MySQL 计数（从 likes 表统计）
	mysqlCount, err := c.getMySQLCount(ctx, targetType, targetId)
	if err != nil {
		return nil, fmt.Errorf("get mysql count failed: %w", err)
	}
	result.MySQLCount = mysqlCount
	
	// 2. 获取 Redis 计数
	redisCount, err := c.likeCache.GetLikeCount(ctx, targetType, targetId)
	if err != nil {
		return nil, fmt.Errorf("get redis count failed: %w", err)
	}
	result.RedisCount = redisCount
	
	// 3. 获取 Redis 关系计数（验证）
	relationCount, err := c.likeCache.GetLikeRelationCount(ctx, targetType, targetId)
	if err != nil {
		return nil, fmt.Errorf("get relation count failed: %w", err)
	}
	result.RelationCount = relationCount
	
	// 4. 判断一致性
	if mysqlCount == redisCount && redisCount == relationCount {
		result.IsConsistent = true
		logx.Debugf("Target consistent: type=%d, id=%d, count=%d", targetType, targetId, mysqlCount)
	} else {
		result.IsConsistent = false
		logx.Warnf("Target inconsistent: type=%d, id=%d, mysql=%d, redis=%d, relation=%d", 
			targetType, targetId, mysqlCount, redisCount, relationCount)
	}
	
	return result, nil
}

func (c *LikeConsistencyChecker) getMySQLCount(ctx context.Context, targetType int32, targetId int64) (int64, error) {
	var tableName string
	if targetType == 1 {
		tableName = "likes"
	} else if targetType == 2 {
		tableName = "likes"
	} else {
		return 0, fmt.Errorf("unknown target type: %d", targetType)
	}
	
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE target_id = ? AND type = ? AND status = 1", tableName)
	var count int64
	err := c.db.QueryRowCtx(ctx, &count, query, targetId, targetType)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	
	return count, nil
}

func (c *LikeConsistencyChecker) getRandomTargets(ctx context.Context, limit int) ([]*CheckTask, error) {
	var tasks []*CheckTask
	
	// 从 posts 和 comments 中随机抽取
	query := `
		SELECT 1 as target_type, id FROM posts ORDER BY RAND() LIMIT ?
		UNION ALL
		SELECT 2 as target_type, id FROM comments ORDER BY RAND() LIMIT ?
	`
	
	rows := []struct {
		TargetType int32 `db:"target_type"`
		TargetId   int64 `db:"id"`
	}{}
	
	err := c.db.QueryRowsCtx(ctx, &rows, query, limit/2, limit/2)
	if err != nil {
		return nil, err
	}
	
	for _, row := range rows {
		tasks = append(tasks, &CheckTask{
			TargetType: row.TargetType,
			TargetId:   row.TargetId,
			Priority:   2,
		})
	}
	
	return tasks, nil
}

func (c *LikeConsistencyChecker) fixInconsistency(ctx context.Context, targetType int32, targetId int64, result *CheckResult) {
	logx.Warnf("Fixing inconsistency for target type=%d, id=%d: mysql=%d, redis=%d, relation=%d", 
		targetType, targetId, result.MySQLCount, result.RedisCount, result.RelationCount)
	
	// 1. 以 Redis Relation Count 为准（最可靠）
	standardCount := result.RelationCount
	
	// 2. 修复 Redis Count
	if result.RedisCount != standardCount {
		err := c.likeCache.SetLikeCount(ctx, targetType, targetId, standardCount)
		if err != nil {
			logx.Errorf("Failed to fix redis count: %v", err)
		} else {
			logx.Infof("Fixed redis count for target type=%d, id=%d to %d", targetType, targetId, standardCount)
		}
	}
	
	// 3. 修复 MySQL Count
	var tableName string
	if targetType == 1 {
		tableName = "posts"
	} else if targetType == 2 {
		tableName = "comments"
	} else {
		logx.Errorf("Unknown target type: %d", targetType)
		return
	}
	
	query := fmt.Sprintf("UPDATE %s SET like_count = ? WHERE id = ?", tableName)
	_, err := c.db.ExecCtx(ctx, query, standardCount, targetId)
	if err != nil {
		logx.Errorf("Failed to fix mysql count: %v", err)
	} else {
		logx.Infof("Fixed mysql count for %s %d to %d", tableName, targetId, standardCount)
	}
}

func (c *LikeConsistencyChecker) ForceCheckAndFix(ctx context.Context, targetType int32, targetId int64) error {
	result, err := c.checkSingleTarget(ctx, targetType, targetId)
	if err != nil {
		return err
	}
	
	if !result.IsConsistent {
		c.fixInconsistency(ctx, targetType, targetId, result)
	}
	
	return nil
}
