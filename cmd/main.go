package main

import (
	"os/exec"

	"github.com/kubevirt/device-plugin-manager/pkg/dpm"
	"jinli.io/device-plugin/pkg/server"
	"k8s.io/klog/v2"
)

func main() {
	l := server.PluginLister{
		ResUpdateChan: make(chan dpm.PluginNameList),
	}

	// 创建manager
	mgr := dpm.NewManager(&l)

	go func() {
		if _, err := exec.Command("nvidia-smi").Output(); err == nil {
			l.ResUpdateChan <- []string{"gpu"}
		}
	}()

	// 运行manager
	// manager包括注册等与kubelet交互服务
	klog.Info("Jinli Device PLugin Start Run!")
	mgr.Run()
}
