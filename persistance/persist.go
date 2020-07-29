package persistance

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"docker.io/go-docker"
	"docker.io/go-docker/api/types"
	"docker.io/go-docker/api/types/filters"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

// Mount holds the info for the mounts of a container
type Mount struct {
	Type        string // Mount or bind
	Path        string
	DockerMount types.Volume // Real docker struct
}

const originPath string = "./backups/"

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

// GetContainers returns a list of the current running containers
func GetContainers() (containers []types.Container) {
	cli, err := docker.NewEnvClient()
	checkErr(err)

	containers, err = cli.ContainerList(context.Background(), types.ContainerListOptions{})
	checkErr(err)

	return
}

// GetContainerMounts returns a list of Mount which contains all the container mounts according to the backups
func GetContainerMounts(containerName string, date string) (mounts []Mount) {
	target := originPath + "/" + date + "/"

	_, err := os.Stat(target)
	if os.IsNotExist(err) {
		return
	}

	files, err := ioutil.ReadDir(target)
	checkErr(err)

	for _, file := range files {
		mounts = append(mounts, Mount{
			Type: strings.Split(file.Name(), "--")[0],
			Path: strings.Replace("/"+strings.Split(file.Name(), "--")[1], "-", "/", -1)})
	}

	return
}

// RecoverContainer is the main function for recovery
// It stops the container, cleans the already created volumes and recreates the container
// The date specifies the selected backup to be recovered
func RecoverContainer(containerName string, date string) {
	cli, err := docker.NewEnvClient()
	checkErr(err)

	containerBackupMounts := GetContainerMounts(containerName, date)
	args := filters.NewArgs(filters.Arg("name", "name="+containerName))
	volumes, err := cli.VolumeList(context.Background(), args)

	// If the container is running, stop it and delete it
	if containerName != "" {
		containers := GetContainers()
		duration := time.Duration(5000000000)
		for _, container := range containers {
			if container.Names[0] == containerName {
				err = cli.ContainerStop(context.Background(), containerName, &duration)
				checkErr(err)
				err = cli.ContainerRemove(context.Background(), containerName, types.ContainerRemoveOptions{RemoveVolumes: true})
				checkErr(err)

			}
		}
	} else { // else, check that volumes doesn't exist
		for _, mount := range containerBackupMounts {
			if mount.Type == "volume" {
				for _, volume := range volumes.Volumes {
					if mount.Path == volume.Mountpoint {
						err = cli.VolumeRemove(context.Background(), volume.Name, true)
						checkErr(err)
					}
				}
			}
		}
	}
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
