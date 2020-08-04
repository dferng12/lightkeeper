package main

import (
	"fmt"
	"os"
	"time"

	"github.com/dferng12/lightkeeper/persistance"

	"github.com/robfig/cron"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage:", os.Args[0], "action")
		fmt.Println("Available actions:")
		fmt.Println("- backup (containername): Create backup")
		fmt.Println("- restore (containername) (date): Restore container from date (mandatory)")
		fmt.Println("\tDate is in format: DD-MM-YYYY")
		fmt.Println("- schedule (cron): Schedule a cron job to periodically backup all")
		return
	}

	if os.Args[1] == "backup" {
		persistance.StoreFromContainer(os.Args[2])
	} else if os.Args[1] == "restore" {
		persistance.RecoverContainer(os.Args[2], os.Args[3])
	} else if os.Args[1] == "schedule" {
		c := cron.New()
		c.AddFunc(os.Args[2], persistance.StoreAllFromConfig)
		c.Start()
		time.Sleep(1000 * time.Second)
	}
}
