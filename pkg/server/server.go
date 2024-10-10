package server

import (
	"context"
	"fmt"
	"strings"

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
	responses := pluginapi.AllocateResponse{}
	for _, req := range reqs.ContainerRequests {
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
