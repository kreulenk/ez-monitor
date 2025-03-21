package inventory

import (
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
	All Group `yaml:"all"`
}

func LoadInventory(yamlData []byte) (*Inventory, error) {
	inv := &Inventory{}
	err := yaml.Unmarshal(yamlData, inv)
	if err != nil {
		return nil, err
	}
	return inv, nil
}

func (i *Inventory) GetHosts() map[string]*Host {
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
