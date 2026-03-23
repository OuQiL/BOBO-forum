# 微服务开发完成后的部署教程

本文档指导你如何在完成一个微服务开发后，将其部署到 Kubernetes 集群。

## 目录

- [1. 准备工作](#1-准备工作)
- [2. 创建微服务目录结构](#2-创建微服务目录结构)
- [3. 编写 Dockerfile](#3-编写-dockerfile)
- [4. 构建镜像](#4-构建镜像)
- [5. 推送镜像到仓库](#5-推送镜像到仓库)
- [6. 更新 K8s 配置](#6-更新-k8s-配置)
- [7. 部署到集群](#7-部署到集群)
- [8. 验证部署](#8-验证部署)
- [9. 常见问题](#9-常见问题)

---

## 1. 准备工作

确保以下工具已安装：

- Docker Desktop (启用 Kubernetes)
- kubectl (已配置连接到本地集群)
- Go 1.22+
- goctl (go-zero 代码生成工具)

```bash
# 验证环境
docker version
kubectl version
go version
goctl --version
```

---

## 2. 创建微服务目录结构

以 `auth` 服务为例，标准目录结构如下：

```
service/auth/
├── auth.go              # 主入口
├── etc/
│   └── auth.yaml        # 配置文件
├── internal/
│   ├── config/
│   │   └── config.go    # 配置结构
│   ├── server/
│   │   └── authserver.go # gRPC 服务
│   ├── logic/
│   │   ├── registerlogic.go
│   │   └── loginlogic.go
│   ├── svc/
│   │   └── servicecontext.go
│   └── model/           # 数据模型 (可选)
│       └── usermodel.go
├── proto/
│   └── auth.proto       # Protobuf 定义
├── pb/
│   └── auth.pb.go       # 生成的代码
├── Dockerfile
└── go.mod
```

### 使用 goctl 生成代码

```bash
# 创建 RPC 服务
cd service/auth
goctl rpc new . --style=go_zero

# 或从 proto 文件生成
goctl rpc protoc proto/auth.proto --go_out=. --go-grpc_out=. --zrpc_out=.
```

---

## 3. 编写 Dockerfile

在微服务目录下创建 `Dockerfile`：

```dockerfile
# service/auth/Dockerfile
FROM golang:1.22-alpine AS builder

WORKDIR /app

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /auth .

# 运行阶段
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /auth .
COPY etc/auth.yaml ./etc/

EXPOSE 8001

CMD ["./auth", "-f", "etc/auth.yaml"]
```

---

## 4. 构建镜像

### 本地开发环境

```bash
# 在项目根目录执行
docker build -t bobo-forum/auth:dev ./service/auth

# 验证镜像
docker images | grep bobo-forum
```

### 生产环境

```bash
# 设置镜像仓库地址
REGISTRY=your-registry.com

# 构建并打标签
docker build -t ${REGISTRY}/bobo-forum/auth:v1.0.0 ./service/auth
```

---

## 5. 推送镜像到仓库

### 本地 Docker Desktop K8s

本地集群可以直接使用本地镜像，无需推送：

```bash
# 确保 Kubernetes 使用本地镜像
# 在 deployment.yaml 中设置 imagePullPolicy: IfNotPresent
```

### 远程仓库

```bash
# 登录镜像仓库
docker login your-registry.com

# 推送镜像
docker push your-registry.com/bobo-forum/auth:v1.0.0
```

---

## 6. 更新 K8s 配置

### 6.1 更新 Deployment

如果新增微服务，创建新的 deployment 文件：

```yaml
# k8s/base/auth-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: auth-service
  namespace: bobo-forum
  labels:
    app: auth-service
spec:
  replicas: 2
  selector:
    matchLabels:
      app: auth-service
  template:
    metadata:
      labels:
        app: auth-service
    spec:
      containers:
        - name: auth
          image: bobo-forum/auth:dev
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 8001
          env:
            - name: PORT
              value: "8001"
          resources:
            requests:
              cpu: 100m
              memory: 128Mi
            limits:
              cpu: 500m
              memory: 512Mi
          livenessProbe:
            tcpSocket:
              port: 8001
            initialDelaySeconds: 10
            periodSeconds: 10
          readinessProbe:
            tcpSocket:
              port: 8001
            initialDelaySeconds: 5
            periodSeconds: 5
```

### 6.2 更新 Service

```yaml
# k8s/base/auth-service.yaml
apiVersion: v1
kind: Service
metadata:
  name: auth-rpc
  namespace: bobo-forum
  labels:
    app: auth-service
spec:
  type: ClusterIP
  selector:
    app: auth-service
  ports:
    - name: rpc
      port: 8001
      targetPort: 8001
      protocol: TCP
```

### 6.3 更新 Kustomization

```yaml
# k8s/base/kustomization.yaml
resources:
  # ... 其他资源
  - auth-deployment.yaml
  - auth-service.yaml
```

---

## 7. 部署到集群

### 开发环境

```bash
# 应用配置
kubectl apply -k k8s/overlays/dev

# 或仅更新特定服务
kubectl apply -f k8s/base/auth-deployment.yaml
kubectl apply -f k8s/base/auth-service.yaml
```

### 生产环境

```bash
# 应用配置
kubectl apply -k k8s/overlays/prod
```

### 更新镜像版本

```bash
# 方法1: 修改 kustomization.yaml 中的镜像标签后重新 apply

# 方法2: 直接设置镜像
kubectl set image deployment/auth-service auth=bobo-forum/auth:v1.0.1 -n bobo-forum

# 方法3: 重启 Pod (拉取最新镜像)
kubectl rollout restart deployment/auth-service -n bobo-forum
```

---

## 8. 验证部署

### 检查资源状态

```bash
# 查看 Pod 状态
kubectl get pods -n bobo-forum

# 查看详细信息
kubectl describe pod <pod-name> -n bobo-forum

# 查看日志
kubectl logs -f deployment/auth-service -n bobo-forum

# 查看所有资源
kubectl get all -n bobo-forum
```

### 测试服务连通性

```bash
# 端口转发
kubectl port-forward svc/auth-rpc 8001:8001 -n bobo-forum

# 在另一个终端测试
grpcurl -plaintext localhost:8001 list
```

### 检查服务发现

```bash
# 进入 gateway 容器
kubectl exec -it deployment/api-gateway -n bobo-forum -- /bin/sh

# 测试 DNS 解析
nslookup auth-rpc
ping auth-rpc

# 测试服务连接
nc -zv auth-rpc 8001
```

---

## 9. 常见问题

### Q1: ImagePullBackOff / ErrImagePull

**原因**: 镜像不存在或无法拉取

**解决**:
```bash
# 检查镜像是否存在
docker images | grep bobo-forum/auth

# 本地构建镜像
docker build -t bobo-forum/auth:dev ./service/auth

# 如果使用远程仓库，确保已登录并推送
docker login your-registry.com
docker push your-registry.com/bobo-forum/auth:dev
```

### Q2: CrashLoopBackOff

**原因**: 容器启动后立即退出

**解决**:
```bash
# 查看日志
kubectl logs <pod-name> -n bobo-forum

# 查看事件
kubectl describe pod <pod-name> -n bobo-forum

# 常见原因:
# - 配置文件错误
# - 端口冲突
# - 依赖服务未启动
# - 环境变量缺失
```

### Q3: 服务间无法通信

**原因**: Service DNS 解析问题

**解决**:
```bash
# 检查 Service 是否存在
kubectl get svc -n bobo-forum

# 检查 Endpoints
kubectl get endpoints -n bobo-forum

# 检查 DNS
kubectl run -it --rm debug --image=busybox --restart=Never -- nslookup auth-rpc.bobo-forum.svc.cluster.local
```

### Q4: Pod 一直处于 Pending 状态

**原因**: 资源不足或调度问题

**解决**:
```bash
# 查看事件
kubectl describe pod <pod-name> -n bobo-forum

# 检查节点资源
kubectl describe nodes

# 可能需要调整资源请求
# 在 deployment.yaml 中修改 resources.requests
```

### Q5: 配置更新不生效

**原因**: ConfigMap 更新后 Pod 未重启

**解决**:
```bash
# 重启 Deployment
kubectl rollout restart deployment/auth-service -n bobo-forum

# 或删除 Pod 让其自动重建
kubectl delete pod -l app=auth-service -n bobo-forum
```

---

## 快速命令参考

```bash
# 完整部署流程
cd service/auth
docker build -t bobo-forum/auth:dev .
kubectl apply -k ../../k8s/overlays/dev
kubectl get pods -n bobo-forum -w

# 查看日志
kubectl logs -f deployment/auth-service -n bobo-forum

# 进入容器调试
kubectl exec -it deployment/auth-service -n bobo-forum -- /bin/sh

# 端口转发测试
kubectl port-forward svc/auth-rpc 8001:8001 -n bobo-forum

# 删除服务
kubectl delete -f k8s/base/auth-deployment.yaml
kubectl delete -f k8s/base/auth-service.yaml

# 完全清理
kubectl delete namespace bobo-forum
```

---

## 检查清单

完成微服务开发后，按此清单检查：

- [ ] 代码已通过本地测试
- [ ] Dockerfile 已创建并测试构建
- [ ] 镜像已构建并推送
- [ ] K8s Deployment 配置正确
- [ ] K8s Service 配置正确
- [ ] ConfigMap 配置已更新（如有新配置项）
- [ ] 已添加到 kustomization.yaml
- [ ] 部署成功，Pod 状态为 Running
- [ ] 日志无错误
- [ ] 服务间通信正常
- [ ] API Gateway 路由配置正确
