package deployment

import (
	"context"
	"fmt"
	"time"

	"docker.io/go-docker/api/types"
	"docker.io/go-docker/api/types/mount"
	"docker.io/go-docker/api/types/network"
	"docker.io/go-docker/api/types/strslice"
	"docker.io/go-docker/api/types/volume"

	"docker.io/go-docker"
	"docker.io/go-docker/api/types/container"
)

func IsContainerRunning(containerName string) bool {
	returnValue := true
	defer func() {
		if err := recover(); err != nil {
			returnValue = false
		}
	}()

	GetContainer(containerName)
	return returnValue
}

// GetContainers returns a list of the current running containers
func GetContainer(containerName string) (container types.Container) {
	cli, err := docker.NewEnvClient()
	checkErr(err)

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	checkErr(err)

	for _, container := range containers {
		if container.Names[0] == "/"+containerName {
			return container
		}
	}

	panic("No such container " + containerName)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func CreateVolume(name string, from string) types.Volume {
	fmt.Println("Creating volume:", name, "from path", from)
	cli, err := docker.NewEnvClient()
	checkErr(err)

	vol, err := cli.VolumeCreate(context.Background(), volume.VolumesCreateBody{Driver: "local", Name: name})
	checkErr(err)

	randomName := "7122b3717buassdiuh1"

	mounts := []mount.Mount{}
	mounts = append(mounts, mount.Mount{Type: mount.Type("volume"), Source: name, Target: "/tmp/volume"})
	mounts = append(mounts, mount.Mount{Type: mount.Type("bind"), Source: from, Target: "/tmp/bind"})

	container := LaunchContainer(randomName, container.Config{Image: "alpine:latest", Cmd: strslice.StrSlice([]string{"tail", "-f", "/dev/null"})}, container.HostConfig{Mounts: mounts}, network.NetworkingConfig{})
	ID, err := cli.ContainerExecCreate(context.Background(), container.ID, types.ExecConfig{User: "root", Cmd: []string{"cp", "-rp", "/tmp/bind/.", "/tmp/volume"}})
	checkErr(err)
	err = cli.ContainerExecStart(context.Background(), ID.ID, types.ExecStartCheck{})
	checkErr(err)

	fmt.Print("Copying data to the volume...")
	for {
		result, err := cli.ContainerExecInspect(context.Background(), ID.ID)
		checkErr(err)

		if !result.Running {
			fmt.Println()
			break
		}
		fmt.Print(".")
		time.Sleep(250 * time.Millisecond)
	}

	StopContainer(randomName)
	RemoveContainer(randomName)

	return vol
}

func StopContainer(containerName string) {
	fmt.Println("Stopping container", containerName)
	cli, err := docker.NewEnvClient()
	checkErr(err)

	duration := time.Duration(5 * time.Second)
	err = cli.ContainerStop(context.Background(), containerName, &duration)
	checkErr(err)
	fmt.Println("Container stopped")
}

func RemoveContainer(containerName string) {
	fmt.Println("Removing container", containerName)
	cli, err := docker.NewEnvClient()
	checkErr(err)

	err = cli.ContainerRemove(context.Background(), containerName, types.ContainerRemoveOptions{})
	checkErr(err)
	fmt.Println("Container removed")
}

func LaunchContainer(containerName string, containerConfig container.Config, hostConfig container.HostConfig, networkingConfig network.NetworkingConfig) types.Container {
	fmt.Println("Launching container", containerName)
	cli, err := docker.NewEnvClient()
	checkErr(err)

	result, err := cli.ContainerCreate(context.Background(), &containerConfig, &hostConfig, &networkingConfig, containerName)
	checkErr(err)

	err = cli.ContainerStart(context.Background(), result.ID, types.ContainerStartOptions{})
	checkErr(err)

	fmt.Println("Container started")

	return GetContainer(containerName)
}
