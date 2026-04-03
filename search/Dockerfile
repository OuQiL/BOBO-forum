FROM golang:1.25-alpine AS builder

WORKDIR /app

# 安装依赖
RUN apk add --no-cache git

# 复制 go.mod
COPY go.mod .
RUN go mod download

# 复制源代码
COPY . .

# 编译
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o search-rpc search.go

# 最终镜像
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/search-rpc .
COPY --from=builder /app/etc/search.yaml ./etc/

EXPOSE 9093

CMD ["./search-rpc", "-f", "etc/search.yaml"]
