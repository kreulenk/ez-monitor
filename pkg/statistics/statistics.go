package statistics

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kreulenk/ez-monitor/pkg/inventory"
	"golang.org/x/crypto/ssh"
	"strconv"
	"strings"
	"time"
)

type HistoricalDataPoint struct {
	Data      float64
	Timestamp time.Time
}

type HostStat struct {
	HostAlias string
	Address   string

	CPUUsage float64
	CPUError error

	MemoryUsage float64
	MemoryTotal float64
	MemoryError error

	DiskUsage float64
	DiskTotal float64
	DiskError error

	NetworkingMBReceived float64
	NetworkingMBSent     float64
	NetworkingError      error

	Timestamp time.Time
}

func StartStatisticsCollection(ctx context.Context, inventoryInfo []inventory.Host) (chan *HostStat, error) {
	hosts, err := connectToHosts(inventoryInfo) // We close the connections when the context cancels in the loop below
	if err != nil {
		return nil, fmt.Errorf("failed to connect to hosts: %s", err)
	}

	statsChan := make(chan *HostStat)
	for _, host := range hosts {
		go func(host ConnectionInfo, statsChan chan *HostStat) {
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

func getCPUUsage(client *ssh.Client) (float64, error) {
	command := "mpstat 1 1 | awk '$12 ~ /[0-9.]+/ {print 100 - $12}' | tail -1"

	alternativeCommand := "top -bn2 -d1 | grep '%Cpu' | tail -1 | awk '{print $2 + $4}'"

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

	total, err = strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse total memory usage: %s", err)
	}

	used, err = strconv.ParseFloat(fields[2], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse used memory usage: %s", err)
	}

	return used, total, nil
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

	used, err = strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse used disk space: %s", err)
	}

	total, err = strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse total disk space: %s", err)
	}

	return used, total, nil
}

func getNetworkingUsage(client *ssh.Client) (sent float64, received float64, err error) {
	command := "ip -s link show | awk '/^[0-9]+: / {iface=$2} iface!=\"lo:\" && $1 ~ /^[0-9]+$/ {rx+=$1; getline; tx+=$1} END {print rx, tx}'"
	output, err := executeCommand(client, command)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to execute command %s: %s", command, err)
	}
	fields := strings.Fields(output)
	if len(fields) < 2 {
		return 0, 0, fmt.Errorf("unexpected output format from df command to get disk usage: %s", output)
	}

	received, err = strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse networking received MB: %s", err)
	}

	sent, err = strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse networking sent MB: %s", err)
	}

	return sent / 1024 / 1024, received / 1024 / 1024, nil
}

func getHostStats(host ConnectionInfo) *HostStat {
	stats := &HostStat{
		HostAlias: host.InventoryInfo.Alias,
		Address:   host.InventoryInfo.Address,
		Timestamp: time.Now(),
	}

	stats.CPUUsage, stats.CPUError = getCPUUsage(host.connectionClient)
	stats.MemoryUsage, stats.MemoryTotal, stats.MemoryError = getMemoryUsage(host.connectionClient)
	stats.DiskUsage, stats.DiskTotal, stats.DiskError = getDiskUsage(host.connectionClient)
	stats.NetworkingMBSent, stats.NetworkingMBReceived, stats.NetworkingError = getNetworkingUsage(host.connectionClient)

	return stats
}

func executeCommand(client *ssh.Client, command string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %s", err)
	}
	defer session.Close()
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr
	err = session.Run(command)
	if err != nil || stderr.Len() > 0 {
		return "", fmt.Errorf("failed to execute command %s: %s", command, err)
	}
	return stdout.String(), nil
}
