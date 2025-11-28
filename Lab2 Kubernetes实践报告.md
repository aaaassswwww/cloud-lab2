# Lab2 Kubernetes实践报告

## 1.使用minikube搭建kubernetes集群

运行`minikube start --driver=docker`

![image-20251126133920563](images\image-20251126133920563.png)

## 2. 在Kubernetes集群中部署中间件

在gomall/k8s/middlewares中添加 Kubernetes manifests，包括 Redis、NATS 的 Deployment/Service， MySQL 的 ConfigMap、PV/PVC、StatefulSet、Service：

- nats-deployment.yaml -> NATS Deployment
- nats-service.yaml -> NATS Service
- mysqLconfigmap.yaml -> MySQL初始化 SQL
- mysqLpvpvc.yaml -> PersistentVolume 与PersistentVolumeClaim (hostPath)
- mysql-headless-service.yaml -> MySQL headless Service (StatefulSet 使用)
- mysql-service.yaml -> MySQL ClusterlP service
- mysql-statefulset.yaml -> MySQL StatefulSet(挂载 PVC 并使用 ConfigMap 初始化)

在集群中应用这些文件

```bash
minikube kubectl -- apply -f gomall/k8s/middlewares; 
```

![image-20251126134120511](images\image-20251126134120511.png)

运行`minikube kubectl -- get po -A`，查看pod都处于Running状态，说明三个中间件的容器已经成功启动。

![image-20251126173521775](images\image-20251126173521775.png)

运行` minikube kubectl -- get svc -n default`，检查`mysql`, `nats`, 和 `redis`已经创建了 `ClusterIP` 类型的 `Service`，说明它们可以在集群内部被访问。

![image-20251126173747518](images\image-20251126173747518.png)

运行`minikube kubectl -- get cm -n default`，已经创建了 `mysql-init-sql`。

运行`minikube kubectl -- describe statefulset mysql -n default` 的输出进一步确认了这个 `ConfigMap` 被正确挂载到了容器的 `/docker-entrypoint-initdb.d/init.sql` 路径，用于初始化。

![image-20251126173947957](images\image-20251126173947957.png)

![image-20251126174059487](images\image-20251126174059487.png)

综上，MySQL 已经实现了数据持久化。get pvc 显示已经创建了mysql-pvc 的 PersistentVolumeClaim，它的状态是 Bound，说明它成功绑定到了一个名为 `mysql-pv` 的 PersistentVolume。describe statefulset 确认了这个 mysql-pvc 被用作 mysql-data 卷，并挂载到了容器的 `/var/lib/mysql` 目录。

## 3. 在Kubernetes集群中部署gomall

**总体思路**

为每个微服务创建 `ConfigMap`（配置）、`Deployment`、`Service`、以及必要的 `Volume`/`PVC`，使得服务在集群内互相发现并稳定运行。

**要点**

- 微服务通过 `<中间件的服务名>:<端口号>` 访问中间件。在 `ConfigMap` 中把中间件地址设置为服务名：
  - MySQL: `address: mysql`、`port: 3306`
  - Redis: `address: "redis:6379"`
  - NATS: `url: "nats://nats:4222"`
- 使用 `ConfigMap` 存储微服务配置并挂载到容器中。每个微服务 YAML 文件顶部定义了一个 `ConfigMap`，并在`Deployment.spec.template.spec.volumes` / `volumeMounts` 中以 `subPath` 的方式挂载到 conf.yaml。这样能在不重建镜像的情况下更新配置。

在集群中应用微服务清单

```bash
minikube kubectl -- apply -f .\gomall\k8s\middlewares\mysql-pv-pvc.yaml
minikube kubectl -- apply -f .\gomall\k8s\middlewares\mysql-configmap.yaml
minikube kubectl -- apply -f .\gomall\k8s\middlewares\mysql-headless-service.yaml
minikube kubectl -- apply -f .\gomall\k8s\middlewares\mysql-service.yaml
minikube kubectl -- apply -f .\gomall\k8s\middlewares\mysql-statefulset.yaml
minikube kubectl -- apply -f .\gomall\k8s\middlewares\redis-deployment.yaml
minikube kubectl -- apply -f .\gomall\k8s\middlewares\redis-service.yaml
minikube kubectl -- apply -f .\gomall\k8s\middlewares\nats-deployment.yaml
minikube kubectl -- apply -f .\gomall\k8s\middlewares\nats-service.yaml
```

