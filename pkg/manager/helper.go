package manager

import (
	"strconv"
	"strings"

	"k8s.io/klog/v2"
)

type int8Slice []int8

// String turns a nil terminated int8Slice into a string
func (s int8Slice) String() string {
	var b []byte
	for _, c := range s {
		if c == 0 {
			break
		}
		b = append(b, byte(c))
	}
	return string(b)
}

func getAdditionalXids(input string) []uint64 {
	if input == "" {
		return nil
	}

	var additionalXids []uint64
	for _, additionalXid := range strings.Split(input, ",") {
		trimmed := strings.TrimSpace(additionalXid)
		if trimmed == "" {
			continue
		}
		xid, err := strconv.ParseUint(trimmed, 10, 64)
		if err != nil {
			klog.Infof("Ignoring malformed Xid value %v: %v", trimmed, err)
			continue
		}
		additionalXids = append(additionalXids, xid)
	}

	return additionalXids
}
