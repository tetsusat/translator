package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	dockerapi "github.com/fsouza/go-dockerclient"
	"github.com/tetsusat/translator/bridge"
)

func assert(err error) {
	if err != nil {
		log.Fatal("fatal: ", err)
	}
}

func parseVLAN(str string) string {
	tmp := strings.Split(str, ".")
	return tmp[1]
}

func parseIPv4Address(str string) string {
	tmp := strings.Split(str, "/")
	return tmp[0]
}

func parseInterface(str string) (string, string, error) {
	fmt.Println(str)
	r := regexp.MustCompile(`([a-zA-z]+)([0-9\.]+)`)

	result := r.FindStringSubmatch(str)
	fmt.Println(result)

	if len(result) != 3 {
		return "", "", errors.New("interface name seems wrong")
	}

	interfaceType := result[1]
	interfaceID := result[2]

	return interfaceType, interfaceID, nil
}

func getConfig(docker *dockerapi.Client) bridge.AllConfig {
	opts := dockerapi.NetworkFilterOpts{
		"driver": map[string]bool{"macvlan": true},
	}

	networks, err := docker.FilteredListNetworks(opts)
	assert(err)

	var tenantConfigs []bridge.TenantConfig
	var floatingIPConfigs []bridge.FloatingIPConfig

	for _, network := range networks {
		// log.Printf("Net: %v", net.Name)
		log.Printf("Name: %s", network.Name)
		log.Printf("Options: %v", network.Options)
		parent, _ := network.Options["parent"]
		vlanID := parseVLAN(parent)
		log.Printf("Gateway: %s", network.IPAM.Config[0].Gateway)

		tenantConfig := bridge.TenantConfig{
			NetworkID: network.ID,
			VRF:       network.Name,
			VlanID:    vlanID,
			Gateway:   network.IPAM.Config[0].Gateway,
		}
		tenantConfigs = append(tenantConfigs, tenantConfig)

		for id, c := range network.Containers {
			container, err := docker.InspectContainer(id)
			assert(err)
			globalIP, ok := container.Config.Labels["global_ip"]
			if ok {
				localIP := parseIPv4Address(c.IPv4Address)
				log.Printf("Container ID: %s", c.ID)
				log.Printf("Global IP: %s", globalIP)
				log.Printf("Local IP: %s", localIP)

				floatingIPConfig := bridge.FloatingIPConfig{
					ContainerID: c.ID,
					VRF:         network.Name,
					GlobalIP:    globalIP,
					LocalIP:     localIP,
				}
				floatingIPConfigs = append(floatingIPConfigs, floatingIPConfig)
			}
		}
	}

	config := bridge.AllConfig{
		TenantConfigs:     tenantConfigs,
		FloatingIPConfigs: floatingIPConfigs,
	}
	return config
}

func main() {

	flag.Parse()

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s [options] <adapter>\n\n", os.Args[0])
	}

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}

	if os.Getenv("DOCKER_HOST") == "" {
		assert(os.Setenv("DOCKER_HOST", "unix:///tmp/docker.sock"))
	}
	docker, err := dockerapi.NewClientFromEnv()
	assert(err)

	b, err := bridge.New(docker, flag.Arg(0)) // ip, user, pass, secret ...
	assert(err)

	events := make(chan *dockerapi.APIEvents)
	assert(docker.AddEventListener(events))
	log.Println("starting translator ...")

	var resyncInterval = 60

	quit := make(chan struct{})

	// Start the resync timer if enabled
	if resyncInterval > 0 {
		resyncTicker := time.NewTicker(time.Duration(resyncInterval) * time.Second)
		go func() {
			for {
				select {
				case <-resyncTicker.C:
					all := getConfig(docker)
					b.Sync(all)
				case <-quit:
					resyncTicker.Stop()
					return
				}
			}
		}()
	}

	for msg := range events {
		switch {
		case msg.Type == "network" && msg.Action == "create":
			log.Println("===> network create")
			log.Printf("Network: %s", msg.ID)
			log.Printf("Actor: %v", msg.Actor.Attributes)
			network, err := docker.NetworkInfo(msg.ID)
			assert(err)

			log.Printf("Name: %s", network.Name)
			log.Printf("Options: %v", network.Options)
			parent, _ := network.Options["parent"]
			vlanID := parseVLAN(parent)
			log.Printf("Gateway: %s", network.IPAM.Config[0].Gateway)
			insideInterface := os.Getenv("INSIDE_INTERFACE")
			insideInterfaceType, insideInterfaceID, err := parseInterface(insideInterface)
			assert(err)
			outsideInterface := os.Getenv("OUTSIDE_INTERFACE")
			outsideInterfaceType, outsideInterfaceID, err := parseInterface(outsideInterface)
			assert(err)

			config := bridge.TenantConfig{
				NetworkID:            msg.ID,
				VRF:                  network.Name,
				VlanID:               vlanID,
				InsideInterfaceType:  insideInterfaceType,
				InsideInterfaceID:    insideInterfaceID,
				OutsideInterfaceType: outsideInterfaceType,
				OutsideInterfaceID:   outsideInterfaceID,
				Gateway:              network.IPAM.Config[0].Gateway,
			}

			go b.AddTenant(config)

		case msg.Type == "network" && msg.Action == "connect":
			log.Println("===> network connect")
			attributes := msg.Actor.Attributes
			assert(err)
			name, _ := attributes["name"]
			log.Printf("Network: %s", name)
			id, _ := attributes["container"]
			log.Printf("Container ID: %s", id)
			container, err := docker.InspectContainer(id)
			assert(err)
			network, _ := container.NetworkSettings.Networks[name]
			log.Printf("IPAddress: %s", network.IPAddress) // ok?
			log.Printf("Labels: %v", container.Config.Labels)
			globalIP, ok := container.Config.Labels["global_ip"]

			if ok {
				config := bridge.FloatingIPConfig{
					ContainerID: id,
					VRF:         name,
					GlobalIP:    globalIP,
					LocalIP:     network.IPAddress,
				}
				go b.AddFloatingIP(config)
			} else {
				log.Printf("no global_ip label, do nothing...")
			}

		case msg.Type == "network" && msg.Action == "disconnect":
			log.Println("===> network disconnect")
			attributes := msg.Actor.Attributes
			assert(err)
			id, _ := attributes["container"]
			log.Printf("Container ID: %s", id)
			go b.DeleteFloatingIP(id)

		case msg.Type == "network" && msg.Action == "destroy":
			log.Println("===> network destroy")
			log.Printf("Network: %s", msg.ID)
			go b.DeleteTenant(msg.ID)
		}
	}
	log.Fatal("fatal: docker event loop closed") // todo: reconnect?
}
