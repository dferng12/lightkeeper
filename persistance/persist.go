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

	"docker.io/go-docker"
	"docker.io/go-docker/api/types"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func GetMounts() []types.MountPoint {
	cli, err := docker.NewEnvClient()
	if err != nil {
		panic(err)
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	mountList := []types.MountPoint{}

	for _, container := range containers {
		for _, mount := range container.Mounts {
			mountList = append(mountList, mount)
		}
	}
	return mountList
}

func StoreFromContainer(container types.Container, path string) bool {
	cli, err := docker.NewEnvClient()
	checkErr(err)

	data, _, err := cli.CopyFromContainer(context.Background(), container.ID[:10], path)
	defer data.Close()
	checkErr(err)

	buf, err := ioutil.ReadAll(data)
	checkErr(err)

	f, err := os.Create(strings.Replace(path[1:]+".tar", "/", "-", -1))
	checkErr(err)

	n2, err := f.Write(buf)
	checkErr(err)

	fmt.Printf("wrote %d bytes\n", n2)
	return true
}

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
