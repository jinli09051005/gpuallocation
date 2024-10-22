package server

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/kubevirt/device-plugin-manager/pkg/dpm"
	"jinli.io/device-plugin/pkg/manager"
	"jinli.io/device-plugin/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	cdiapi "tags.cncf.io/container-device-interface/pkg/cdi"
)

type Plugin struct {
	nvmlmgr manager.NvmlManager
	health  chan *pluginapi.Device
	stop    chan interface{}
}

var _ dpm.PluginInterfaceStart = &Plugin{}
var _ dpm.PluginInterfaceStop = &Plugin{}
var _ dpm.PluginInterface = &Plugin{}

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
	nodename := os.Getenv("NODE_NAME")
	current, err := util.GetCurrentPod(ctx, nodename)
	if err != nil {
		return &pluginapi.AllocateResponse{}, err
	}

	responses := pluginapi.AllocateResponse{}
	for idx, req := range reqs.ContainerRequests {
		// req代表单个容器，要处理pod中每个容器
		response := pluginapi.ContainerAllocateResponse{}
		// 将vgpu ID -> physical gpu ID
		req.DevicesIDs = util.GetUuids(req.DevicesIDs)
		// 1、环境变量形式，非CDI模式，需要nvidia-container-runtime
		// response.Envs = map[string]string{
		// 	"NVIDIA_VISIBLE_DEVICES": strings.Join(req.DevicesIDs, ","),
		// }

		// 2、注解形式, CDI模式
		// response.Annotations["cdi.k8s.io/nvidia-device-plugin_uuid值"] = "nvidia.com/gpu=uuid,nvidia.com/gds=all,nvidia.com/mofed=all"
		responseID := uuid.New().String()
		key := "cdi.k8s.io/nvidia-device-plugin_" + responseID
		values := []string{}
		for _, id := range req.DevicesIDs {
			v := "nvidia.com/gpu=" + id
			values = append(values, v)
		}
		valueStr, err := cdiapi.AnnotationValue(values)
		if err != nil {
			return nil, fmt.Errorf("CDI annotation failed: %w", err)
		}
		response.Annotations = map[string]string{
			key: valueStr,
		}

		// 3、CDIDevice形式，CDI模式
		// cdidevices := []*pluginapi.CDIDevice{}
		// for _, id := range req.DevicesIDs {
		// 	cdidevices = append(cdidevices, &pluginapi.CDIDevice{
		// 		Name: fmt.Sprintf("nvidia.com/gpu=%s", id),
		// 	})
		// }
		// response.CDIDevices = cdidevices

		// 判断请求的ID是否有效
		for _, id := range req.DevicesIDs {
			if !manager.DeviceExists(p.nvmlmgr.Devs, id) {
				return nil, fmt.Errorf("error to get allocate response: unknown device: %s", id)
			}
		}

		// gpu资源限制
		os.MkdirAll("/tmp/vgpulock", 0777)
		os.Chmod("/tmp/vgpulock", 0777)
		util.LimitGPUMemAndCores(&response, current, int32(idx))

		// 更新Pod环境变量
		// env["UUID"] = "uuid1,uuid2"
		env := corev1.EnvVar{
			Name:  "UUID",
			Value: valueStr,
		}
		current.Spec.Containers[idx].Env = append(current.Spec.Containers[idx].Env, env)
		responses.ContainerResponses = append(responses.ContainerResponses, &response)
	}

	// 更新Pod分配状态及环境变量
	err = util.UpdateCurrentPod(ctx, current)
	if err != nil {
		return &pluginapi.AllocateResponse{}, fmt.Errorf("error to add allocateStatus annotations: %v", err)
	}
	return &responses, nil
}
