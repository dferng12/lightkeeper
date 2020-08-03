package main

import (
	"fmt"
	"liti0s/litios/lightkeeper/persistance"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage:", os.Args[0], "action containername")
		fmt.Println("Available actions:")
		fmt.Println("- backup: Create backup")
		fmt.Println("- restore: Restore container from date (mandatory)")
		return
	}

	if os.Args[1] == "backup" {
		persistance.StoreFromContainer(os.Args[2])
	} else if os.Args[1] == "restore" {
		persistance.RecoverContainer(os.Args[2], os.Args[3])
	}
}
