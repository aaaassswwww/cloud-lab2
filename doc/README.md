# 云原生软件技术2025秋 Lab2 Kubernetes实践
> TA: 曹勇虎 黄秋瑞

在本次Lab中，我们将实践以下内容：
- 搭建Kubernetes集群，将Lab1中的应用迁移到Kubernetes集群中
- 学习Helm的使用，并将lab1应用通过Helm进行部署
  
本次Lab使用的案例系统在功能上**与Lab1完全一致**，因此在Lab1中对案例系统的理解和分析将对本次Lab有很大的帮助。在细节上进行了如下修改：

- 去掉了`product`微服务中多余的etcd相关代码，避免造成困扰
- 将所有的配置项整合到`<服务名>/conf/dev/conf.yaml`中。避免部分配置在`.yaml`文件中，而另一些配置在`.env`文件中，避免造成困扰
- 为了让同学们实践`Kubernetes`中`Service`的使用，Lab2中的`gomall`不再依赖`consul`注册中心，而是直接通过`<服务名>:<端口号>`的方式进行调用（例如`cart:8883`）。这一改动的相关代码位于`<服务名>/infra/rpc/client.go`及`utils/utils.go`中，有兴趣的同学可以看一下源码

在`Kubernetes`中，`Service`会自动为每个微服务创建一个DNS记录，方便其他微服务进行调用。如果你要在`Kuberentes`外运行Lab2，你需要自己手动配一下`hosts`，否则会找不到DNS记录。你可以在`/etc/hosts`中添加如下内容

```bash
127.0.0.1 cart
127.0.0.1 checkout
127.0.0.1 email
127.0.0.1 frontend
127.0.0.1 order
127.0.0.1 payment
127.0.0.1 product
127.0.0.1 user
```

  
## 1. 搭建Kubernetes集群

有以下⼏种⽅式可以搭建Kubernetes集群：

