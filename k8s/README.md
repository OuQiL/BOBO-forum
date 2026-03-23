# K8s 部署指南

## 前置要求

- Docker Desktop 启用 Kubernetes
- kubectl 已配置连接到本地 K8s 集群
- (可选) Ingress Controller (如 nginx-ingress)

## 项目架构

```
bobo-forum namespace
├── api-gateway (8080)     - API网关，统一入口
├── auth-rpc (8001)        - 认证服务
├── post-rpc (8002)        - 帖子服务
├── search-rpc (8003)      - 搜索服务
└── interaction-rpc (8004) - 互动服务
```

## 服务发现方式

本项目使用 Kubernetes 原生服务发现，无需 Etcd：

- 服务名: `auth-rpc.bobo-forum.svc.cluster.local:8001`
- 简写: `auth-rpc:8001` (同 namespace 内)

## 快速部署

### 1. 构建镜像

```bash
# Windows PowerShell
docker build -t bobo-forum/gateway:latest ./api-gateway
```

### 2. 部署到 K8s

```bash
# 开发环境
kubectl apply -k k8s/overlays/dev

# 生产环境
kubectl apply -k k8s/overlays/prod
```

### 3. 验证部署

```bash
# 查看所有资源
kubectl get all -n bobo-forum

# 查看 pods
kubectl get pods -n bobo-forum

# 查看服务
kubectl get svc -n bobo-forum

# 查看日志
kubectl logs -f deployment/api-gateway -n bobo-forum
```

### 4. 本地访问

```bash
# 端口转发
kubectl port-forward svc/api-gateway 8080:8080 -n bobo-forum

# 访问 API
curl http://localhost:8080/api/v1/search?keyword=test
```

## 文件结构

```
k8s/
├── base/                          # 基础配置
│   ├── namespace.yaml             # 命名空间
│   ├── configmap.yaml             # 配置映射
│   ├── auth-deployment.yaml       # Auth 服务部署
│   ├── auth-service.yaml          # Auth 服务
│   ├── post-deployment.yaml       # Post 服务部署
│   ├── post-service.yaml          # Post 服务
│   ├── search-deployment.yaml     # Search 服务部署
│   ├── search-service.yaml        # Search 服务
│   ├── interaction-deployment.yaml # Interaction 服务部署
│   ├── interaction-service.yaml   # Interaction 服务
│   ├── gateway-deployment.yaml    # Gateway 部署
│   ├── gateway-service.yaml       # Gateway 服务
│   ├── ingress.yaml               # Ingress 入口
│   └── kustomization.yaml         # Kustomize 配置
└── overlays/
    ├── dev/                       # 开发环境
    │   └── kustomization.yaml
    └── prod/                      # 生产环境
        └── kustomization.yaml
```

## 配置说明

### ConfigMap

- `bobo-forum-config`: 各服务端口配置
- `gateway-config`: Gateway 配置文件

### 资源限制

| 环境 | CPU Request | CPU Limit | Memory Request | Memory Limit |
|------|-------------|-----------|----------------|--------------|
| Dev  | 50m         | 200m      | 64Mi           | 256Mi        |
| Prod | 200m        | 1000m     | 256Mi          | 1Gi          |

### 副本数

| 服务 | Dev | Prod |
|------|-----|------|
| auth | 1   | 3    |
| post | 1   | 3    |
| search | 1 | 2    |
| interaction | 1 | 3  |
| gateway | 1 | 3   |

## 常用命令

```bash
# 删除所有资源
kubectl delete -k k8s/overlays/dev

# 重启某个服务
kubectl rollout restart deployment/auth-service -n bobo-forum

# 扩缩容
kubectl scale deployment/api-gateway --replicas=3 -n bobo-forum

# 进入容器
kubectl exec -it deployment/api-gateway -n bobo-forum -- /bin/sh

# 查看资源使用
kubectl top pods -n bobo-forum
```

## TODO

- [ ] 添加 HPA (Horizontal Pod Autoscaler)
- [ ] 添加 Prometheus 监控
- [ ] 添加分布式追踪
- [ ] 配置 TLS/HTTPS
- [ ] 添加 NetworkPolicy
- [ ] 配置 Pod Disruption Budget
