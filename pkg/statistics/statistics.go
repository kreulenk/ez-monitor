package statistics

import (
	"context"
	"fmt"
	"github.com/kreulenk/ez-monitor/pkg/inventory"
	"github.com/mitchellh/go-homedir"
	"golang.org/x/crypto/ssh"
	"os"
	"strconv"
	"strings"
	"time"
)

type HostStats struct {
	NameOfHost string // The term hostname makes naming this field difficult...
	Address    string

	CPUUsage float64
	CPUError error

	MemoryUsage float64
	MemoryTotal float64
	MemoryError error

	DiskUsage float64
	DiskTotal float64
	DiskError error

	Timestamp time.Time
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

func getDiskUsage(client *ssh.Client) (used float64, total float64, err error) {
	command := "df -m --output=used,size / | tail -1"
	output, err := executeCommand(client, command)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to execute command %s: %s", command, err)
	}
	fields := strings.Fields(output)
	if len(fields) < 2 {
		return 0, 0, fmt.Errorf("unexpected output format from df command to get disk usage: %s", output)
	}

	usedDisk, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse used disk space: %s", err)
	}

	totalDisk, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse total disk space: %s", err)
	}

	return usedDisk, totalDisk, nil
}

func getHostStats(host ConnectionInfo) HostStats {
	stats := HostStats{
		NameOfHost: host.InventoryInfo.Name,
		Address:    host.InventoryInfo.Address,
		Timestamp:  time.Now(),
	}

	cpuUsage, err := getCPUUsage(host.connectionClient)
	if err != nil {
		stats.CPUError = err
	}
	stats.CPUUsage = cpuUsage

	memUsed, memTotal, err := getMemoryUsage(host.connectionClient)
	if err != nil {
		stats.MemoryError = err
	}
	stats.MemoryUsage = memUsed
	stats.MemoryTotal = memTotal

	diskUsage, diskTotal, err := getDiskUsage(host.connectionClient)
	if err != nil {
		stats.DiskError = err
	}

	stats.DiskUsage = diskUsage
	stats.DiskTotal = diskTotal

	return stats
}

func StartStatisticsCollection(ctx context.Context, inventoryInfo []inventory.Host) (chan HostStats, error) {
	hosts, err := connectToHosts(inventoryInfo) // We close the connections when the context cancels in the loop below
	if err != nil {
		return nil, fmt.Errorf("failed to connect to hosts: %s", err)
	}

	statsChan := make(chan HostStats)
	for _, host := range hosts {
		go func(host ConnectionInfo, statsChan chan HostStats) {
			stat := getHostStats(host)
			statsChan <- stat
			ticker := time.NewTicker(time.Second * 2)
			for {
				select {
				case <-ticker.C:
					stat := getHostStats(host)
					statsChan <- stat
				case <-ctx.Done():
					host.connectionClient.Close()
					return
				}
			}
		}(host, statsChan)
	}
	return statsChan, nil
}
