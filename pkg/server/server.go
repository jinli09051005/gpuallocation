package server

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"jinli.io/device-plugin/pkg/manager"
	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const (
	resnamespace = "jinli.io"
	resname      = "gpu"
)

type Plugin struct {
	nvmlmgr manager.NvmlManager
	health  chan *pluginapi.Device
	stop    chan interface{}
}

// manager自动调用
func (p *Plugin) Start() error {
	// 健康检查
	go func() {
		err := p.nvmlmgr.CheckHealth(p.stop, p.health)
		if err != nil {
			klog.Infof("Failed to start health check: %v; continuing with health checks disabled", err)
		}
	}()
	return nil
}

// manager自动调用
func (p *Plugin) Stop() error {
	close(p.stop)
	return nil
}

func (p *Plugin) GetDevicePluginOptions(ctx context.Context, e *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{
		PreStartRequired:                true,
		GetPreferredAllocationAvailable: true,
	}, nil
}

func (p *Plugin) PreStartContainer(ctx context.Context, r *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	return &pluginapi.PreStartContainerResponse{}, nil
}

func (p *Plugin) GetPreferredAllocation(ctx context.Context, r *pluginapi.PreferredAllocationRequest) (*pluginapi.PreferredAllocationResponse, error) {
	return &pluginapi.PreferredAllocationResponse{}, nil
}

func (p *Plugin) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
	var devices []*pluginapi.Device
	for _, v := range p.nvmlmgr.Devs {
		devices = append(devices, &v.Device)
	}
	if err := s.Send(&pluginapi.ListAndWatchResponse{Devices: devices}); err != nil {
		return err
	}

	for {
		select {
		case <-p.stop:
			return nil
		case dev := <-p.health:
			dev.Health = pluginapi.Unhealthy
			klog.Infof("%s:%s device marked unhealthy", fmt.Sprintf(resnamespace+"/"+resname), dev.ID)
			if err := s.Send(&pluginapi.ListAndWatchResponse{Devices: devices}); err != nil {
				return nil
			}
		}
	}
}

func (p *Plugin) Allocate(ctx context.Context, reqs *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	hostPath := "/usr/local/vgpu"
	responses := pluginapi.AllocateResponse{}
	for _, req := range reqs.ContainerRequests {
		// req代表单个容器，要处理单个pod所有容器
		response := pluginapi.ContainerAllocateResponse{
			Envs: map[string]string{
				"NVIDIA_VISIBLE_DEVICES": strings.Join(req.DevicesIDs, ","),
			},
		}

		for _, id := range req.DevicesIDs {
			if !deviceExists(p.nvmlmgr.Devs, id) {
				return nil, fmt.Errorf("error to get allocate response: unknown device: %s", id)
			}
		}

		// gpu资源限制
		for i := range req.DevicesIDs {
			memKey := fmt.Sprintf("CUDA_DEVICE_MEMORY_LIMIT_%v", i)
			response.Envs[memKey] = "20m"
		}
		//HAMI-core中CUDA_DEVICE_MEMORY_LIMIT_ID（限制容器指定设备显存）会覆盖CUDA_DEVICE_MEMORY_LIMIT（限制容器所有设备显存）
		response.Envs["CUDA_DEVICE_MEMORY_LIMIT_0"] = "20m"
		response.Envs["CUDA_DEVICE_SM_LIMIT"] = "50"
		response.Envs["CUDA_DEVICE_MEMORY_SHARED_CACHE"] = fmt.Sprintf("%s/%v.cache", hostPath, uuid.New().String())
		response.Envs["CUDA_OVERSUBSCRIBE"] = "true"

		cacheFileHostDirectory := "/usr/local/vgpu/containers/poduid_containername"
		os.RemoveAll(cacheFileHostDirectory)

		os.MkdirAll(cacheFileHostDirectory, 0777)
		os.Chmod(cacheFileHostDirectory, 0777)
		os.MkdirAll("/tmp/vgpulock", 0777)
		os.Chmod("/tmp/vgpulock", 0777)

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

		responses.ContainerResponses = append(responses.ContainerResponses, &response)
	}

	return &responses, nil
}

func deviceExists(devs []*manager.Device, id string) bool {
	for _, v := range devs {
		if v.ID != id {
			continue
		}
		return true
	}
	return false
}
