package persistance

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"liti0s/litios/lightkeeper/config"
	"liti0s/litios/lightkeeper/deployment"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"docker.io/go-docker/api/types/network"

	"github.com/docker/go-connections/nat"

	"docker.io/go-docker"
	"docker.io/go-docker/api/types"
	"docker.io/go-docker/api/types/container"
	"docker.io/go-docker/api/types/filters"
	"docker.io/go-docker/api/types/mount"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

type Volume struct {
	Name    string
	SrcPath string
	DstPath string
}

type Bind struct {
	SrcPath string
	DstPath string
}

const originPath string = "/home/litios/Projects/lightkeeper/backups"
const restorePath string = "/tmp/lightkeeper/"

//TODO CHECK IF RESTOREPATH EXISTS

// StoreFromContainer persists the container data in the mounts
// It retrieves the data and generates the tars in the originPath folder
// The backups are marked with the date and the container name
// Name has the form: type--path.tar
func StoreFromContainer(container types.Container) bool {
	cli, err := docker.NewEnvClient()
	checkErr(err)

	destPath := originPath + container.Names[0] + "/" + time.Now().Format("02-01-2006") + "/"
	os.MkdirAll(destPath, os.ModePerm)

	for _, mount := range container.Mounts {
		path := mount.Destination
		data, _, err := cli.CopyFromContainer(context.Background(), container.ID[:10], path)
		defer data.Close()
		checkErr(err)

		buf, err := ioutil.ReadAll(data)
		checkErr(err)

		f, err := os.Create(destPath + strings.Replace(string(mount.Type)+"--"+path[1:]+".tar", "/", "-", -1))
		checkErr(err)

		n2, err := f.Write(buf)
		checkErr(err)

		fmt.Printf("wrote %d bytes\n", n2)
	}
	return true
}

// GetContainerMounts returns a list of Mount which contains all the container mounts according to the backups
func GetContainerMounts(containerName string, date string, config config.Container) (binds []Bind, volumes []Volume) {
	target := originPath + containerName + "/" + date + "/"

	_, err := os.Stat(target)
	if os.IsNotExist(err) {
		return
	}

	files, err := ioutil.ReadDir(target)
	checkErr(err)

	for _, file := range files {
		mountType := strings.Split(file.Name(), "--")[0]
		dstPath := strings.Replace("/"+strings.Split(file.Name(), "--")[1], "-", "/", -1)

		if mountType == "volume" {
			volumeName := ""
			for _, mount := range config.Mounts {
				if mount.DstPath == dstPath[0:len(dstPath)-4] {
					volumeName = mount.From
				}
			}

			if volumeName == "" {
				panic("Volume not found")
			}

			volumes = append(volumes, Volume{
				DstPath: dstPath[0 : len(dstPath)-4],
				SrcPath: target + file.Name(),
				Name:    volumeName})
		} else {
			binds = append(binds, Bind{
				DstPath: dstPath[0 : len(dstPath)-4],
				SrcPath: target + file.Name()})
		}
	}

	return
}

func RecreateMounts(binds []Bind, volumes []Volume, allVolumes []*types.Volume) []mount.Mount {
	cli, err := docker.NewEnvClient()
	checkErr(err)

	createdMounts := []mount.Mount{}
	err = os.RemoveAll(restorePath + "tmp")
	checkErr(err)

	for _, containerVolume := range volumes {
		err := os.Mkdir(restorePath+"tmp", 0755)
		checkErr(err)

		for _, volume := range allVolumes {
			if containerVolume.DstPath == volume.Mountpoint {
				fmt.Println("Removing volume", volume.Name)
				err = cli.VolumeRemove(context.Background(), volume.Name, true)
				checkErr(err)
			}
		}
		checkErr(Untartar(containerVolume.SrcPath, restorePath+"tmp"))
		deployment.CreateVolume(containerVolume.Name, restorePath+"tmp")
		err = os.RemoveAll(restorePath + "tmp")
		checkErr(err)

		createdMounts = append(createdMounts, mount.Mount{Type: mount.Type("volume"), Source: containerVolume.Name, Target: containerVolume.DstPath})

	}

	for _, bind := range binds {
		bindLocalPath := strings.Replace(bind.DstPath[1:len(bind.DstPath)], "/", "-", -1)
		targetPath := restorePath + bindLocalPath
		err = os.RemoveAll(targetPath)
		checkErr(err)

		err := os.Mkdir(targetPath, 0755)
		checkErr(err)

		checkErr(Untartar(bind.SrcPath, targetPath))

		createdMounts = append(createdMounts, mount.Mount{Type: mount.Type("bind"), Source: targetPath, Target: bind.DstPath})
	}

	return createdMounts
}

