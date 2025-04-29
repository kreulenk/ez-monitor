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
	InventoryInfo     inventory.Host
	connectionClient  *ssh.Client
	connectionSession *ssh.Session
}

func connectToHosts(inventoryInfo []inventory.Host) ([]ConnectionInfo, error) {
	var wg sync.WaitGroup

	errChan := make(chan error, len(inventoryInfo))
	connChan := make(chan ConnectionInfo, len(inventoryInfo))

	for _, host := range inventoryInfo {
		wg.Add(1)
		go func(host inventory.Host) {
			defer wg.Done()
			client, session, err := connectToHost(host)
			if err != nil {
				errChan <- err
				return
			}
			connChan <- ConnectionInfo{
				InventoryInfo:     host,
				connectionClient:  client,
				connectionSession: session,
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

func getAuthMethods(host inventory.Host) ([]ssh.AuthMethod, error) {
	var authMethods []ssh.AuthMethod

	if host.Password != "" {
		authMethods = append(authMethods, ssh.Password(host.Password))
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
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	return authMethods, nil
}

func connectToHost(host inventory.Host) (*ssh.Client, *ssh.Session, error) {
	authMethods, err := getAuthMethods(host)
	if err != nil {
		return nil, nil, err
	}

	sshConfig := &ssh.ClientConfig{
		User:            host.Username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Probably switch to ssh.FixedHostKey or ssh.KnownHosts
		Timeout:         time.Second * 10,
	}
	port := 22
	if host.Port != 0 {
		port = host.Port
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host.Address, port), sshConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to %s: %s", host.Alias, err)
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, nil, fmt.Errorf("failed to open session on %s: %s", host.Address, err)
	}

	return client, session, nil
}
