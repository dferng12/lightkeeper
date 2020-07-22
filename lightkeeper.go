package main

import (
	"liti0s/litios/lightkeeper/persistance"
)

func main() {
	containers := persistance.GetContainers()
	persistance.StoreFromContainer(containers[0], "/var/log")

	persistance.Untartar("var-log.tar", "./")
}