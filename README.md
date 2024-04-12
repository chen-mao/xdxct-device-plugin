# xdxct-device-plugin

xdxct-device-plugin 用于在k8s集群中调用 xdxgpu

### Requirement

- xdxct-docker-toolkit
- xdxgpu >= 162
- kubernets

### 部署
```shell
kubectl apply -f xdxct-device-plugin.yml
```

### Example
```shell
# 添加config
kubectl create configmap config-files --from-file=examples/config-maps
kubectl get configmap config-files
```

- xdxsmi
```shell
kubectl apply examples/utility/xdxsmi-demo-daemonset.yaml
kubectl logs xdxsmi-demo-ds-<>
```

- mpv-video
```shell
# 需要修改yaml以映射视频文件夹
kubectl apply examples/mpv-video/mpv-video-demo-deployment.yaml
# 进入容器
kubectl exec -it mpv-video-demo-<> -- bash
$ root@<>: mpv --no-audio -hwaccel=vaapi-copy <your_video_path_in_container>

### 编译
```shell
make
```