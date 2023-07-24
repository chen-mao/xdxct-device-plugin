package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"path"
	"strings"
	"time"

	"google.golang.org/grpc"

	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const (
	resourceName           = "xdxct.com/gpu"
	serverSock             = pluginapi.DevicePluginPath + "xdxct.sock"
	envDisableHealthChecks = "DP_DISABLE_HEALTHCHECKS"
	allHealthChecks        = "xids"
)

// XdxctDevicePlugin implements the Kubernetes device plugin API
type XdxctDevicePlugin struct {
	devs   []*pluginapi.Device
	socket string

	stop   chan interface{}
	health chan *pluginapi.Device

	server *grpc.Server
}

// NewXdxctDevicePlugin returns an initialized XdxctDevicePlugin
func NewXdxctDevicePlugin() *XdxctDevicePlugin {
	return &XdxctDevicePlugin{
		devs:   GetDevices(),
		socket: serverSock,

		stop:   make(chan interface{}),
		health: make(chan *pluginapi.Device),
	}
}

func (m *XdxctDevicePlugin) GetDevicePluginOptions(context.Context, *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{}, nil
}

// dial establishes the gRPC communication with the registered device plugin.
func (plugin *XdxctDevicePlugin) dial(unixSocketPath string, timeout time.Duration) (*grpc.ClientConn, error) {
	c, err := grpc.Dial(unixSocketPath, grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithTimeout(timeout),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}),
	)

	if err != nil {
		return nil, err
	}

	return c, nil
}

func (plugin *XdxctDevicePlugin) Register() error {
	conn, err := plugin.dial(pluginapi.KubeletSocket, 5*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pluginapi.NewRegistrationClient(conn)
	req := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     path.Base(plugin.socket),
		ResourceName: string(resourceName),
		Options: &pluginapi.DevicePluginOptions{
			GetPreferredAllocationAvailable: true,
		},
	}

	_, err = client.Register(context.Background(), req)
	if err != nil {
		return err
	}
	return nil
}

func (plugin *XdxctDevicePlugin) initialize() {
	plugin.server = grpc.NewServer([]grpc.ServerOption{}...)
	plugin.stop = make(chan interface{})
	plugin.health = make(chan *pluginapi.Device)
}

// Start starts the gRPC server of the device plugin
func (plugin *XdxctDevicePlugin) Start() error {
	plugin.initialize()
	err := plugin.Serve()
	if err != nil {
		klog.Info("Could not start device plugin")
		plugin.cleanup()
		return err
	}

	klog.Info("Starting to serve %s", plugin.socket)

	err = plugin.Register()
	if err != nil {
		klog.Infof("could not registry device plugin: %s", err)
		plugin.Stop()
		return err
	}
	klog.Info("Registered device plugin for %s with Kubelet", resourceName)

	return nil
}

// Stop stops the gRPC server
func (plugin *XdxctDevicePlugin) Stop() error {
	if plugin == nil || plugin.server == nil {
		return nil
	}

	klog.Infof("Stopping to serve '%s'", plugin.socket)
	plugin.server.Stop()
	if err := os.Remove(plugin.socket); err != nil && !os.IsNotExist(err) {
		return err
	}
	plugin.cleanup()
	return nil
}

// ListAndWatch lists devices and update that list according to the health status
func (plugin *XdxctDevicePlugin) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
	s.Send(&pluginapi.ListAndWatchResponse{Devices: plugin.devs})

	for {
		select {
		case <-plugin.stop:
			return nil
		}
	}
}

// Allocate which return list of devices.
func (plugin *XdxctDevicePlugin) Allocate(ctx context.Context, reqs *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	devs := plugin.devs
	respones := pluginapi.AllocateResponse{}

	for _, req := range reqs.ContainerRequests {
		response := pluginapi.ContainerAllocateResponse{
			Envs: map[string]string{
				"NVIDIA_VISIBLE_DEVICES": strings.Join(req.DevicesIDs, ","),
			},
		}

		for _, id := range req.DevicesIDs {
			if !deviceExists(devs, id) {
				return nil, fmt.Errorf("invalid allocation request: unknown devices: %s", id)
			}
		}

		respones.ContainerResponses = append(respones.ContainerResponses, &response)
	}

	return &respones, nil
}

func (m *XdxctDevicePlugin) PreStartContainer(context.Context, *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	return &pluginapi.PreStartContainerResponse{}, nil
}

func (plugin *XdxctDevicePlugin) cleanup() {
	close(plugin.stop)
	plugin.server = nil
	plugin.health = nil
	plugin.stop = nil
}

// GetPreferredAllocation returns the preferred allocation from the set of devices specified in the request
func (plugin *XdxctDevicePlugin) GetPreferredAllocation(ctx context.Context, r *pluginapi.PreferredAllocationRequest) (*pluginapi.PreferredAllocationResponse, error) {
	return &pluginapi.PreferredAllocationResponse{}, nil
}

// Serve starts the gRPC server and register the device plugin to Kubelet
func (plugin *XdxctDevicePlugin) Serve() error {

	os.Remove(plugin.socket)
	sock, err := net.Listen("unix", plugin.socket)
	if err != nil {
		return err
	}

	pluginapi.RegisterDevicePluginServer(plugin.server, plugin)
	go func() {
		lastCrashTime := time.Now()
		restartCount := 0
		for {
			err := plugin.server.Serve(sock)
			if err == nil {
				break
			}

			if restartCount > 5 {
				klog.Fatal("GRPC server has repeatedly crashed")
			}

			timeSinceLastCrash := time.Since(lastCrashTime).Seconds()
			lastCrashTime = time.Now()

			if timeSinceLastCrash > 3600 {
				restartCount = 1
			} else {
				restartCount++
			}
		}
	}()

	conn, err := plugin.dial(plugin.socket, 5*time.Second)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}
