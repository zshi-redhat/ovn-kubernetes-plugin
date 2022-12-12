package main

import (
	"errors"
	"fmt"
	"net"
	"os"

	"github.com/golang/glog"
	"gopkg.in/yaml.v2"
)

type OVNKubernetesPlugin struct {
	Name    string
	Version string
}

const (
	pluginVersion  = "0.1"
	ConfigFileName = "ovn.yaml"
)

var Plugin OVNKubernetesPlugin

type OVNKubernetesConfig struct {
	// Configuration for microshift-ovs-init.service
	OVSInit OVSInit `json:"ovsInit,omitempty"`
	// MTU to use for the geneve tunnel interface.
	// This must be 100 bytes smaller than the uplink mtu.
	// Default is 1400.
	MTU uint32 `json:"mtu,omitempty"`
}

type OVSInit struct {
	// disable microshift-ovs-init.service.
	// OVS bridge "br-ex" needs to be configured manually when disableOVSInit is true.
	DisableOVSInit bool `json:"disableOVSInit,omitempty"`
	// Uplink interface for OVS bridge "br-ex"
	GatewayInterface string `json:"gatewayInterface,omitempty"`
	// Uplink interface for OVS bridge "br-ex1"
	ExternalGatewayInterface string `json:"externalGatewayInterface,omitempty"`
}

func init() {
	Plugin = OVNKubernetesPlugin{
		Name:    "ovn-kubernetes",
		Version: pluginVersion,
	}
}

func (o *OVNKubernetesConfig) ValidateOVSBridge(bridge string) error {
	_, err := net.InterfaceByName(bridge)
	if err != nil {
		return err
	}
	return nil
}

func (o *OVNKubernetesConfig) withDefaults() *OVNKubernetesConfig {
	o.OVSInit.DisableOVSInit = false
	o.MTU = 1400
	return o
}

func newOVNKubernetesConfigFromFile(path string) (*OVNKubernetesConfig, error) {
	o := new(OVNKubernetesConfig)
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(buf, &o)
	if err != nil {
		return nil, fmt.Errorf("parsing OVNKubernetes config: %v", err)
	}
	return o, nil
}

func NewOVNKubernetesConfigFromFileOrDefault(path string) (*OVNKubernetesConfig, error) {
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			glog.Info("OVNKubernetes config file not found, assuming default values")
			return new(OVNKubernetesConfig).withDefaults(), nil
		}
		return nil, fmt.Errorf("failed to get OVNKubernetes config file: %v", err)
	}

	o, err := newOVNKubernetesConfigFromFile(path)
	if err == nil {
		glog.Info("got OVNKubernetes config from file %q", path)
		return o, nil
	}
	return nil, fmt.Errorf("getting OVNKubernetes config: %v", err)
}

func (p *OVNKubernetesPlugin) Name() string {
	return p.Name
}

func (p *OVNKubernetesPlugin) Version() string {
	return p.Version
}

func (p *OVNKubernetesPlugin) GetManifests() map[string]string {
	manifests := map[string][]string{}

	manifests["namespace"] = []string{
		"components/ovn/namespace.yaml",
	}
	manifests["serviceaccount"] = []string{
		"components/ovn/node/serviceaccount.yaml",
		"components/ovn/master/serviceaccount.yaml",
	}
	manifests["role"] = []string{
		"components/ovn/role.yaml",
	}
	manifests["rolebinding"] = []string{
		"components/ovn/rolebinding.yaml",
	}
	manifests["clusterrole"] = []string{
		"components/ovn/clusterrole.yaml",
	}
	manifests["clusterrolebinding"] = []string{
		"components/ovn/clusterrolebinding.yaml",
	}
	manifests["configmap"] = []string{
		"components/ovn/configmap.yaml",
	}
	manifests["daemonset"] = []string{
		"components/ovn/master/daemonset.yaml",
		"components/ovn/node/daemonset.yaml",
	}
}

func (p *OVNKubernetesPlugin) ValidateConfig() bool {
	return true
}

func (p *OVNKubernetesPlugin) GetRenderParams() *OVNKubernetesConfig {
	c, err := NewOVNKubernetesConfigFromFileOrDefault("/etc/microshift/ovn.yaml")
	if err != nil {
		return &OVNKubernetesConfig{}
	}
	return c
}
