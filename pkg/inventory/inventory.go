package inventory

import (
	"fmt"
	"gopkg.in/ini.v1"
	"os"
	"strconv"
)

type Host struct {
	Name              string
	Username          string
	Password          string
	Address           string
	Port              int
	SshPrivateKeyFile string
}

func LoadInventory(filename string) ([]Host, error) {
	f, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cfg, err := ini.Load(f)
	if err != nil {
		return nil, fmt.Errorf("failed to load ini data: %s", err)
	}

	hostMap := make(map[string]Host)
	for _, section := range cfg.Sections() {
		hostname := section.Name()
		if hostname == ini.DefaultSection {
			for _, key := range section.Keys() {
				return nil, fmt.Errorf("variables %s=%s must be defined under a host section", key.Name(), key.Value())
			}
			continue
		}
		if _, ok := hostMap[hostname]; ok {
			return nil, fmt.Errorf("duplicate host section: %s", hostname)
		}

		host := Host{
			Name: hostname,
		}
		for _, key := range section.Keys() {
			switch key.Name() {
			case "username":
				host.Username = key.Value()
			case "password":
				host.Password = key.Value()
			case "address":
				host.Address = key.Value()
			case "port":
				host.Port, err = strconv.Atoi(key.Value())
				if err != nil {
					return nil, fmt.Errorf("invalid port in section %s: %s", hostname, err)
				}
			case "ssh_private_key_file":
				host.SshPrivateKeyFile = key.Value()
			default:
				return nil, fmt.Errorf("unknown variable %s for host %s", key.Name(), hostname)
			}
		}
		hostMap[hostname] = host
	}

	hostList := make([]Host, 0, len(hostMap))
	for _, host := range hostMap {
		hostList = append(hostList, host)
	}

	return hostList, nil
}
