package manager

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

type Device struct {
	pluginapi.Device
	Index            string
	TotalMemory      uint64
	ComputeCapbility string
	Paths            []string
}

// nvidia.go
func getDevices(nvmllib nvml.Interface) ([]*Device, error) {
	count, ret := nvmllib.DeviceGetCount()
	if ret != nvml.SUCCESS {
		return nil, fmt.Errorf("error getting device count: %v", ret)
	}

	var devs []*Device
	for i := 0; i < count; i++ {
		var dev Device

		device, ret := nvmllib.DeviceGetHandleByIndex(i)
		if ret != nvml.SUCCESS {
			return nil, fmt.Errorf("error getting device handle for index '%v': %v", i, ret)
		}

		index := fmt.Sprintf("%v", i)

		paths, err := getPaths(device)
		if err != nil {
			return nil, fmt.Errorf("error getting device paths: %v", err)
		}

		computeCapability, err := getComputeCapability(device)
		if err != nil {
			return nil, fmt.Errorf("error getting device compute capability: %w", err)
		}

		totalMemory, err := getTotalMemery(device)
		if err != nil {
			return nil, fmt.Errorf("error getting device memory: %w", err)
		}

		uuid, ret := device.GetUUID()
		if ret != nvml.SUCCESS {
			return nil, fmt.Errorf("error getting device uuid for index '%v': %v", i, ret)
		}

		pluginapiDevs := pluginapi.Device{
			ID:     uuid,
			Health: pluginapi.Healthy,
		}

		hasNuma, numa, err := getNumaNode(device)
		if err != nil {
			return nil, fmt.Errorf("error getting device NUMA node: %v", err)
		}
		if hasNuma {
			pluginapiDevs.Topology = &pluginapi.TopologyInfo{
				Nodes: []*pluginapi.NUMANode{
					{
						ID: int64(numa),
					},
				},
			}
		}

		dev.Device = pluginapiDevs
		dev.Index = index
		dev.TotalMemory = totalMemory
		dev.ComputeCapbility = computeCapability
		dev.Paths = paths
		devs = append(devs, &dev)
	}

	return devs, nil
}

func getPaths(device nvml.Device) ([]string, error) {
	minor, ret := device.GetMinorNumber()
	if ret != nvml.SUCCESS {
		return nil, fmt.Errorf("error getting GPU device minor number: %v", ret)
	}
	path := fmt.Sprintf("/dev/nvidia%d", minor)

	return []string{path}, nil
}

func getComputeCapability(device nvml.Device) (string, error) {
	major, minor, ret := device.GetCudaComputeCapability()
	if ret != nvml.SUCCESS {
		return "", ret
	}
	return fmt.Sprintf("%d.%d", major, minor), nil
}

func getTotalMemery(device nvml.Device) (uint64, error) {
	info, ret := device.GetMemoryInfo()
	if ret != nvml.SUCCESS {
		return 0, ret
	}
	return info.Total, nil
}

func getNumaNode(device nvml.Device) (bool, int, error) {
	info, ret := device.GetPciInfo()
	if ret != nvml.SUCCESS {
		return false, 0, fmt.Errorf("error getting PCI Bus Info of device: %v", ret)
	}

	// Discard leading zeros.
	busID := strings.ToLower(strings.TrimPrefix(int8Slice(info.BusId[:]).String(), "0000"))

	b, err := os.ReadFile(fmt.Sprintf("/sys/bus/pci/devices/%s/numa_node", busID))
	if err != nil {
		return false, 0, nil
	}

	node, err := strconv.Atoi(string(bytes.TrimSpace(b)))
	if err != nil {
		return false, 0, fmt.Errorf("eror parsing value for NUMA node: %v", err)
	}

	if node < 0 {
		return false, 0, nil
	}

	return true, node, nil
}
