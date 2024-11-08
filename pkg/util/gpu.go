package util

import (
	"fmt"
	"os"

	"github.com/google/uuid"
	corev1 "k8s.io/api/core/v1"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

func LimitGPUMemAndCores(response *pluginapi.ContainerAllocateResponse, current *corev1.Pod, idx int32) {
	// gpu资源限制
	// HAMI-core中CUDA_DEVICE_MEMORY_LIMIT_ID（限制容器指定设备显存）会覆盖CUDA_DEVICE_MEMORY_LIMIT（限制容器所有设备显存）
	// 例如：response.Envs["CUDA_DEVICE_MEMORY_LIMIT_0"] = "20m"
	// 获取gpumem、gpucores
	hostPath := "/usr/local/jinli"
	envs := current.Spec.Containers[idx].Env
        response.Envs = make(map[string]string)
	for _, v := range envs {
		if v.Name == "GPUMEM" {
			response.Envs["CUDA_DEVICE_MEMORY_LIMIT"] = fmt.Sprintf("%vm", v.Value)
		}
		if v.Name == "GPUCORES" {
			response.Envs["CUDA_DEVICE_SM_LIMIT"] = fmt.Sprintf("%v", v.Value)
		}
	}

	response.Envs["CUDA_DEVICE_MEMORY_SHARED_CACHE"] = fmt.Sprintf("%s/%v.cache", hostPath, uuid.New().String())
	response.Envs["CUDA_OVERSUBSCRIBE"] = "true"

	cacheFileHostDirectory := fmt.Sprintf("%s/containers/%s_%s", hostPath, current.UID, current.Spec.Containers[idx].Name)
	os.RemoveAll(cacheFileHostDirectory)
	os.MkdirAll(cacheFileHostDirectory, 0777)
	os.Chmod(cacheFileHostDirectory, 0777)

	response.Mounts = append(response.Mounts,
		&pluginapi.Mount{
			ContainerPath: fmt.Sprintf("%s/libvgpu.so", hostPath),
			HostPath:      hostPath + "/libvgpu.so",
			ReadOnly:      true},
		&pluginapi.Mount{
			ContainerPath: hostPath,
			HostPath:      cacheFileHostDirectory,
			ReadOnly:      false},
		&pluginapi.Mount{
			ContainerPath: "/tmp/vgpulock",
			HostPath:      "/tmp/vgpulock",
			ReadOnly:      false},
		&pluginapi.Mount{
			ContainerPath: "/etc/ld.so.preload",
			HostPath:      hostPath + "/ld.so.preload",
			ReadOnly:      true},
	)
}
