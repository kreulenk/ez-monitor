package main

import (
	"ez-monitor/inventory"
	"fmt"
	"os"
)

func main() {
	f, err := os.ReadFile("./test/groups-no-all.yml")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	inv, err := inventory.LoadInventory(f)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	hosts := inv.GetHosts()

	fmt.Println("The following hosts were found in the inventory:")
	for hostName, hostVars := range hosts {
		fmt.Printf("%s\n", hostName)
		for k, v := range hostVars.Variables {
			fmt.Printf("\t%s: %v\n", k, v)
		}
	}
}
