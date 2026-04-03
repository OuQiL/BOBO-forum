package logic

import (
	"context"
	"search/api/proto"
	"search/internal/svc"
	"search/internal/sync"
)

// HealthCheckLogic 健康检查逻辑
type HealthCheckLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewHealthCheckLogic 创建健康检查逻辑
func NewHealthCheckLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HealthCheckLogic {
	return &HealthCheckLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// HealthCheck 执行健康检查
func (l *HealthCheckLogic) HealthCheck(req *proto.HealthCheckRequest) (*proto.HealthCheckResponse, error) {
	// 检查 ES 连接
	esConnected := false
	esMessage := "Elasticsearch not connected"

	if l.svcCtx.ES != nil {
		healthy, msg, _ := l.svcCtx.ES.HealthCheck(l.ctx)
		esConnected = healthy
		esMessage = msg
	}

	// 检查 MySQL 连接
	mysqlConnected := false
	if l.svcCtx.MySQL != nil {
		db, _ := l.svcCtx.MySQL.DB()
		if err := db.Ping(); err == nil {
			mysqlConnected = true
		}
	}

	healthy := esConnected && mysqlConnected
	message := "All services healthy"
	if !healthy {
		message = esMessage
		if !mysqlConnected {
			message += "; MySQL not connected"
		}
	}

	return &proto.HealthCheckResponse{
		Healthy:               healthy,
		Message:               message,
		ElasticsearchConnected: esConnected,
	}, nil
}

// SyncLogic 数据同步逻辑
type SyncLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewSyncLogic 创建同步逻辑
func NewSyncLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncLogic {
	return &SyncLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// SyncData 同步数据
func (l *SyncLogic) SyncData(req *proto.SyncRequest) (*proto.SyncResponse, error) {
	syncer := sync.NewSyncer(l.svcCtx)

	var count int
	var err error

	switch req.SyncType {
	case "full":
		count, err = syncer.SyncAll(l.ctx)
	case "incremental":
		// 增量同步：这里简化为全量同步，实际应该根据更新时间同步
		count, err = syncer.SyncAll(l.ctx)
	default:
		// 默认全量同步
		count, err = syncer.SyncAll(l.ctx)
	}

	if err != nil {
		return &proto.SyncResponse{
			Success:     false,
			Message:     err.Error(),
			SyncedCount: int32(count),
		}, nil
	}

	return &proto.SyncResponse{
		Success:     true,
		Message:     "Data synced successfully",
		SyncedCount: int32(count),
	}, nil
}
