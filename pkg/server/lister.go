package server

import (
	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/kubevirt/device-plugin-manager/pkg/dpm"
	"jinli.io/device-plugin/pkg/manager"
	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

type PluginLister struct {
	ResUpdateChan chan dpm.PluginNameList
}

func (l *PluginLister) GetResourceNamespace() string {
	return resnamespace
}

func (l *PluginLister) Discover(pluginListCh chan dpm.PluginNameList) {
	for {
		select {
		case resList := <-l.ResUpdateChan:
			pluginListCh <- resList
		case <-pluginListCh:
			return
		}
	}
}

func (l *PluginLister) NewPlugin(resName string) dpm.PluginInterface {
	nvmllib := nvml.New()
	nvmlmgr, err := manager.NewNvmlManagers(nvmllib)
	if err != nil {
		klog.Errorf("failed to create nvmlmanager,err: %v", err)
		return nil
	}
	return &Plugin{
		nvmlmgr: *nvmlmgr,
		stop:    make(chan interface{}),
		health:  make(chan *pluginapi.Device),
	}
}
