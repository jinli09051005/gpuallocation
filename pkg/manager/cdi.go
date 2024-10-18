package manager

import (
	"fmt"
	"path/filepath"

	nvdevice "github.com/NVIDIA/go-nvlib/pkg/nvlib/device"
	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/NVIDIA/nvidia-container-toolkit/pkg/nvcdi"
	roottransform "github.com/NVIDIA/nvidia-container-toolkit/pkg/nvcdi/transform/root"
	"github.com/sirupsen/logrus"
	cdiapi "tags.cncf.io/container-device-interface/pkg/cdi"
)

const (
	cdiRoot = "/var/run/cdi"
)

// 物理GPU CDI
func (mgr *NvmlManager) CreateCDISpecFile() error {
	mgrlog := logrus.StandardLogger()
	driverRoot := "/"
	targetDriverRoot := driverRoot
	deviceNamer, err := nvcdi.NewDeviceNamer(nvcdi.DeviceNameStrategyUUID)
	if err != nil {
		return err
	}
	vendor := "nvidia.com"
	cdilibs := make(map[string]nvcdi.Interface)
	cdilibs["gpu"], err = nvcdi.New(
		nvcdi.WithLogger(mgrlog),
		nvcdi.WithNvmlLib(mgr.nvml),
		nvcdi.WithDeviceLib(nvdevice.New(mgr.nvml)),
		nvcdi.WithDriverRoot(driverRoot),
		nvcdi.WithDeviceNamers(deviceNamer),
		nvcdi.WithVendor(vendor),
		nvcdi.WithClass("gpu"),
	)
	if err != nil {
		return fmt.Errorf("failed to create nvcdi library: %v", err)
	}

	for class, cdilib := range cdilibs {
		mgrlog.Infof("Generating CDI spec for resource: %s/%s", vendor, class)

		if class == "gpu" {
			ret := mgr.nvml.Init()
			if ret != nvml.SUCCESS {
				return fmt.Errorf("failed to initialize NVML: %v", ret)
			}
			defer mgr.nvml.Shutdown()
		}

		spec, err := cdilib.GetSpec()
		if err != nil {
			return fmt.Errorf("failed to get CDI spec: %v", err)
		}

		err = roottransform.New(
			roottransform.WithRoot(driverRoot),
			roottransform.WithTargetRoot(targetDriverRoot),
		).Transform(spec.Raw())
		if err != nil {
			return fmt.Errorf("failed to transform driver root in CDI spec: %v", err)
		}

		raw := spec.Raw()
		specName, err := cdiapi.GenerateNameForSpec(raw)
		if err != nil {
			return fmt.Errorf("failed to generate spec name: %v", err)
		}

		err = spec.Save(filepath.Join(cdiRoot, specName+".json"))
		if err != nil {
			return fmt.Errorf("failed to save CDI spec: %v", err)
		}
	}

	return nil
}
