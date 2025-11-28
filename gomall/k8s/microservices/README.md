# Gomall 微服务部署指南

本目录包含 gomall 系统所有微服务的 Kubernetes 部署文件。

## 目录结构

```
microservices/
├── namespace.yaml                # gomall 命名空间
├── cart-configmap.yaml          # cart 服务配置
├── cart-deployment.yaml         # cart 服务部署
├── cart-service.yaml            # cart 服务
├── checkout-configmap.yaml      # checkout 服务配置
├── checkout-deployment.yaml     # checkout 服务部署
├── checkout-service.yaml        # checkout 服务
├── email-configmap.yaml         # email 服务配置
├── email-deployment.yaml        # email 服务部署
├── email-service.yaml           # email 服务
├── order-configmap.yaml         # order 服务配置
├── order-deployment.yaml        # order 服务部署
├── order-service.yaml           # order 服务
├── payment-configmap.yaml       # payment 服务配置
├── payment-deployment.yaml      # payment 服务部署
├── payment-service.yaml         # payment 服务
├── product-configmap.yaml       # product 服务配置
├── product-deployment.yaml      # product 服务部署
├── product-service.yaml         # product 服务
├── user-configmap.yaml          # user 服务配置
├── user-deployment.yaml         # user 服务部署
├── user-service.yaml            # user 服务
├── frontend-configmap.yaml      # frontend 服务配置
├── frontend-deployment.yaml     # frontend 服务部署
└── frontend-service.yaml        # frontend 服务
```

## 架构说明

### 微服务列表

| 微服务 | 端口 | 功能 | 依赖的中间件 |
|--------|------|------|--------------|
| cart | 8883 | 购物车管理 | MySQL |
| checkout | 8884 | 结账功能 | NATS |
| email | 8888 | 邮件发送 | NATS |
| order | 8885 | 订单管理 | MySQL, Redis |
| payment | 8886 | 支付功能 | MySQL |
| product | 8881 | 商品管理 | MySQL, Redis |
| user | 8882 | 用户管理 | MySQL |
| frontend | 8080 | 前端页面 | Redis |

### 命名空间设计

- **default**: 中间件服务（mysql、redis、nats）
- **gomall**: 业务微服务

## 部署步骤

### 前提条件

1. Kubernetes 集群已经搭建完成
2. 中间件（mysql、redis、nats）已经在 `default` 命名空间部署完成

### 部署命令

1. **创建命名空间**
   ```bash
   kubectl apply -f namespace.yaml
   ```

2. **部署所有微服务**
   ```bash
   kubectl apply -f .
   ```

   或者逐个部署：
   ```bash
   # 先部署 ConfigMap
   kubectl apply -f cart-configmap.yaml
   kubectl apply -f checkout-configmap.yaml
   kubectl apply -f email-configmap.yaml
   kubectl apply -f order-configmap.yaml
   kubectl apply -f payment-configmap.yaml
   kubectl apply -f product-configmap.yaml
   kubectl apply -f user-configmap.yaml
   kubectl apply -f frontend-configmap.yaml

   # 再部署 Deployment
   kubectl apply -f cart-deployment.yaml
   kubectl apply -f checkout-deployment.yaml
   kubectl apply -f email-deployment.yaml
   kubectl apply -f order-deployment.yaml
   kubectl apply -f payment-deployment.yaml
   kubectl apply -f product-deployment.yaml
   kubectl apply -f user-deployment.yaml
   kubectl apply -f frontend-deployment.yaml

   # 最后部署 Service
   kubectl apply -f cart-service.yaml
   kubectl apply -f checkout-service.yaml
   kubectl apply -f email-service.yaml
   kubectl apply -f order-service.yaml
   kubectl apply -f payment-service.yaml
   kubectl apply -f product-service.yaml
   kubectl apply -f user-service.yaml
   kubectl apply -f frontend-service.yaml
   ```

3. **检查部署状态**
   ```bash
   # 查看所有 Pod
   kubectl get pods -n gomall

   # 查看所有 Service
   kubectl get svc -n gomall

   # 查看详细状态
   kubectl describe pod <pod-name> -n gomall
   ```

4. **访问前端服务**
   ```bash
   kubectl port-forward service/frontend 8080:8080 -n gomall
   ```
   
   然后在浏览器访问: http://localhost:8080

## 配置说明

### ConfigMap

每个微服务都有对应的 ConfigMap，用于存储配置文件 `conf.yaml`。主要配置项包括：

- **服务配置**: 服务名称、监听地址、日志配置
- **中间件地址**: 使用完全限定域名（FQDN）格式访问中间件
  - MySQL: `mysql.default.svc.cluster.local:3306`
  - Redis: `redis.default.svc.cluster.local:6379`
  - NATS: `nats.default.svc.cluster.local:4222`
- **限流配置**: 令牌桶算法参数

### 跨命名空间服务访问

由于中间件在 `default` 命名空间，微服务在 `gomall` 命名空间，因此需要使用完全限定域名（FQDN）访问：

格式: `<service-name>.<namespace>.svc.cluster.local`

例如:
- `mysql.default.svc.cluster.local`
- `redis.default.svc.cluster.local`
- `nats.default.svc.cluster.local`

### 同命名空间服务访问

微服务之间的调用（都在 gomall 命名空间）可以直接使用服务名：
- `cart:8883`
- `checkout:8884`
- `product:8881`
- 等等

## 故障排查

### 查看日志
```bash
kubectl logs <pod-name> -n gomall
kubectl logs <pod-name> -n gomall -f  # 实时查看
```

### 进入容器
```bash
kubectl exec -it <pod-name> -n gomall -- /bin/sh
```

### 常见问题

1. **Pod 无法启动**: 检查镜像是否存在，配置是否正确
2. **服务无法连接**: 检查 Service 的 selector 是否匹配 Pod 的 labels
3. **中间件连接失败**: 确认中间件服务在 default 命名空间已正常运行
4. **配置未生效**: 检查 ConfigMap 是否正确挂载到容器

## 清理资源

删除所有微服务：
```bash
kubectl delete -f .
```

或删除整个命名空间：
```bash
kubectl delete namespace gomall
```
