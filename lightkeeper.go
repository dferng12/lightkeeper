package main

import (
	"liti0s/litios/lightkeeper/deployment"
	"liti0s/litios/lightkeeper/persistance"
)

func main() {
	//fmt.Println(config)
	//deployment.CreateVolume("test", "/home/litios/test")
	containers := deployment.GetContainers()
	//persistance.StoreFromContainer(containers[0])

	persistance.RecoverContainer(containers[0].Names[0], "31-07-2020")
	//persistance.Untartar("var-log.tar", "./")
}
