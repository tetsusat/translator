package bridge

type AdapterFactory interface {
	New() AbstractAdapter
}

type AbstractAdapter interface {
	AddTenant(tenant TenantConfig) error
	DeleteTenant(id string) error
	AddFloatingIP(float FloatingIPConfig) error
	DeleteFloatingIP(id string) error
	Sync(AllConfig) error
}

type TenantConfig struct {
	NetworkID            string `json:"-"`
	VRF                  string `json:"vrf"`
	VlanID               string `json:"vlan_id"`
	InsideInterfaceType  string `json:"inside_interface_type"`
	InsideInterfaceID    string `json:"inside_interface_id"`
	OutsideInterfaceType string `json:"outside_interface_type"`
	OutsideInterfaceID   string `json:"outside_interface_id"`
	Gateway              string `json:"gateway"`
}

type FloatingIPConfig struct {
	ContainerID string `json:"-"`
	VRF         string `json:"vrf"`
	GlobalIP    string `json:"global_ip"`
	LocalIP     string `json:"local_ip"`
}

type AllConfig struct {
	TenantConfigs     []TenantConfig     `json:"tenant_configs"`
	FloatingIPConfigs []FloatingIPConfig `json:"floating_ip_configs"`
}
