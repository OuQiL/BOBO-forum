@echo off
REM Search 微服务测试脚本 (Windows 版本)

setlocal enabledelayedexpansion

echo ========================================
echo   Search 微服务测试脚本
echo ========================================
echo.

REM 设置 ES 地址
set ES_HOST=%ES_HOST:localhost=%
set ES_PORT=%ES_PORT:9200%

if "%ES_HOST%"=="" set ES_HOST=localhost
if "%ES_PORT%"=="" set ES_PORT=9200

echo Elasticsearch 地址：http://%ES_HOST%:%ES_PORT%
echo.

echo ========================================
echo 1. 测试 Elasticsearch 连接
echo ========================================

curl -s http://%ES_HOST%:%ES_PORT%/_cluster/health > nul 2>&1
if %errorlevel% equ 0 (
    echo [OK] Elasticsearch 可达
    echo.
    echo 集群信息:
    curl -s http://%ES_HOST%:%ES_PORT%/
    echo.
) else (
    echo [ERROR] Elasticsearch 不可达
    echo 请检查 ES 是否运行：docker ps ^| grep elasticsearch
    exit /b 1
)

echo ========================================
echo 2. 检查 ik 分词器
echo ========================================

REM 测试 ik 分词器
curl -s -X POST "http://%ES_HOST%:%ES_PORT%/_analyze" ^
  -H "Content-Type: application/json" ^
  -d "{\"analyzer\":\"ik_smart\",\"text\":\"Python 编程教程\"}" > %TEMP%\ik_test.json

findstr /c:"tokens" %TEMP%\ik_test.json > nul
if %errorlevel% equ 0 (
    echo [OK] ik_smart 分词器可用
    echo.
    echo 分词测试结果:
    type %TEMP%\ik_test.json
    del %TEMP%\ik_test.json
) else (
    echo [ERROR] ik_smart 分词器不可用
    echo.
    echo 请安装 ik 分词器:
    echo docker exec -it elasticsearch bin/elasticsearch-plugin install ^
    echo   https://github.com/medcl/elasticsearch-analysis-ik/releases/download/v8.11.0/elasticsearch-analysis-ik-8.11.0.zip
    exit /b 1
)

echo ========================================
echo 3. 检查索引状态
echo ========================================

echo 现有索引:
curl -s "http://%ES_HOST%:%ES_PORT%/_cat/indices?v" | findstr /C:"posts" /C:"users"
if %errorlevel% neq 0 (
    echo 暂无搜索索引
)
echo.

echo ========================================
echo 4. 运行 Go 连接测试
echo ========================================

cd /d "%~dp0search"

if exist "go.mod" (
    echo 运行 ES 连接测试...
    echo.
    
    go test -v ./internal/es -run TestConnectionOnly
    if %errorlevel% equ 0 (
        echo.
        echo [OK] 连接测试通过
    ) else (
        echo.
        echo [ERROR] 连接测试失败
        echo 请检查 go.mod 和依赖是否正确安装
        exit /b 1
    )
    
    echo.
    echo 运行健康检查测试...
    echo.
    
    go test -v ./internal/es -run TestHealthCheck
    if %errorlevel% equ 0 (
        echo.
        echo [OK] 健康检查测试通过
    ) else (
        echo.
        echo [WARNING] 健康检查测试失败（可能是认证问题）
    )
) else (
    echo [ERROR] go.mod 文件不存在
    echo 请确保在正确的目录运行脚本
    exit /b 1
)

echo.
echo ========================================
echo 测试总结
echo ========================================
echo.
echo [OK] 所有测试完成
echo.
echo 下一步操作:
echo 1. 启动 search-rpc 服务：cd search ^&^& go run search.go -f etc/search.yaml
echo 2. 同步数据到 ES: grpcurl -plaintext -d "{\"sync_type\":\"full\"}" localhost:9093 search.SearchService/SyncData
echo 3. 测试搜索功能：grpcurl -plaintext -d "{\"keyword\":\"Python\",\"type\":\"post\"}" localhost:9093 search.SearchService/Search
echo.

endlocal