- **使⽤Minikube**：[Minikube](https://minikube.sigs.k8s.io/docs/)是⼀个在本地运⾏Kubernetes的⼯具，它可以在单个节点上运⾏⼀个Kubernetes集群。可以⽤于开发、测试和本地部署。
- **使⽤k3d**：[k3d](https://k3d.io/stable/)是⼀个在Docker中运⾏k3s的⼯具，可以在本地快速搭建⼀个Kubernetes集群，适合⽤于开发、测试和本地部署。
- **使⽤k3s**：[k3s](https://k3s.io/)是⼀个轻量级的Kubernetes发⾏版，可以在**真实的多个节点（物理机或虚拟机）上运⾏⼀个Kubernetes集群**，适合⽤于⽣产环境。k3s轻量且易于安装，内置了⽹络插件、存储插件等，可以快速搭建⼀个⽣产级别的Kubernetes集群。 
- **使⽤kubeadm**：[kubeadm](https://kubernetes.io/zh-cn/docs/reference/setup-tools/kubeadm/)是⼀个⽤于快速部署Kubernetes集群的⼯具，可以在真实的多个节点（物理机或虚拟机）上运⾏⼀个Kubernetes集群，适合⽤于⽣产环境。相⽐于k3s，kubeadm需要⼿动安装⽹络插件、存储插件等。kubeadm安装的是原⽣的Kubernetes，更加灵活，但也更加复杂。

在本次实验中，你可以选择任意一种方式搭建Kubernetes集群。如果只是单机开发和测试，建议选择Minikube或k3d。如果你想要在真实的多个节点上搭建一个Kubernetes集群来玩一玩，建议选择k3s。

### 1.1 使用Minikube搭建Kubernetes集群（与1.2二选一）

相信大家在课堂实践中已经完成了。

### 1.2 使用k3d搭建Kubernetes集群（与1.1二选一）

首先，参考[k3d的官方文档](https://k3d.io/stable/#installation)安装好k3d工具。

接着，使用助教定义好的配置文件在本地启动一个k3d集群。关于k3d集群的配置文件可以参考[Using Config Files](https://k3d.io/stable/usage/configfile/)

```bash
k3d cluster create --config k3d/k3d.yaml
```

## 2. 在Kubernetes集群中部署中间件

本次需要部署的中间件包含`redis`，`mysql`和`nats`。你目前已经可以通过`gomall-middlewares/docker-compose.yaml`，利用`docker`命令在本地启动这些中间件。**本次Lab中，你的任务是将这些中间件部署到Kubernetes集群中。**

在Kubernetes中，StatefulSet是用于管理有状态应用的API对象。它可以确保每个Pod都有一个唯一的标识符，并且可以在Pod重启时保持数据的一致性。StatefulSet通常用于数据库等有状态应用。与此不同，Deployment是用于管理无状态应用的API对象。它可以确保每个Pod都有一个唯一的标识符，但不保证Pod重启时数据的一致性。Deployment通常用于Web应用、API服务等无状态应用。我们可以根据中间件不同的特性，选择使用Deployment还是StatefulSet来部署中间件。

以下要点与提示可供参考：

- `redis`：使用Deployment来部署Redis。Redis是一个内存数据存储系统，通常用于缓存和消息队列等场景。在`gomall`系统中我们暂时不需要保存`redis`的状态，因此可以使用Deployment来部署Redis。
- `nats`：与`redis`类似，使用Deployment来部署NATS。NATS是一个轻量级的消息队列系统，通常用于微服务之间的异步通信。在`gomall`系统中我们暂时不需要保存`nats`的状态，因此可以使用Deployment来部署NATS。
- `mysql`：使用StatefulSet来部署MySQL。MySQL是一个关系型数据库，通常用于存储数据。在`gomall`系统中我们需要保存`mysql`的状态，因此可以使用StatefulSet来部署MySQL。
- 数据存储方案：在部署`mysql`时，你可以选择使用`hostPath`或`PersistentVolumeClaim`与`PersistentVolume`来存储数据。`hostPath`是将宿主机的目录挂载到容器中，适合开发和测试环境；而`PersistentVolumeClaim`与`PersistentVolume`是Kubernetes提供的持久化存储方案，适合生产环境。如果你在本次实验中探索了`PersistentVolumeClaim`与`PersistentVolume`的使用，可以得到更多的分数。
- 使用`Service`将中间件暴露出来。`Service`是Kubernetes提供的负载均衡和服务发现的API对象，可以将Pod暴露为一个服务。在本次实验中，你需要使用`Service`将中间件暴露出来，以便其他微服务可以访问它们。
- 使用`ConfigMap`存储`mysql`启动时的初始化sql脚本，并挂在到`mysql`容器中，实现初始化。
  
请编写相关的`yaml`文件来部署中间件，并将其放在`gomall/k8s/middlewares`目录下。

## 3. 在Kubernetes集群中部署gomall

`gomall`是一个简单的商城系统，包含了如下的微服务：

- `cart`: 购物车服务，提供了购物车的增删改查功能，位于`app/cart`目录，镜像为`buwandocker/cart:lab2`
- `checkout`: 结账服务，提供了结账功能，位于`app/checkout`目录，镜像为`buwandocker/checkout:lab2`
- `email`: 邮件服务，提供了发送邮件的功能，位于`app/email`目录，镜像为`buwandocker/email:lab2`
- `order`: 订单服务，提供了订单的增删改查功能，位于`app/order`目录，镜像为`buwandocker/order:lab2`
- `payment`: 支付服务，提供了支付功能，位于`app/payment`目录，镜像为`buwandocker/payment:lab2`
- `product`: 商品服务，提供了商品的增删改查功能，位于`app/product`目录，镜像为`buwandocker/product:lab2`
- `user`: 用户服务，提供了用户的增删改查功能，位于`app/user`目录，镜像为`buwandocker/user:lab2`

此外，还有一个`frontend`服务，提供了一个简单的前端页面，并将请求转发给上述的微服务，位于`frontend`目录，镜像为`buwandocker/frontend:lab2`。所有微服务的源代码与Dockerfile均已提供，你可以直接使用上述助教构建好的镜像进行实验，也可以使用`docker build`命令自行构建镜像。本次Lab中，**你的任务是将这些微服务部署到Kubernetes集群中，并确保服务间可以正常通信，并且可以通过`frontend`服务访问到所有的微服务**。

当你完成中间件部署后，你应该可以通过以下方法访问在`Kubernetes`集群中运行的`gomall`系统：

```bash
# 通过`kubectl port-forward`命令将`frontend`服务的端口映射到本地，然后在浏览器中访问`http://localhost:8080`。例如：
kubectl port-forward service/frontend 8080:8080
```

以下要点与提示可供参考：

- 可以为每个微服务(e.g. `cart`)创建一个`Deployment`资源和`Service`资源。`Deployment`资源用于管理微服务的Pod，`Service`资源用于将微服务暴露出来供其他微服务访问。在`Service`中，你需要正确地声明端口号和选择器，以便其他微服务可以通过`<服务名>:<端口号>`的方式访问到该微服务。
- 你的微服务可以通过`<中间件的服务名>:<端口号>`的方式访问到中间件。你需要在每个微服务的配置文件中修改中间件的地址为`<中间件的服务名>:<端口号>`的方式。
- 你可以使用`ConfigMap`来存储微服务的配置文件，并将其挂载到微服务的容器中。这样可以避免在`docker build`时打包写死的配置文件到镜像中。
- 建议将业务微服务与中间件部署在不同的命名空间中，以便于管理和维护（仅为TA个人喜好，是否这样做不影响评分）。
  
请编写相关的`yaml`文件来部署微服务，并将其放在`gomall/k8s/microservices`目录下。

**tips1**: 假设在`Kubernetes`集群中，`ns1`是一个命名空间，`svc1`是一个服务，则集群内部访问`svc1`的方式为：`svc1.ns1.svc.cluster.local`。`Kubeternetes`集群内部`CoreDNS`会自动为每个服务创建一个DNS记录，方便其他服务进行调用。如果发起调用的服务与被调用的服务在同一个命名空间中，则可以直接使用`svc1`来访问；如果不在同一个命名空间中，则需要使用`svc1.ns1.svc.cluster.local`来访问。

## 4. 扩缩容与负载均衡实验

我们知道当系统的负载增加时，可以通过水平扩展和负载均衡来提高系统的性能和可用性。`Kubernetes`集群中的`Deployment`都支持水平扩展（Horizontal Scaling），只需要通过修改`Deployment`的副本数（replicas）即可实现。`Kubernetes`会为每个`Pod`分配一个唯一的IP地址。通过`Service`可以将多个`Pod`暴露为一个服务。当其他服务请求`Service`时，`Service`会自动将请求分发到后端的多个`Pod`上，从而实现负载均衡。

在本次试验中，**你将体验到`Kubernetes`集群中的水平扩展和负载均衡的功能**。为了模拟高负载下每个微服务实例响应变慢的情况，助教**人为地**在每个微服务中添加了一个[基于令牌桶算法的限流器](https://www.cnblogs.com/DTinsight/p/18221858)。相关代码位于`<服务名>/main.go`中，例如在`product`服务中，人为限流的代码如下：

```go
func kitexInit() (opts []server.Option) {
    ......

	rateLimitMiddleware := func(next endpoint.Endpoint) endpoint.Endpoint {
		rateLimiter := rate.NewLimiter(rate.Limit(conf.GetConf().RateLimiter.Rate), conf.GetConf().RateLimiter.BucketSize)
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

    ......

	return
}
```

相关配置文件如下：

```yaml
rate_limiter:
  enabled: true
  bucket_size: 3
  rate: 3
```

上述代码限制了令牌桶的大小3，令牌的发放速率为3次/s。如果请求没有获取到令牌，则会阻塞等待，直到获取到令牌为止。你可以通过修改`rate`和`bucket_size`的值来调整限流的策略。在上述限制下，如果请求速率长期超过3次/s，则会导致请求响应时间快速增加。因此，你需要通过扩容`product`服务的副本数来应对这种情况。

请将以下内容写入实验报告中：

- 扩容前后对比：在扩容前后，在一定并发数量的情况下，扩容前后的响应时间和QPS的对比（截图）
- 扩容所使用的命令或步骤

**tips**: 你可以`hey`命令来测试微服务的性能。`hey`是一个HTTP负载测试工具，可以用来测试HTTP服务的性能。例如`hey -c 20 -z 10s http://localhost:8080`表示使用20个并发连接，持续10秒钟向`http://localhost:8080`发送请求。你可以根据自己的需求调整并发连接数和请求数量。为了简单起见，你只需要测试访问`gomall`的主界面，即`http://<frontend-ip>:<frontend-port>`即可，这个主界面会调用`product`微服务以获取所有商品信息，因此访问主界面的请求响应时间会依赖于访问`product`微服务的请求响应时间。

**tips2**: 为检查负载均衡正常工作，你可以查看每个`product`微服务的`Pod`的日志。若在多个`Pod`中都能看到`ListProductsService:`的日志，则说明负载均衡正常工作。


## 5. 滚动更新实验

在`Kubernetes`中，滚动更新是指在不影响服务可用性的情况下，逐步将旧版本的应用替换为新版本的应用。`Kubernetes`会自动管理Pod的创建和删除，以确保在更新过程中始终有足够的Pod可用。在更新过程中，如果新创建的`Pod`无法正常工作（不能通过健康检查），则滚动更新会停止，并确保始终有一些旧版本的`Pod`能对外提供服务，以避免服务中断。当发现滚动更新失败后，可以使用`kubectl rollout undo`命令将应用回滚到上一个版本。

在本次试验中，助教已经为你准备好了一个存在故障的`product`微服务的镜像`buwandocker/product:lab2-unhealthy`，你需要利用这个镜像来进行滚动更新实验。请将以下内容写入实验报告或相关的`.yaml`文件中：

- 定义健康检查的方式：可以通过检查`product`服务的`gRPC`端口(8881)上是否有被`tcp`监听到来判断`product`服务是否健康（`buwandocker/product:lab2-unhealthy`不会监听这个端口）。你需要在`product`的`Deployment`中定义健康检查。
- 设置更新策略：可以通过设置`maxUnavailable`和`maxSurge`来控制滚动更新的策略。`maxUnavailable`表示在更新过程中，最多有多少个Pod不可用；`maxSurge`表示在更新过程中，最多有多少个Pod是新创建的。你可以根据自己的需求设置这两个参数。
- 你更新`product`服务的镜像的命令或步骤。
- 更新过程中的相关截图，包括更新前后的Pod状态、健康检查的状态等。
- 更新失败后的回滚命令或步骤。

## 6. Helm Chart实验

在lab1中，当部署 Docker 应⽤时，通过逐条执⾏ Docker 命令⾏确实有些麻烦。使⽤ Docker Compose 可以实现⼀键部署。⽽当部署⾃⼰编写的 Kubernetes 应⽤时，通过逐条执⾏`kubectl apply`命令同样会感到不便，在卸载应⽤时也是。这样也很难对资源进⾏统⼀管理。⽽ Helm Chart 可以很好地解决这个问题。在这⼀步，**我们将把之前编写的Kubernetes yaml⽂件（资源）打包成Helm Chart**，并使⽤Helm部署应⽤。这样任何⼈都可以通过⼀句 `helm install`命令和⼀个配置⽂件来⼀键部署您的应⽤。

作为⼀名合格的开发者，学会阅读官⽅⽂档是必不可少的。由于官⽅ Helm Chart 教程已经⾮常详细，⽽且有中⽂版指南，因此我们建议你⾃⼰阅读并实践。

你编写的Helm Chart应该满⾜如下的要求：

- 包含[任务3](#3-在kubernetes集群中部署gomall)中创建的所有资源，包括Deployment、Service、ConfigMap等。**不需要包含中间件的部署**。
- 注意各个yaml之间的依赖关系以及在helm chart中template的先后加载方式。
- 用户启动Helm Chart时，应该可以通过values.yaml文件或者命令行参数配置以下信息，并提供默认值。
  - `mysql`、`redis`和`nats`的信息，如`host`、`port`、`username`、`password`等。
  - 每个服务对应`Deployment`下`Pod`的副本数。
  - 其他你觉得适合自定义的参数，并在实验报告中说明原因。
  
请将编写的Helm Chart放在`gomall/k8s/helm`目录下，并在实验报告中说明如何使用Helm Chart进行部署。

## 提交内容

对于每个任务，你需要提交以下内容：

| 任务                       | 要提交的内容                                     | 备注                                                                                       |
| -------------------------- | --------------------------------------------- | ------------------------------------------------------------------------------------------ |
| 任务一：搭建Kubernetes集群 | 你们选择的搭建方法，并简要描述搭建过程和搭建结果 | 可以写在实验报告中                                                                         |
| 任务二：部署中间件         | 部署中间件的yaml文件                             | 放在`gomall/k8s/middlewares`目录下                                                         |
| 任务三：部署微服务         | 部署微服务的yaml文件                             | 放在`gomall/k8s/microservices`目录下                                                       |
| 任务四：扩缩容与负载均衡   | 扩缩容与负载均衡的实验报告                       | 将实验过程，截图和结果写在实验报告中                                                       |
| 任务五：滚动更新实验       | 滚动更新实验的yaml文件和实验报告                 | 修改`gomall/k8s/microservices`目录下相关的`yaml`文件，将实验过程，截图和结果写在实验报告中 |
| 任务六：Helm Chart实验     | Helm Chart的目录和实验报告                       | 将Helm Chart放在`gomall/k8s/helm`目录下，将实验过程，截图和结果写在实验报告中              |
| 任务七：**简短**的实验报告 | 实验报告                                         | 简要描述实验过程与结果，并附上组员的分工情况                                               |

请将整个目录打包成一个压缩包，并命名为`lab2-<小组编号>.zip`，并提交到`Elearning`上。

**截止日期：2025年11月30日 23:59**

## 评分标准

本次Lab的评分标准如下：

| 评估项             | 权重 |
| ------------------ | ---- |
| 搭建Kubernetes集群 | 10%  |
| 部署中间件         | 15%  |
| 部署微服务         | 25%  |
| 扩缩容与负载均衡   | 10%  |
| 滚动更新实验       | 10%  |
| Helm Chart实验     | 20%  |
| 实验报告           | 10%  |