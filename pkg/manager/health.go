package manager

import (
	"fmt"
	"strings"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"jinli.io/device-plugin/pkg/util"
	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

func (mgr *NvmlManager) CheckHealth(stop <-chan interface{}, unhealthy chan<- *pluginapi.Device) error {
	var devices []*pluginapi.Device
	for _, v := range mgr.Devs {
		devices = append(devices, &v.Device)
	}
	return mgr.checkHealth(stop, devices, unhealthy)
}

func (mgr *NvmlManager) checkHealth(stop <-chan interface{}, vDevices []*pluginapi.Device, unhealthy chan<- *pluginapi.Device) error {
	var devices []*pluginapi.Device
	var vuuids []string
	for _, v := range vDevices {
		vuuids = append(vuuids, v.ID)
	}
	uuids := util.GetUuids(vuuids)
	// 只要找到一个vDevice的拓扑信息，就可以获得对应的Device的拓扑信息
	for _, v := range uuids {
		for _, w := range vDevices {
			if !strings.Contains(w.ID, v) {
				continue
			}
			device := pluginapi.Device{
				ID:       v,
				Health:   w.Health,
				Topology: w.Topology,
			}
			devices = append(devices, &device)
			break
		}
	}

	applicationErrorXids := []uint64{
		13, // Graphics Engine Exception
		31, // GPU memory page fault
		43, // GPU stopped processing
		45, // Preemptive cleanup, due to previous errors
		68, // Video processor exception
	}

	skippedXids := make(map[uint64]bool)
	for _, id := range applicationErrorXids {
		skippedXids[id] = true
	}


	ret := mgr.nvml.Init()
	if ret != nvml.SUCCESS {
		return fmt.Errorf("error to initialize NVML: %v", ret)
	}
	defer func() {
		ret := mgr.nvml.Shutdown()
		if ret != nvml.SUCCESS {
			klog.Infof("Error shutting down NVML: %v", ret)
		}
	}()

	eventSet, ret := mgr.nvml.EventSetCreate()
	if ret != nvml.SUCCESS {
		return fmt.Errorf("failed to create event set: %v", ret)
	}
	defer func() {
		_ = eventSet.Free()
	}()

	parentToDeviceMap := make(map[string]*pluginapi.Device)

	eventMask := uint64(nvml.EventTypeXidCriticalError | nvml.EventTypeDoubleBitEccError | nvml.EventTypeSingleBitEccError)
	for _, d := range devices {
		uuid := d.GetID()
		parentToDeviceMap[uuid] = d

		gpu, ret := mgr.nvml.DeviceGetHandleByUUID(uuid)
		if ret != nvml.SUCCESS {
			klog.Infof("unable to get device handle from UUID: %v; marking it as unhealthy", ret)
			// d要转换为所有包含Device ID的vDevice
			// 用For循环发送出去
			for _, v := range vDevices {
				if d.ID == v.ID {
					unhealthy <- v
				}
			}
			continue
		}

		supportedEvents, ret := gpu.GetSupportedEventTypes()
		if ret != nvml.SUCCESS {
			klog.Infof("unable to determine the supported events for %v: %v; marking it as unhealthy", d.ID, ret)
			for _, v := range vDevices {
				if d.ID == v.ID {
					unhealthy <- v
				}
			}
			continue
		}

		ret = gpu.RegisterEvents(eventMask&supportedEvents, eventSet)
		if ret == nvml.ERROR_NOT_SUPPORTED {
			klog.Warningf("Device %v is too old to support healthchecking.", d.ID)
		}
		if ret != nvml.SUCCESS {
			klog.Infof("Marking device %v as unhealthy: %v", d.ID, ret)
			for _, v := range vDevices {
				if d.ID == v.ID {
					unhealthy <- v
				}
			}
		}
	}

	for {
		select {
		case <-stop:
			return nil
		default:
		}

		e, ret := eventSet.Wait(5000)
		if ret == nvml.ERROR_TIMEOUT {
			continue
		}
		if ret != nvml.SUCCESS {
			klog.Infof("Error waiting for event: %v; Marking all devices as unhealthy", ret)
			for _, d := range devices {
				for _, v := range vDevices {
					if d.ID == v.ID {
						unhealthy <- v
					}
				}
			}
			continue
		}

		if e.EventType != nvml.EventTypeXidCriticalError {
			klog.Infof("Skipping non-nvmlEventTypeXidCriticalError event: %+v", e)
			continue
		}

		if skippedXids[e.EventData] {
			klog.Infof("Skipping event %+v", e)
			continue
		}

		klog.Infof("Processing event %+v", e)
		eventUUID, ret := e.Device.GetUUID()
		if ret != nvml.SUCCESS {
			klog.Infof("Failed to determine uuid for event %v: %v; Marking all devices as unhealthy.", e, ret)
			for _, d := range devices {
				for _, v := range vDevices {
					if d.ID == v.ID {
						unhealthy <- v
					}
				}
			}
			continue
		}

		d, exists := parentToDeviceMap[eventUUID]
		if !exists {
			klog.Infof("Ignoring event for unexpected device: %v", eventUUID)
			continue
		}

		klog.Infof("XidCriticalError: Xid=%d on Device=%s; marking device as unhealthy.", e.EventData, d.ID)
		for _, v := range vDevices {
			if d.ID == v.ID {
				unhealthy <- v
			}
		}
	}
}
