package stats

import (
	"ez-monitor/pkg/inventory"
	"fmt"
	"github.com/mitchellh/go-homedir"
	"golang.org/x/crypto/ssh"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type HostStats struct {
	Name        string
	Address     string
	CPUUsage    float64
	MemoryUsage float64
	MemoryTotal float64
	Error       error
	Timestamp   time.Time
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

func executeCommand(client *ssh.Client, command string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %s", err)
	}
	defer session.Close()
	output, err := session.CombinedOutput(command)
	if err != nil {
		return "", fmt.Errorf("failed to execute command %s: %s", command, err)
	}
	return string(output), nil
}

func getCPUUsage(client *ssh.Client) (float64, error) {
	command := "mpstat 1 1 | awk '$12 ~ /[0-9.]+/ {print 100 - $12}' | tail -1"

	alternativeCommand := "top -bn1 | grep '%Cpu' | awk '{print $2 + $4}'"

	output, err := executeCommand(client, command)
	if err != nil {
		// Try alternative command
		var errAlt error
		output, errAlt = executeCommand(client, alternativeCommand)
		if errAlt != nil {
			return 0, fmt.Errorf("failed to execute command %s: %s: failed to execute alternative command: %s: %s", command, err, alternativeCommand, errAlt)
		}
	}
	output = strings.TrimSpace(output)
	cpuUsage, err := strconv.ParseFloat(output, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse cpu usage: %s", err)
	}
	return cpuUsage, nil
}

func getMemoryUsage(client *ssh.Client) (used float64, total float64, err error) {
	command := "free -m | grep 'Mem:'"
	output, err := executeCommand(client, command)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to execute command %s: %s", command, err)
	}
	fields := strings.Fields(output)
	if len(fields) < 3 {
		return 0, 0, fmt.Errorf("unexpected output format from free command to get memory usage: %s", output)
	}

	totalMem, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse total memory usage: %s", err)
	}

	usedMem, err := strconv.ParseFloat(fields[2], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse used memory usage: %s", err)
	}

	return usedMem, totalMem, nil
}

func getHostStats(host inventory.Host) HostStats {
	stats := HostStats{
		Name:      host.Name,
		Address:   host.Address,
		Timestamp: time.Now(),
	}

	client, err := connectToHost(host)
	if err != nil {
		stats.Error = err
		return stats
	}
	defer client.Close()

	cpuUsage, err := getCPUUsage(client)
	if err != nil {
		stats.Error = err
		return stats
	}
	stats.CPUUsage = cpuUsage

	memUsed, memTotal, err := getMemoryUsage(client)
	if err != nil {
		stats.Error = err
		return stats
	}
	stats.MemoryUsage = memUsed
	stats.MemoryTotal = memTotal
	return stats
}

func CollectHostStats(hosts []inventory.Host) []HostStats {
	var wg sync.WaitGroup
	statsChan := make(chan HostStats, len(hosts))

	for _, host := range hosts {
		wg.Add(1)
		go func(host inventory.Host) {
			defer wg.Done()
			stat := getHostStats(host)
			statsChan <- stat
		}(host)
	}

	wg.Wait()
	close(statsChan)

	var results []HostStats
	for stats := range statsChan {
		results = append(results, stats)
	}

	return results
}
