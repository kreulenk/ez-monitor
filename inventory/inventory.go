package inventory

import (
	"fmt"
	"gopkg.in/ini.v1"
	"gopkg.in/yaml.v3"
	"maps"
)

// Variable can be applied to either a host or a group

type Host struct {
	Variables map[string]interface{} `yaml:",inline"`
}

type Group struct {
	Hosts     map[string]*Host       `yaml:"hosts,omitempty"`
	Variables map[string]interface{} `yaml:"vars,omitempty"`
	Children  map[string]*Group      `yaml:"children,omitempty"`
}

// Inventory represents the complete inventory
type Inventory struct {
	All       *Group `yaml:"all"`
	Ungrouped *Group `yaml:"ungrouped,omitempty"`
}

func LoadInventory(data []byte, format string) (*Inventory, error) {
	switch format {
	case "yaml", "yml":
		return loadYamlInventory(data)
	case "ini":
		return loadIniInventory(data)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

func loadIniInventory(iniData []byte) (*Inventory, error) {
	cfg, err := ini.LoadSources(ini.LoadOptions{AllowBooleanKeys: true}, iniData)
	if err != nil {
		return nil, fmt.Errorf("failed to load ini data: %s", err)
	}

	inv := &Inventory{
		All: &Group{
			Hosts:     make(map[string]*Host),
			Variables: make(map[string]interface{}),
			Children:  make(map[string]*Group),
		},
		Ungrouped: &Group{
			Hosts:     make(map[string]*Host),
			Variables: make(map[string]interface{}),
		},
	}

	for _, section := range cfg.Sections() {
		if section.Name() == ini.DefaultSection {
			for _, key := range section.Keys() {
				inv.Ungrouped.Hosts[key.Name()] = &Host{
					Variables: make(map[string]interface{}),
				}
			}
		} else {
			group := &Group{
				Hosts:     make(map[string]*Host),
				Variables: make(map[string]interface{}),
			}
			for _, key := range section.Keys() {
				group.Hosts[key.Name()] = &Host{
					Variables: make(map[string]interface{}),
				}
			}
			inv.All.Children[section.Name()] = group
		}
	}

	return inv, nil
}

func loadYamlInventory(yamlData []byte) (*Inventory, error) {
	// First attempt to unmarshal assuming the inv file starts with the all: group
	inv := &Inventory{}
	err := yaml.Unmarshal(yamlData, inv)
	if err != nil {
		return nil, err
	}
	if len(inv.GetHosts()) != 0 {
		return inv, nil
	}

	// Second attempt to unmarshal assuming the inv file starts with groups
	group := map[string]*Group{}
	err = yaml.Unmarshal(yamlData, &group)
	if err != nil {
		return nil, err
	}
	inv = &Inventory{All: &Group{Children: group}}
	if len(inv.GetHosts()) != 0 {
		return inv, nil
	}
	return nil, fmt.Errorf("no hosts found in inventory")
}

func (i *Inventory) GetHosts() map[string]*Host {
	if i.All == nil {
		return nil
	}
	allHosts := maps.Clone(i.All.Hosts)
	groupVars := maps.Clone(i.All.Variables) // Track the groupVars that will be passed down

	if allHosts == nil {
		allHosts = make(map[string]*Host)
	}
	if groupVars == nil {
		groupVars = make(map[string]interface{})
	}

	// Add all inventory
	for hostKey, hostVal := range allHosts { // Iterate over top level hosts
		for groupVarKey, groupVarVal := range groupVars { // Iterate over global vars
			if hostVal.Variables != nil {
				if hostVal.Variables == nil {
					allHosts[hostKey].Variables = map[string]interface{}{groupVarKey: groupVarVal}
				} else if _, ok := hostVal.Variables[groupVarKey]; !ok { // Set the group var if it is not overridden by a host var
					allHosts[hostKey].Variables[groupVarKey] = groupVarVal
				}
			}
		}
	}
	for _, group := range i.All.Children {
		getHostsHelper(allHosts, maps.Clone(groupVars), group)
	}

	return allHosts
}

func getHostsHelper(allHosts map[string]*Host, varsSoFar map[string]interface{}, currGroup *Group) {
	for hostKey, hostVal := range currGroup.Hosts {
		allHosts[hostKey] = mergeHostVars(allHosts[hostKey], hostVal)
		varsSoFar = mergeVars(varsSoFar, currGroup.Variables)
		for groupVarKey, groupVarVal := range varsSoFar { // Iterate over global vars
			varsSoFar[groupVarKey] = groupVarVal
			if hostVal.Variables == nil {
				allHosts[hostKey].Variables = map[string]interface{}{groupVarKey: groupVarVal}
			} else if _, ok := hostVal.Variables[groupVarKey]; !ok { // Set the group var if it is not overridden by a host var
				allHosts[hostKey].Variables[groupVarKey] = groupVarVal
			}
		}
	}
	for _, group := range currGroup.Children {
		getHostsHelper(allHosts, maps.Clone(varsSoFar), group)
	}
}

func mergeHostVars(hostAlreadyFoundInInventory *Host, newlyFoundHost *Host) *Host {
	if hostAlreadyFoundInInventory == nil {
		return newlyFoundHost
	}

	for k, v := range newlyFoundHost.Variables {
		hostAlreadyFoundInInventory.Variables[k] = v
	}
	return hostAlreadyFoundInInventory
}

func mergeVars(groupVarsAlreadyFoundInInventory map[string]interface{}, newGroupVars map[string]interface{}) map[string]interface{} {
	for k, v := range newGroupVars {
		groupVarsAlreadyFoundInInventory[k] = v
	}
	return groupVarsAlreadyFoundInInventory
}
