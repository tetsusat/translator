package ios_restconf

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/tetsusat/translator/bridge"
)

func init() {
	bridge.Register(new(Factory), "ios-restconf")
}

type Factory struct{}

func (f *Factory) New() bridge.AbstractAdapter {
	return &IOSRestconfAdapter{
		ip:              os.Getenv("IOS_MGMT_IP"),
		user:            os.Getenv("IOS_USER"),
		pass:            os.Getenv("IOS_PASS"),
		client:          &http.Client{},
		enablePass:      os.Getenv("IOS_ENABLE_PASS"),
		tenantConfigs:   map[string]bridge.TenantConfig{},
		floatingConfigs: map[string]bridge.FloatingIPConfig{},
	}
}

type IOSRestconfAdapter struct {
	ip              string
	user            string
	pass            string
	enablePass      string
	client          *http.Client
	tenantConfigs   map[string]bridge.TenantConfig
	floatingConfigs map[string]bridge.FloatingIPConfig
}

func (i *IOSRestconfAdapter) AddTenant(tenant bridge.TenantConfig) error {
	// tenant
	var tenantBody bytes.Buffer
	tenantTmpl, err := template.New("tenant").Parse(tenantTemplate)
	err = tenantTmpl.Execute(&tenantBody, tenant)
	log.Printf("BODY: %s", tenantBody.String())
	tenantURI := fmt.Sprintf("http://%s/restconf/api/running/native", i.ip)
	log.Printf("URL: %s", tenantURI)
	err = i.doPatch(tenantURI, &tenantBody)
	if err != nil {
		return err
	}
	i.tenantConfigs[tenant.NetworkID] = tenant

	return nil
}

func (i *IOSRestconfAdapter) DeleteTenant(id string) error {
	log.Printf("ID: %s", id)
	tenant, _ := i.tenantConfigs[id]

	dpatURI := fmt.Sprintf("http://%s/restconf/api/running/native/ip/nat/inside/source/list/NAT", i.ip)
	err := i.doDelete(dpatURI)
	if err != nil {
		return err
	}

	interfacePolicyURI := fmt.Sprintf("http://%s/restconf/api/running/native/interface/%s/%s.%s/ip/policy", i.ip, tenant.InsideInterfaceType, tenant.InsideInterfaceID, tenant.VlanID)
	log.Printf("URL: %s", interfacePolicyURI)
	err = i.doDelete(interfacePolicyURI)
	if err != nil {
		return err
	}

	interfaceAddressURI := fmt.Sprintf("http://%s/restconf/api/running/native/interface/%s/%s.%s/ip/address/primary", i.ip, tenant.InsideInterfaceType, tenant.InsideInterfaceID, tenant.VlanID)
	log.Printf("URL: %s", interfaceAddressURI)
	err = i.doDelete(interfaceAddressURI)
	if err != nil {
		return err
	}

	interfaceVrfURI := fmt.Sprintf("http://%s/restconf/api/running/native/interface/%s/%s.%s/ip-vrf/ip/vrf/forwarding", i.ip, tenant.InsideInterfaceType, tenant.InsideInterfaceID, tenant.VlanID)
	log.Printf("URL: %s", interfaceVrfURI)
	err = i.doDelete(interfaceVrfURI)
	if err != nil {
		return err
	}

	interfaceURI := fmt.Sprintf("http://%s/restconf/api/running/native/interface/%s/%s.%s", i.ip, tenant.InsideInterfaceType, tenant.InsideInterfaceID, tenant.VlanID)
	log.Printf("URL: %s", interfaceURI)
	err = i.doDelete(interfaceURI)
	if err != nil {
		return err
	}

	vrfURI := fmt.Sprintf("http://%s/restconf/api/running/native/ip/vrf/%s", i.ip, tenant.VRF)
	log.Printf("URL: %s", vrfURI)
	err = i.doDelete(vrfURI)
	if err != nil {
		return err
	}

	delete(i.tenantConfigs, tenant.NetworkID)
	return nil
}

func (i *IOSRestconfAdapter) AddFloatingIP(float bridge.FloatingIPConfig) error {
	// static nat
	if float.GlobalIP != "" {
		var floatingBody bytes.Buffer
		floatingTmpl, err := template.New("static nat").Parse(floatingIPTemplate)
		err = floatingTmpl.Execute(&floatingBody, float)
		if err != nil {
			return err
		}
		log.Printf("BODY: %s", floatingBody.String())
		snatURI := fmt.Sprintf("http://%s/restconf/api/running/native/ip/nat", i.ip)
		log.Printf("URL: %s", snatURI)
		err = i.doPatch(snatURI, &floatingBody)
		if err != nil {
			return err
		}
	}

	i.floatingConfigs[float.ContainerID] = float
	return nil
}

func (i *IOSRestconfAdapter) DeleteFloatingIP(id string) error {
	float, _ := i.floatingConfigs[id]
	if float.GlobalIP != "" {
		snatURI := fmt.Sprintf("http://%s/restconf/api/running/native/ip/nat/inside/source/static/nat-static-transport-list/%s,%s", i.ip, float.LocalIP, float.GlobalIP)
		log.Printf("URL: %s", snatURI)
		err := i.doDelete(snatURI)
		if err != nil {
			return err
		}
	}
	delete(i.floatingConfigs, float.ContainerID)
	return nil
}

func (*IOSRestconfAdapter) Sync(all bridge.AllConfig) error {
	return nil
}

func (i *IOSRestconfAdapter) doPatch(uri string, body *bytes.Buffer) error {
	//r := strings.NewReader(body)
	req, err := http.NewRequest("PATCH", uri, body)
	req.Header.Set("Content-Type", "application/vnd.yang.data+json")
	req.Header.Add("Accept", "application/vnd.yang.data+json")
	req.SetBasicAuth(i.user, i.pass)
	res, err := i.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	log.Printf("Status: %s", res.Status)
	b, err := ioutil.ReadAll(res.Body)
	log.Printf(string(b))
	return nil
}

func (i *IOSRestconfAdapter) doDelete(uri string) error {
	//r := strings.NewReader(body)
	req, err := http.NewRequest("DELETE", uri, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/vnd.yang.data+json")
	req.Header.Add("Accept", "application/vnd.yang.data+json")
	req.SetBasicAuth(i.user, i.pass)
	res, err := i.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	log.Printf("Status: %s", res.Status)
	b, err := ioutil.ReadAll(res.Body)
	log.Printf(string(b))
	return nil
}
