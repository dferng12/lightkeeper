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
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

const originPath string = "./backups/"

func GetContainers() []types.Container {
	cli, err := docker.NewEnvClient()
	checkErr(err)

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	checkErr(err)

	return containers
}

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
