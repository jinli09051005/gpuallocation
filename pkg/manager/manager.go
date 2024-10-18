package manager

import (
	"context"
	"fmt"
	"os"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"jinli.io/device-plugin/pkg/util"
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
	devs, gpumems, err := getVDevices(nvmllib)
	if err != nil {
		return nil, fmt.Errorf("error to get devices: %v", err)
	}

	// 添加GPU显存注解
	// jinli.io/gpumems=uuid1-1024,uuid2-2048
	nodename := os.Getenv("NODE_NAME")
	err = util.UpdateCurrentNode(context.TODO(), nodename, gpumems)
	if err != nil {
		return nil, fmt.Errorf("error to add gpumemes annotations: %v", err)
	}

	mgr := &NvmlManager{
		Devs: devs,
		nvml: nvmllib,
	}

	return mgr, nil
}
