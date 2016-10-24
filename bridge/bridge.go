package bridge

import (
	"errors"
	"sync"

	dockerapi "github.com/fsouza/go-dockerclient"
)

type Bridge struct {
	sync.Mutex
	registry AbstractAdapter
	docker   *dockerapi.Client
}

func (b *Bridge) AddTenant(tenant TenantConfig) {
	b.Lock()
	defer b.Unlock()
	b.registry.AddTenant(tenant)
}

func (b *Bridge) DeleteTenant(id string) {
	b.Lock()
	defer b.Unlock()
	b.registry.DeleteTenant(id)
}

func (b *Bridge) AddFloatingIP(float FloatingIPConfig) {
	b.Lock()
	defer b.Unlock()
	b.registry.AddFloatingIP(float)
}

func (b *Bridge) DeleteFloatingIP(id string) {
	b.Lock()
	defer b.Unlock()
	b.registry.DeleteFloatingIP(id)
}

func (b *Bridge) Sync(all AllConfig) {
	b.Lock()
	defer b.Unlock()
	b.registry.Sync(all)
}

var factories = map[string]AdapterFactory{}

func New(docker *dockerapi.Client, adapter string) (*Bridge, error) {
	factory, found := factories[adapter]
	if !found {
		return nil, errors.New("unrecognized adapter: " + adapter)
	}

	return &Bridge{
		docker:   docker,
		registry: factory.New(),
	}, nil
}

func Register(adapter AdapterFactory, name string) {
	factories[name] = adapter
}