![image-20251126175135535](images\image-20251126175135535.png)

运行`minikube kubectl -- get pods -o wide`，可能会出现一些pod起不来的情况

![image-20251126181804352](images\image-20251126181804352.png)

![image-20251126181835767](images\image-20251126181835767.png)

运行` minikube kubectl -- logs cart-79f7f79fd9-rcchw`查看cart 日志，发现MySQL 还没有为 cart 等服务创建数据库并且授予 gomall 对这些数据库的权限。因为 MySQL 数据存储已被 PV 持久化，重新应用 ConfigMap 不会重新初始化，所以在运行中的 MySQL 实例里手动创建数据库并授权。

```bash
minikube kubectl -- exec mysql-0 -- mysql -uroot -proot123 -e "CREATE DATABASE IF NOT EXISTS cart DEFAULT CHARACTER SET utf8mb4;"
minikube kubectl -- exec mysql-0 -- mysql -uroot -proot123 -e "CREATE DATABASE IF NOT EXISTS product DEFAULT CHARACTER SET utf8mb4;"
minikube kubectl -- exec mysql-0 -- mysql -uroot -proot123 -e "CREATE DATABASE IF NOT EXISTS `user` DEFAULT CHARACTER SET utf8mb4;"
minikube kubectl -- exec mysql-0 -- mysql -uroot -proot123 -e "CREATE DATABASE IF NOT EXISTS `order` DEFAULT CHARACTER SET utf8mb4;"
minikube kubectl -- exec mysql-0 -- mysql -uroot -proot123 -e "CREATE DATABASE IF NOT EXISTS payment DEFAULT CHARACTER SET utf8mb4;"
minikube kubectl -- exec mysql-0 -- mysql -uroot -proot123 -e "CREATE USER IF NOT EXISTS 'gomall'@'%' IDENTIFIED BY 'gomall123';"
minikube kubectl -- exec mysql-0 -- mysql -uroot -proot123 -e "GRANT ALL PRIVILEGES ON cart.* TO 'gomall'@'%';"
minikube kubectl -- exec mysql-0 -- mysql -uroot -proot123 -e "GRANT ALL PRIVILEGES ON product.* TO 'gomall'@'%';"
minikube kubectl -- exec mysql-0 -- mysql -uroot -proot123 -e "GRANT ALL PRIVILEGES ON `user`.* TO 'gomall'@'%';"
minikube kubectl -- exec mysql-0 -- mysql -uroot -proot123 -e "GRANT ALL PRIVILEGES ON `order`.* TO 'gomall'@'%';"
minikube kubectl -- exec mysql-0 -- mysql -uroot -proot123 -e "GRANT ALL PRIVILEGES ON payment.* TO 'gomall'@'%';"
minikube kubectl -- exec mysql-0 -- mysql -uroot -proot123 -e "FLUSH PRIVILEGES;"
minikube kubectl -- exec mysql-0 -- mysql -uroot -proot123 -e 'CREATE DATABASE IF NOT EXISTS `order` DEFAULT CHARACTER SET utf8mb4;'
minikube kubectl -- exec mysql-0 -- mysql -uroot -proot123 -e 'GRANT ALL PRIVILEGES ON `order`.* TO ''gomall''@''%'';'
minikube kubectl -- exec mysql-0 -- mysql -uroot -proot123 -e 'FLUSH PRIVILEGES;';
minikube kubectl -- rollout restart deployment cart product order payment user
```

现在可以了

![image-20251126182329937](images\image-20251126182329937.png)

运行`kubectl port-forward service/frontend 8080:8080`

发现可以正常访问在 Kubernetes 集群中运行的 gomall 系统。

![image-20251126182524577](images\image-20251126182524577.png)