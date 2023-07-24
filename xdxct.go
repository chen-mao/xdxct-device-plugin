package main

import (
	"fmt"
	"log"
	"strconv"

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
		device, ret := xdxml.DeviceGetHandleByIndex(i)
		check(ret)

		ID, ret := device.GetMinorNumber()
		if ret != xdxml.SUCCESS {
			log.Fatalf("Unable to get id of device at index %v: %v", ID, ret)
		}
		fmt.Printf("ID: %v\n", ID)

		devs = append(devs, &pluginapi.Device{
			ID:     strconv.Itoa(ID),
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
