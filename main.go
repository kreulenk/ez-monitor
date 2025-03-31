package main

import (
	"ez-monitor/inventory"
	"fmt"
	"os"
	"strings"
)

func main() {
	filename := "./test/ungrouped-host.ini"
	f, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	inv, err := inventory.LoadInventory(f, strings.Split(filename, ".")[len(strings.Split(filename, "."))-1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	hosts := inv.GetHosts()

	fmt.Println("The following hosts were found in the inventory:")
	for hostName, hostVars := range hosts {
		fmt.Printf("%s\n", hostName)
		if hostVars != nil {
			for k, v := range hostVars.Variables {
				fmt.Printf("\t%s: %v\n", k, v)
			}
		}
	}
}
