package main

import (
	"fmt"
	"log"

	xdxml "github.com/chen-mao/go-xdxml/pkg/xdxml"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

func check(ret xdxml.Return) {
	if ret != xdxml.SUCCESS {
		log.Panicln("Fatal:", ret)
	}
}

func GetDevices() []*pluginapi.Device {
	n, ret := xdxml.DeviceGetCount()
	check(ret)

	var devs []*pluginapi.Device
	for i := 0; i < n; i++ {
		var device xdxml.Device

		ret := xdxml.DeviceGetHandleByIndex(i, &device)
		check(ret)

		ID, ret := device.GetUUID()
		if ret != xdxml.SUCCESS {
			log.Fatalf("Unable to get id of device at index %v: %v", ID, ret)
		}
		fmt.Printf("ID: %v\n", ID)

		devs = append(devs, &pluginapi.Device{
			ID:     "0x" + ID[:8],
			Health: pluginapi.Healthy,
		})
	}

	return devs
}

func deviceExists(devs []*pluginapi.Device, id string) bool {
	for _, d := range devs {
		if d.ID == id {
			return true
		}
	}
	return false
}
