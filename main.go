package main

import (
	"ez-monitor/inventory"
	"fmt"
	"os"
)

func main() {
	f, err := os.ReadFile("./test/inv-complete.yml")
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

	fmt.Println(hosts)
}
