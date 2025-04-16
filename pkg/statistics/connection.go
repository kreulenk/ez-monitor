package statistics

import (
	"errors"
	"ez-monitor/pkg/inventory"
	"golang.org/x/crypto/ssh"
	"sync"
)

type ConnectionInfo struct {
	InventoryInfo    inventory.Host
	connectionClient *ssh.Client
}

func connectToHosts(inventoryInfo []inventory.Host) ([]ConnectionInfo, error) {
	var wg sync.WaitGroup

	errChan := make(chan error, len(inventoryInfo))
	connChan := make(chan ConnectionInfo, len(inventoryInfo))

	for _, host := range inventoryInfo {
		wg.Add(1)
		go func(host inventory.Host) {
			defer wg.Done()
			client, err := connectToHost(host)
			if err != nil {
				errChan <- err
				return
			}
			connChan <- ConnectionInfo{
				InventoryInfo:    host,
				connectionClient: client,
			}
		}(host)
	}
	wg.Wait()
	close(errChan)
	close(connChan)

	if len(errChan) != 0 {
		var errs []error
		for e := range errChan {
			errs = append(errs, e)
		}
		return nil, errors.Join(errs...)
	}

	var hosts []ConnectionInfo
	for host := range connChan {
		hosts = append(hosts, host)
	}

	return hosts, nil
}
