package deployment

import (
	"context"
	"liti0s/litios/lightkeeper/persistance"

	"docker.io/go-docker/api/types"
	"docker.io/go-docker/api/types/mount"
	"docker.io/go-docker/api/types/network"
	"docker.io/go-docker/api/types/volume"

	"docker.io/go-docker"
	"docker.io/go-docker/api/types/container"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func CreateVolume(name string, from string) {
	cli, err := docker.NewEnvClient()
	checkErr(err)

	_, err = cli.VolumeCreate(context.Background(), volume.VolumesCreateBody{Driver: "overlay2", Name: name})
	checkErr(err)

	randomName := "7122b3717buassdiuh1"

	mounts := []mount.Mount{}
	mounts = append(mounts, mount.Mount{Type: mount.Type("volume"), Source: name, Target: "/tmp/volume"})
	mounts = append(mounts, mount.Mount{Type: mount.Type("bind"), Source: from, Target: "/tmp/bind"})

	container := LaunchContainer(randomName, container.Config{Image: "alpine"}, container.HostConfig{Mounts: mounts}, network.NetworkingConfig{})
	cli.ContainerExecCreate(context.Background(), container.ID, types.ExecConfig{User: "root", Cmd: []string{"cp", "-rp", "/tmp/bind", "/tmp/volume"}})

	// TODO tick to check if copy finished
}

func LaunchContainer(containerName string, containerConfig container.Config, hostConfig container.HostConfig, networkingConfig network.NetworkingConfig) types.Container {
	cli, err := docker.NewEnvClient()
	checkErr(err)

	result, err := cli.ContainerCreate(context.Background(), &containerConfig, &hostConfig, &networkingConfig, containerName)
	checkErr(err)

	containers := persistance.GetContainers()
	for _, container := range containers {
		if container.ID == result.ID {
			return container
		}
	}

	panic("Container not running")
}
