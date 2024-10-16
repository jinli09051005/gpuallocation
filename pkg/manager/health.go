package manager

import (
	"fmt"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
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

func (mgr *NvmlManager) checkHealth(stop <-chan interface{}, devices []*pluginapi.Device, unhealthy chan<- *pluginapi.Device) error {
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

	eventSet, ret := mgr.nvml.EventSetCreate()
	if ret != nvml.SUCCESS {
		return fmt.Errorf("failed to create event set: %v", ret)
	}
	defer func() {
		_ = eventSet.Free()
	}()

	for _, additionalXid := range getAdditionalXids("xids") {
		skippedXids[additionalXid] = true
	}

	parentToDeviceMap := make(map[string]*pluginapi.Device)

	eventMask := uint64(nvml.EventTypeXidCriticalError | nvml.EventTypeDoubleBitEccError | nvml.EventTypeSingleBitEccError)
	for _, d := range devices {
		uuid := d.GetID()
		parentToDeviceMap[uuid] = d

		gpu, ret := mgr.nvml.DeviceGetHandleByUUID(uuid)
		if ret != nvml.SUCCESS {
			klog.Infof("unable to get device handle from UUID: %v; marking it as unhealthy", ret)
			unhealthy <- d
			continue
		}

		supportedEvents, ret := gpu.GetSupportedEventTypes()
		if ret != nvml.SUCCESS {
			klog.Infof("unable to determine the supported events for %v: %v; marking it as unhealthy", d.ID, ret)
			unhealthy <- d
			continue
		}

		ret = gpu.RegisterEvents(eventMask&supportedEvents, eventSet)
		if ret == nvml.ERROR_NOT_SUPPORTED {
			klog.Warningf("Device %v is too old to support healthchecking.", d.ID)
		}
		if ret != nvml.SUCCESS {
			klog.Infof("Marking device %v as unhealthy: %v", d.ID, ret)
			unhealthy <- d
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
				unhealthy <- d
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
				unhealthy <- d
			}
			continue
		}

		d, exists := parentToDeviceMap[eventUUID]
		if !exists {
			klog.Infof("Ignoring event for unexpected device: %v", eventUUID)
			continue
		}

		klog.Infof("XidCriticalError: Xid=%d on Device=%s; marking device as unhealthy.", e.EventData, d.ID)
		unhealthy <- d
	}
}
