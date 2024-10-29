## 项目需求
### 注册资源
```
资源名称:jinli.io/gpu
资源数量:默认复制100个vGPU，示例如下：
        主机有2块物理gpu，p1gpu、p2gpu
        上报顺序为p1gpu-1，p2gpu-2，p1gpu-3，...
```
### 分配资源
```
Pod的每个容器通过环境变量限制显存和计算资源，通过resources选中gpu数量
ENV[GPUMEM] 整数, 代表显存数，单位m
ENV[GPUCORES] 整数 1到100，代表单个gpu的计算能力的百分比
Limits:jinli.io/gpu 整数，代表vgpu数量
```
### 管理状态
```
添加显存注解jinli.io/gpumems=uuid1-1024,uuid2-2048
更新分配状态注解AllocateStatus：allocated
添加Pod中每个Container的ENV[UUID](env["UUID"] = "uuid1,uuid2")
```

### 设备限额
```
根据Pod中每个Container的ENV[GPUMEM]和ENV[GPUCORES]
```

## 参考实现
### 设备插件框架
```
参考github.com/kubevirt/device-plugin-manager/pkg/dpm
```

### GPU操作
```
参考nvidia-device-plugin
```

### 设备限额
```
参考Project-HAMi/HAMi-core实现GPU显存和计算资源大小的分配
```

## 负载编排
```
apiVersion: v1
kind: Pod
metadata:
  name: jinli-gpu
spec:
  schedulerName: jinli-scheduler
  containers:
  - name: gpu-container
image: jinli.harbor.com/jinlik8s-device/app:v1.2.1
    command: ["sh", "-c", "tail -f /dev/null"]
    env:
    - name: GPUCORES[ 需要20%的SM计算能力]
      value: 20
    - name: GPUMEM[ 需要200m显存]
      value: 200
    - name: THRESHOLD
      value: 1e-6
    resources:
      limits:
        jinli.io/gpu: 1
```