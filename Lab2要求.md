# 云原生软件技术2025秋 Lab2 Kubernetes实践
TA: 曹勇虎 黄秋瑞


## 一、Lab核心目标与细节修改
### 1. 实践内容
- 搭建Kubernetes集群，将Lab1中的应用迁移到Kubernetes集群中
- 学习Helm的使用，并将Lab1应用通过Helm进行部署

### 2. 与Lab1的关联
本次Lab使用的案例系统在功能上与Lab1完全一致，Lab1中对案例系统的理解和分析对本次Lab有重要帮助。

### 3. 细节修改
- 去掉了`product`微服务中多余的etcd相关代码，避免造成困扰
- 将所有的配置项整合到`<服务名>/conf/dev/conf.yaml`中，避免部分配置在`.yaml`文件、部分在`.env`文件的混乱
- Lab2中的`gomall`不再依赖consul注册中心，而是直接通过`<服务名>:<端口号>`的方式调用（例如`cart:8883`），相关代码位于`<服务名>/infra/rpc/client.go`及`utils/utils.go`中
- 若在Kubernetes外运行Lab2，需手动配置`/etc/hosts`以解决DNS记录问题，需添加如下内容：
  ```
  127.0.0.1 cart
  127.0.0.1 checkout
  127.0.0.1 email
  127.0.0.1 frontend
  127.0.0.1 order
  127.0.0.1 payment
  127.0.0.1 product
  127.0.0.1 user
  ```


## 二、搭建Kubernetes集群
### 1. 集群搭建方式对比
| 搭建工具 | 特点 | 适用场景 |
| --- | --- | --- |
| Minikube | 在本地单个节点运行Kubernetes集群 | 开发、测试、本地部署 |
| k3d | 在Docker中运行k3s，可快速搭建本地集群 | 开发、测试、本地部署 |
| k3s | 轻量级Kubernetes发行版，内置网络/存储插件 | 生产环境（物理机/虚拟机多节点） |
| kubeadm | 快速部署原生Kubernetes，需手动安装网络/存储插件 | 生产环境（物理机/虚拟机多节点），灵活性高但复杂度高 |

**选择建议**：单机开发测试选Minikube或k3d；多节点实践选k3s。


### 2. 具体搭建步骤（二选一）
#### 2.1 使用Minikube搭建（1.1与1.2二选一）
课堂实践中已完成，此处略。

#### 2.2 使用k3d搭建（1.1与1.2二选一）
1. 参考k3d官方文档安装k3d工具
2. 使用助教定义的配置文件启动k3d集群（配置文件参考官方“Using Config Files”文档），命令如下：
   ```bash
   k3d cluster create --config k3d/k3d.yaml
   ```


## 三、在Kubernetes集群中部署中间件
### 1. 需部署的中间件
`redis`、`mysql`、`nats`（原可通过`gomall-middlewares/docker-compose.yaml`用Docker本地启动，本次需部署到K8s集群）

### 2. Deployment与StatefulSet区别
| 资源类型 | 用途 | 数据一致性保障 | 适用场景 |
| --- | --- | --- | --- |
| Deployment | 管理无状态应用 | 不保证Pod重启后数据一致性 | Web应用、API服务等无状态服务 |
| StatefulSet | 管理有状态应用 | 确保每个Pod唯一标识，Pod重启后数据一致 | 数据库等有状态服务 |

### 3. 中间件部署方案
| 中间件 | 部署方式 | 原因 | 关键配置 |
| --- | --- | --- | --- |
| redis | Deployment | 内存数据存储，gomall系统暂无需保存redis状态 | 使用Service暴露服务；无需持久化存储 |
| nats | Deployment | 轻量级消息队列，gomall系统暂无需保存nats状态 | 使用Service暴露服务；无需持久化存储 |
| mysql | StatefulSet | 关系型数据库，gomall系统需保存数据状态 | 1. 数据存储：可选`hostPath`（开发测试，挂载宿主机目录）或`PersistentVolumeClaim(PVC)+PersistentVolume(PV)`（生产环境，K8s持久化方案，探索此方案可获额外分数）；<br>2. 用Service暴露服务；<br>3. 用ConfigMap存储mysql初始化sql脚本，挂载到容器实现初始化 |

