package ios_ansible

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/tetsusat/translator/bridge"
)

func init() {
	bridge.Register(new(Factory), "ios-ansible")
}

type Factory struct{}

func (f *Factory) New() bridge.AbstractAdapter {
	return &IOSAnsibleAdapter{
		ip:              os.Getenv("IOS_MGMT_IP"),
		tenantConfigs:   map[string]bridge.TenantConfig{},
		floatingConfigs: map[string]bridge.FloatingIPConfig{},
	}
}

type IOSAnsibleAdapter struct {
	ip              string
	tenantConfigs   map[string]bridge.TenantConfig
	floatingConfigs map[string]bridge.FloatingIPConfig
}

func (r *IOSAnsibleAdapter) AddTenant(tenant bridge.TenantConfig) error {
	b, err := json.Marshal(tenant)
	args := fmt.Sprintf("ansible-playbook /playbooks/ios-add-tenant.yml -i %s, --extra-vars '%s'", r.ip, string(b))
	log.Println(args)
	cmd := exec.Command("sh", "-c", args)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		log.Println(stderr.String())
	} else {
		log.Println(out.String())
		r.tenantConfigs[tenant.NetworkID] = tenant
	}
	return err
}

func (r *IOSAnsibleAdapter) DeleteTenant(id string) error {
	log.Printf("ID: %s", id)
	tenant, _ := r.tenantConfigs[id]
	b, err := json.Marshal(tenant)
	args := fmt.Sprintf("ansible-playbook /playbooks/ios-delete-tenant.yml -i %s, --extra-vars '%s'", r.ip, string(b))
	log.Println(args)
	cmd := exec.Command("sh", "-c", args)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		log.Println(stderr.String())
	} else {
		log.Println(out.String())
		delete(r.tenantConfigs, tenant.VRF)
	}
	return err
}

func (r *IOSAnsibleAdapter) AddFloatingIP(float bridge.FloatingIPConfig) error {
	b, err := json.Marshal(float)
	args := fmt.Sprintf("ansible-playbook /playbooks/ios-add-floating-ip.yml -i %s, --extra-vars '%s'", r.ip, string(b))
	log.Println(args)
	cmd := exec.Command("sh", "-c", args)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		log.Println(stderr.String())
	} else {
		log.Println(out.String())
		r.floatingConfigs[float.ContainerID] = float
	}
	return err
}

func (r *IOSAnsibleAdapter) DeleteFloatingIP(id string) error {
	float, _ := r.floatingConfigs[id]
	log.Printf("Global IP: %s", float.GlobalIP)
	if float.GlobalIP != "" {
		b, err := json.Marshal(float)
		args := fmt.Sprintf("ansible-playbook /playbooks/ios-delete-floating-ip.yml -i %s, --extra-vars '%s'", r.ip, string(b))
		log.Println(args)
		cmd := exec.Command("sh", "-c", args)
		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			log.Println(stderr.String())
		} else {
			log.Println(out.String())
			delete(r.floatingConfigs, float.ContainerID)
		}
		return err
	}
	return nil
}

func (r *IOSAnsibleAdapter) Sync(all bridge.AllConfig) error {
	b, err := json.Marshal(all)
	log.Printf("%s", string(b))
	args := fmt.Sprintf("ansible-playbook /playbooks/ios-add-all.yml -i %s, --extra-vars '%s'", r.ip, string(b))
	log.Println(args)
	cmd := exec.Command("sh", "-c", args)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		log.Println(stderr.String())
	} else {
		log.Println(out.String())
	}
	return err
}
