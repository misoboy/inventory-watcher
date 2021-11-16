# Inventory Watcher (by Golang)
재고 조회를 GoLang 으로 구현한 데몬 애플리케이션 입니다.
1. SAMG SHOP
2. AdenBike Shop
3. PPOMPPU
4. RULIWEB

## Prerequisite
GoLang (1.17.2) 버전으로 작성 되었습니다.

## Dependency
```
go mod tidy
```

## Go Build & Run
```
# Build
GOOS=linux GOARCH=amd64 go build -o ./build/_output/bin/application ./cmd/watcher/

# Run
go run ./cmd/watcher/main.go
```

## Docker Build & Push
```
# Build
docker build -t misoboy/inventory-watcher:latest -f ./build/Dockerfile ./build/_output/

# Push
docker push misoboy/inventory-watcher:latest
```

## Kubernetes Resources
```
# Resource
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: inventory-watcher-config
data:
  TELEGRAM_API_URL: ""
  TELEGRAM_BOT_TOKEN: ""
  TELEGRAM_CHAT_ID: ""
---
apiVersion: v1
kind: Pod
metadata:
  labels:
    run: inventory-watcher
  name: inventory-watcher
spec:
  containers:
  - image: misoboy/inventory-watcher:latest
    name: inventory-watcher
    envFrom:
    - configMapRef:
        name: inventory-watcher-config
---
```
