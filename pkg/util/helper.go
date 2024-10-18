package util

import (
	"strings"

	set "github.com/deckarep/golang-set"
)

// 兼容time slice,形成的副版本号
// 主版本号::副版本号
func GetUuids(all []string) []string {
	var uuids []string
	for _, v := range all {
		split := strings.SplitN(string(v), "::", 2)
		if len(split) != 2 {
			uuids = append(uuids, v)
		} else {
			uuids = append(uuids, split[0])
		}
	}
	return RemoveDuplicates(uuids)
}

func RemoveDuplicates(slice []string) []string {
	mySet := set.NewSet()
	var result []string
	for _, item := range slice {
		if mySet.Contains(item) {
			continue
		}
		mySet.Add(item)
		result = append(result, item)
	}
	return result
}

func Int8SliceToString(s []int8) string {
	var b []byte
	for _, c := range s {
		if c == 0 {
			break
		}
		b = append(b, byte(c))
	}
	return string(b)
}
