package statistics

import (
	"errors"
	"fmt"
	"github.com/kreulenk/ez-monitor/pkg/inventory"
	"github.com/mitchellh/go-homedir"
	"golang.org/x/crypto/ssh"
	"os"
	"sync"
	"time"
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

func getAuthMethod(host inventory.Host) (ssh.AuthMethod, error) {
	if host.Password != "" {
		return ssh.Password(host.Password), nil
	}

	if host.SshPrivateKeyFile != "" {
		dir, err := homedir.Expand(host.SshPrivateKeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to expand SSH private key file %s: %s", host.SshPrivateKeyFile, err)
		}

		key, err := os.ReadFile(dir)
		if err != nil {
			return nil, fmt.Errorf("failed to read SSH private key file %s: %s", dir, err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("failed to parse SSH private key file %s: %s", dir, err)
		}
		return ssh.PublicKeys(signer), nil
	}
	return nil, fmt.Errorf("either password or ssh_private_key_file must be specified for each host")
}

func connectToHost(host inventory.Host) (*ssh.Client, error) {
	authMethod, err := getAuthMethod(host)
	if err != nil {
		return nil, err
	}

	sshConfig := &ssh.ClientConfig{
		User: host.Username,
		Auth: []ssh.AuthMethod{
			authMethod,
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Probably switch to ssh.FixedHostKey or ssh.KnownHosts
		Timeout:         time.Second * 10,
	}
	port := 22
	if host.Port != 0 {
		port = host.Port
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host.Address, port), sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %s", host.Address, err)
	}

	return client, nil
}