### 4. 提交要求
编写部署中间件的yaml文件，放在`gomall/k8s/middlewares`目录下。


## 四、在Kubernetes集群中部署gomall
### 1. gomall系统组成
#### 1.1 微服务列表
| 微服务 | 功能 | 目录 | 镜像 |
| --- | --- | --- | --- |
| cart | 购物车增删改查 | app/cart | buwandocker/cart:lab2 |
| checkout | 结账功能 | app/checkout | buwandocker/checkout:lab2 |
| email | 发送邮件 | app/email | buwandocker/email:lab2 |
| order | 订单增删改查 | app/order | buwandocker/order:lab2 |
| payment | 支付功能 | app/payment | buwandocker/payment:lab2 |
| product | 商品增删改查 | app/product | buwandocker/product:lab2 |
| user | 用户增删改查 | app/user | buwandocker/user:lab2 |
| frontend | 前端页面，转发请求到其他微服务 | frontend | buwandocker/frontend:lab2 |

#### 1.2 镜像使用说明
所有微服务的源代码与Dockerfile已提供，可直接使用助教构建的镜像，也可通过`docker build`自行构建。


### 2. 系统访问方式
完成中间件部署后，通过`kubectl port-forward`命令将`frontend`服务端口映射到本地，再通过浏览器访问，命令如下：
```bash
kubectl port-forward service/frontend 8080:8080
```
访问地址：`http://localhost:8080`


### 3. 部署要点与提示
- 为每个微服务（如cart）创建Deployment（管理Pod）和Service（暴露服务供其他微服务访问），Service需正确声明端口号和选择器，确保其他微服务可通过`<服务名>:<端口号>`访问
- 微服务访问中间件的地址需修改为`<中间件服务名>:<端口号>`（在微服务配置文件中调整）
- 用ConfigMap存储微服务配置文件，挂载到容器中，避免配置写死到镜像
- 建议将业务微服务与中间件部署在不同命名空间（仅TA个人喜好，不影响评分）
- K8s集群内部服务访问规则：<br>  若调用方与被调用服务（如svc1）在同一命名空间，直接用`svc1`访问；<br>  若不在同一命名空间（如svc1在ns1中），需用`svc1.ns1.svc.cluster.local`访问（CoreDNS自动创建DNS记录）


### 4. 提交要求
编写部署微服务的yaml文件，放在`gomall/k8s/microservices`目录下。


## 五、扩缩容与负载均衡实验
### 1. 实验原理
- 水平扩展：通过修改Deployment的副本数（replicas）实现，应对负载增加
- 负载均衡：Service将多个Pod暴露为一个服务，自动分发请求到后端Pod


### 2. 限流模拟
助教在每个微服务中添加基于令牌桶算法的限流器，代码位于`<服务名>/main.go`（以product服务为例）：
```go
func kitexInit() (opts []server.Option) {
......
rateLimitMiddleware := func(next endpoint.Endpoint) endpoint.Endpoint {
rateLimiter := rate.NewLimiter(rate.Limit(conf.GetConf().RateLimiter.Rate),
conf.GetConf().RateLimiter.BucketSize)
return func(ctx context.Context, req, resp interface{}) (err error) {
if conf.GetConf().RateLimiter.Enabled {
err = rateLimiter.Wait(ctx)
if err != nil {
return err
}
}
return next(ctx, req, resp)
}
}
return
}
```
限流配置文件：
```yaml
rate_limiter:
enabled: true
bucket_size: 3
rate: 3
```
- 限制：令牌桶大小3，发放速率3次/秒；请求速率长期超3次/秒会导致响应时间增加，需扩容product服务应对。


### 3. 实验报告要求
需包含以下内容：
- 扩容前后对比：一定并发量下，响应时间和QPS的对比（附截图）
- 扩容所使用的命令或步骤

