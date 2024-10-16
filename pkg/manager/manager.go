package manager

import (
	"fmt"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"k8s.io/klog/v2"
)

type NvmlManager struct {
	Devs []*Device
	nvml nvml.Interface
}

func NewNvmlManagers(nvmllib nvml.Interface) (*NvmlManager, error) {
	ret := nvmllib.Init()
	if ret != nvml.SUCCESS {
		return nil, fmt.Errorf("error to initialize NVML: %v", ret)
	}
	defer func() {
		ret := nvmllib.Shutdown()
		if ret != nvml.SUCCESS {
			klog.Infof("Error shutting down NVML: %v", ret)
		}
	}()

	// 获取设备信息
	devs, err := getDevices(nvmllib)
	if err != nil {
		return nil, fmt.Errorf("error to get devices: %v", err)
	}

	mgr := &NvmlManager{
		Devs: devs,
		nvml: nvmllib,
	}

	return mgr, nil
}
