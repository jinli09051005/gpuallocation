package main

import "fmt"

func main() {
	devices := []demo{
		{ID: "a"},
		{ID: "b"},
	}

	var vDevices []demo
	physicalDevNum := len(devices)
	for i := 0; i < 100; i++ {
		index := i
		if i >= physicalDevNum {
			index = i % physicalDevNum
		}
		vDevcice := devices[index]
		vDevcice.ID = fmt.Sprintf("%s::%d", vDevcice.ID, i)
		vDevices = append(vDevices, vDevcice)
	}
	fmt.Println(vDevices)
}

type demo struct {
	ID string
}
