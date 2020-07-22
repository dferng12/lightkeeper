package main

import (
	"liti0s/litios/lightkeeper/persistance"
)

func main() {
	containers := persistance.GetContainers()
	persistance.StoreFromContainer(containers[0])

	//persistance.Untartar("var-log.tar", "./")
}
