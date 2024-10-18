package main

import (
	"os/exec"

	"github.com/kubevirt/device-plugin-manager/pkg/dpm"
	"jinli.io/device-plugin/pkg/server"
	"k8s.io/klog/v2"
)

func main() {
	// 初始化插件列表对象
	l := server.PluginLister{
		ResUpdateChan: make(chan dpm.PluginNameList),
	}

	// 创建manager，并关联插件列表对象
	mgr := dpm.NewManager(&l)

	go func() {
		// 判断节点是否有NVIDIA GPU
		if _, err := exec.Command("nvidia-smi").Output(); err == nil {
			// 设备插件名称列表
			// socket -> jinli.io_gpu.sock
			// 节点有gpu，注册设备插件
			l.ResUpdateChan <- []string{"gpu"}
		}
	}()

	// 运行manager
	// manager包括注册等与kubelet交互服务
	klog.Info("Jinli Device PLugin Start Run!")
	mgr.Run()
}