**工具提示**：用`hey`命令测试性能，示例：`hey -c 20 -z 10s http://localhost:8080`（20个并发连接，持续10秒发送请求）；测试gomall主界面即可（依赖product服务）。<br>**负载均衡验证**：查看各product Pod日志，若均能看到`ListProductsService:`日志，说明负载均衡正常。


## 六、滚动更新实验
### 1. 实验原理
- 滚动更新：不中断服务的情况下，逐步替换旧版本应用；若新Pod健康检查失败，更新停止，保留旧Pod提供服务
- 回滚：更新失败后，用`kubectl rollout undo`命令回滚到上一版本


### 2. 实验准备
故障镜像：`buwandocker/product:lab2-unhealthy`（不监听product服务的gRPC端口8881）


### 3. 实验要求
#### 3.1 定义健康检查
在product的Deployment中定义健康检查，通过检查gRPC端口（8881）是否有TCP监听判断服务健康。

#### 3.2 设置更新策略
通过`maxUnavailable`（更新过程中最多不可用的Pod数）和`maxSurge`（更新过程中最多新增的Pod数）控制滚动更新策略，按需设置参数。

#### 3.3 实验步骤与报告内容
- 更新product服务镜像的命令或步骤
- 更新过程相关截图（更新前后Pod状态、健康检查状态等）
- 更新失败后的回滚命令或步骤

#### 3.4 提交要求
- 修改`gomall/k8s/microservices`目录下相关yaml文件
- 实验过程、截图、结果写入实验报告


## 七、Helm Chart实验
### 1. Helm Chart作用
解决K8s应用部署中`kubectl apply`逐条执行的不便，实现一键部署（`helm install`）和资源统一管理，类似Docker Compose对Docker的作用。


### 2. 学习建议
参考官方Helm Chart教程（含中文版指南），掌握核心使用方法。


### 3. Helm Chart编写要求
- 包含任务3（部署微服务）的所有资源（Deployment、Service、ConfigMap等），无需包含中间件部署
- 处理各yaml间的依赖关系及Helm Chart中template的先后加载方式
- 支持通过`values.yaml`或命令行参数配置以下信息（需提供默认值）：
  1. `mysql`、`redis`、`nats`的信息（host、port、username、password等）
  2. 每个服务Deployment下Pod的副本数
  3. 其他自定义参数（需在实验报告中说明自定义原因）


### 4. 提交要求
- Helm Chart放在`gomall/k8s/helm`目录下
- 实验报告中说明如何使用Helm Chart进行部署


## 八、提交内容汇总
| 任务 | 要提交的内容 | 备注 |
| --- | --- | --- |
| 任务一：搭建Kubernetes集群 | 选择的搭建方法，简要描述搭建过程和结果 | 写在实验报告中 |
| 任务二：部署中间件 | 部署中间件的yaml文件 | 放在`gomall/k8s/middlewares`目录下 |
| 任务三：部署微服务 | 部署微服务的yaml文件 | 放在`gomall/k8s/microservices`目录下 |
| 任务四：扩缩容与负载均衡 | 扩缩容与负载均衡的实验报告 | 含实验过程、截图、结果 |
| 任务五：滚动更新实验 | 滚动更新实验的yaml文件和实验报告 | 修改`gomall/k8s/microservices`下相关yaml；报告含过程、截图、结果 |
| 任务六：Helm Chart实验 | Helm Chart的目录和实验报告 | Chart放在`gomall/k8s/helm`目录下；报告说明部署方法 |
| 任务七：简短的实验报告 | 实验报告 | 简要描述实验过程与结果，附上组员分工 |


### 提交规范
- 将整个目录打包为压缩包，命名为`lab2-<小组编号>.zip`
- 提交到Elearning
- 截止日期：2025年11月30日 23:59


## 九、评分标准
| 评估项 | 权重 |
| --- | --- |
| 搭建Kubernetes集群 | 10% |
| 部署中间件 | 15% |
| 部署微服务 | 25% |
| 扩缩容与负载均衡 | 10% |
| 滚动更新实验 | 10% |
| Helm Chart实验 | 20% |
| 实验报告 | 10% |