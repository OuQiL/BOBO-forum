package logic

import (
	"context"
	"testing"

	"search/api/proto"
	"search/internal/config"
	"search/internal/svc"
)

// TestHealthCheckLogic_HealthCheck 测试健康检查逻辑
func TestHealthCheckLogic_HealthCheck(t *testing.T) {
	svcCtx := &svc.ServiceContext{
		Config: config.Config{},
		Context: context.Background(),
		// ES 和 MySQL 为 nil，测试降级情况
	}

	logic := NewHealthCheckLogic(context.Background(), svcCtx)
	
	req := &proto.HealthCheckRequest{}
	
	resp, err := logic.HealthCheck(req)
	
	if err != nil {
		t.Fatalf("HealthCheck should not return error, got: %v", err)
	}

	if resp == nil {
		t.Fatal("Response should not be nil")
	}

	// 由于 ES 和 MySQL 都是 nil，应该返回不健康
	if resp.Healthy {
		t.Error("Expected healthy=false when services are not connected")
	}

	if resp.ElasticsearchConnected {
		t.Error("Expected elasticsearch_connected=false")
	}
}

// TestHealthCheckLogic_Message 测试健康检查消息
func TestHealthCheckLogic_Message(t *testing.T) {
	svcCtx := &svc.ServiceContext{
		Config:  config.Config{},
		Context: context.Background(),
	}

	logic := NewHealthCheckLogic(context.Background(), svcCtx)
	
	req := &proto.HealthCheckRequest{}
	resp, _ := logic.HealthCheck(req)

	// 验证消息不为空
	if resp.Message == "" {
		t.Error("Expected non-empty message")
	}

	t.Logf("Health check message: %s", resp.Message)
}

// TestSyncLogic_SyncData 测试数据同步逻辑
func TestSyncLogic_SyncData(t *testing.T) {
	svcCtx := &svc.ServiceContext{
		Config:  config.Config{},
		Context: context.Background(),
		// MySQL 为 nil，同步会失败
	}

	logic := NewSyncLogic(context.Background(), svcCtx)
	
	req := &proto.SyncRequest{
		SyncType: "full",
	}
	
	resp, err := logic.SyncData(req)
	
	if err != nil {
		// 预期会有错误，因为 MySQL 未连接
		t.Logf("SyncData returned error (expected): %v", err)
	}

	if resp == nil {
		t.Fatal("Response should not be nil")
	}

	// 验证响应字段存在
	t.Logf("Sync response: success=%v, message=%s, count=%d", 
		resp.Success, resp.Message, resp.SyncedCount)
}

// TestSyncLogic_SyncType 测试不同的同步类型
func TestSyncLogic_SyncType(t *testing.T) {
	tests := []struct {
		name     string
		syncType string
	}{
		{"全量同步", "full"},
		{"增量同步", "incremental"},
		{"默认同步", ""},
		{"未知类型", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svcCtx := &svc.ServiceContext{
				Config:  config.Config{},
				Context: context.Background(),
			}

			logic := NewSyncLogic(context.Background(), svcCtx)
			
			req := &proto.SyncRequest{
				SyncType: tt.syncType,
			}
			
			resp, err := logic.SyncData(req)
			
			// 不验证成功与否（因为需要实际数据库），只验证响应存在
			if resp == nil {
				t.Error("Response should not be nil")
			}
			
			if err != nil {
				t.Logf("Sync returned error (may be expected): %v", err)
			}
		})
	}
}

// TestHealthCheckLogic_NilServices 测试服务为 nil 时的健康检查
func TestHealthCheckLogic_NilServices(t *testing.T) {
	svcCtx := &svc.ServiceContext{
		Config:  config.Config{},
		Context: context.Background(),
		ES:      nil,
		MySQL:   nil,
	}

	logic := NewHealthCheckLogic(context.Background(), svcCtx)
	
	req := &proto.HealthCheckRequest{}
	resp, err := logic.HealthCheck(req)
	
	if err != nil {
		t.Fatalf("HealthCheck should handle nil services gracefully, got: %v", err)
	}

	if resp.Healthy {
		t.Error("Expected healthy=false when all services are nil")
	}

	t.Logf("Health check with nil services: %s", resp.Message)
}

// TestSyncLogic_ContextCancellation 测试上下文取消
func TestSyncLogic_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 立即取消

	svcCtx := &svc.ServiceContext{
		Config:  config.Config{},
		Context: ctx,
	}

	logic := NewSyncLogic(ctx, svcCtx)
	
	req := &proto.SyncRequest{
		SyncType: "full",
	}
	
	// 上下文已取消，同步应该失败或返回空结果
	resp, err := logic.SyncData(req)
	
	if err != nil {
		t.Logf("Sync with cancelled context returned error (expected): %v", err)
	}

	if resp == nil {
		t.Log("Response is nil due to cancelled context")
	}
}

// TestHealthCheckLogic_MultipleCalls 测试多次调用健康检查
func TestHealthCheckLogic_MultipleCalls(t *testing.T) {
	svcCtx := &svc.ServiceContext{
		Config:  config.Config{},
		Context: context.Background(),
	}

	logic := NewHealthCheckLogic(context.Background(), svcCtx)
	
	req := &proto.HealthCheckRequest{}
	
	// 多次调用
	var prevHealthy bool
	for i := 0; i < 3; i++ {
		resp, err := logic.HealthCheck(req)
		if err != nil {
			t.Fatalf("HealthCheck call %d failed: %v", i+1, err)
		}

		if i > 0 && resp.Healthy != prevHealthy {
			t.Error("Health status should be consistent across calls")
		}

		prevHealthy = resp.Healthy
		t.Logf("Call %d: healthy=%v, message=%s", i+1, resp.Healthy, resp.Message)
	}
}

// BenchmarkHealthCheckLogic_HealthCheck 基准测试：健康检查性能
func BenchmarkHealthCheckLogic_HealthCheck(b *testing.B) {
	svcCtx := &svc.ServiceContext{
		Config:  config.Config{},
		Context: context.Background(),
	}

	logic := NewHealthCheckLogic(context.Background(), svcCtx)
	req := &proto.HealthCheckRequest{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = logic.HealthCheck(req)
	}
}

// BenchmarkSyncLogic_SyncData 基准测试：数据同步性能
func BenchmarkSyncLogic_SyncData(b *testing.B) {
	svcCtx := &svc.ServiceContext{
		Config:  config.Config{},
		Context: context.Background(),
	}

	logic := NewSyncLogic(context.Background(), svcCtx)
	req := &proto.SyncRequest{
		SyncType: "full",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = logic.SyncData(req)
	}
}
