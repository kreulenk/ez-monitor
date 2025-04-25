package inventory

import (
	"fmt"
	"golang.org/x/term"
	"gopkg.in/ini.v1"
	"os"
	"strconv"
	"strings"
)

type Host struct {
	Alias             string
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

	// If we see an encrypted password we will prompt the user for this value and then save it to this var for further passwords
	var encPassword string

	hostMap := make(map[string]Host)
	for _, section := range cfg.Sections() {
		hostAlias := section.Name()
		if hostAlias == ini.DefaultSection {
			for _, key := range section.Keys() {
				return nil, fmt.Errorf("variables %s=%s must be defined under a host section", key.Name(), key.Value())
			}
			continue
		}
		if _, ok := hostMap[hostAlias]; ok {
			return nil, fmt.Errorf("duplicate host section: %s", hostAlias)
		}

		host := Host{
			Alias: hostAlias,
		}
		for _, key := range section.Keys() {
			switch key.Name() {
			case "username":
				host.Username = key.Value()
			case "password":
				if strings.HasPrefix(key.Value(), ezMonitorEncDelimiter) {
					if encPassword == "" {
						fmt.Println("Please enter your encryption password to decrypt the passwords in this file.")
						encPasswordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
						fmt.Println() // Print a newline after password input
						if err != nil {
							return nil, fmt.Errorf("failed to read encryption password: %s", err)
						}
						encPassword = string(encPasswordBytes)
					}
					host.Password, err = decrypt(key.Value(), encPassword)
					if err != nil {
						return nil, fmt.Errorf("failed to decrypt password: %s", err)
					}
				} else {
					host.Password = key.Value()
				}
			case "address":
				host.Address = key.Value()
			case "port":
				host.Port, err = strconv.Atoi(key.Value())
				if err != nil {
					return nil, fmt.Errorf("invalid port in section %s: %s", hostAlias, err)
				}
			case "ssh_private_key_file":
				host.SshPrivateKeyFile = key.Value()
			default:
				return nil, fmt.Errorf("unknown variable %s for host %s", key.Name(), hostAlias)
			}
		}
		hostMap[hostAlias] = host
	}

	hostList := make([]Host, 0, len(hostMap))
	for _, host := range hostMap {
		hostList = append(hostList, host)
	}

	return hostList, nil
}