// RecoverContainer is the main function for recovery
// It stops the container, cleans the already created volumes and recreates the container
// The date specifies the selected backup to be recovered
func RecoverContainer(containerName string, date string) types.Container {
	cli, err := docker.NewEnvClient()
	checkErr(err)

	allConfig := config.ConfigData
	var containerConfig config.Container

	for _, config := range allConfig.Containers {
		if "/"+config.Name == containerName {
			containerConfig = config
			break
		}
	}

	if containerConfig.Name == "" {
		panic("Container config not found")
	}

	containerBinds, containerVolumes := GetContainerMounts(containerName, date, containerConfig)
	// If the container is running, stop it and delete it
	if containerName != "" {
		containers := deployment.GetContainers()
		for _, container := range containers {
			if container.Names[0] == containerName {
				deployment.StopContainer(containerName)
				deployment.RemoveContainer(containerName)
			}
		}
	}

	args := filters.NewArgs(filters.Arg("name", "name="+containerName))
	volumes, err := cli.VolumeList(context.Background(), args)
	mounts := RecreateMounts(containerBinds, containerVolumes, volumes.Volumes)

	ports := nat.PortMap{}
	for hostPort, containerPort := range containerConfig.Ports {
		hostBinding := nat.PortBinding{
			HostIP:   "0.0.0.0",
			HostPort: strconv.Itoa(hostPort),
		}
		containerPort, err := nat.NewPort(strings.Split(containerPort, "/")[1], strings.Split(containerPort, "/")[0])
		checkErr(err)

		ports[containerPort] = []nat.PortBinding{hostBinding}
	}
	container := deployment.LaunchContainer(containerName, container.Config{Image: containerConfig.Image, Env: containerConfig.Env}, container.HostConfig{Mounts: mounts, PortBindings: ports}, network.NetworkingConfig{})

	return container
}

// Untartar deals with the process of untaring the backups
// This function preserve the permissions and ownership of the files
// CAUTION: In order to work, this code must be run with sudo permissions
func Untartar(tarName, xpath string) (err error) {
	tarFile, err := os.Open(tarName)

	defer tarFile.Close()
	absPath, err := filepath.Abs(xpath)

	tr := tar.NewReader(tarFile)

	// untar each segment
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		checkErr(err)

		// determine proper file path info
		finfo := hdr.FileInfo()
		fileName := hdr.Name
		absFileName := filepath.Join(absPath, fileName)
		// if a dir, create it, then go to next segment
		if finfo.Mode().IsDir() {
			if err := os.MkdirAll(absFileName, 0755); err != nil {
				return err
			}
			continue
		}
		// create new file with original file mode
		file, err := os.OpenFile(
			absFileName,
			os.O_RDWR|os.O_CREATE|os.O_TRUNC,
			finfo.Mode().Perm(),
		)
		checkErr(err)

		err = file.Chown(hdr.Uid, hdr.Gid)
		checkErr(err)

		n, cpErr := io.Copy(file, tr)
		if closeErr := file.Close(); closeErr != nil {
			return err
		}
		checkErr(cpErr)

		if n != finfo.Size() {
			return fmt.Errorf("wrote %d, want %d", n, finfo.Size())
		}
	}
	return nil
}
