package service_decorators

import "github.com/hashicorp/consul/api"

// ConfigStorage is to store the configuration
type ConfigStorage interface {
	Get(name string) ([]byte, error)
}

// ConsulConfigStorage the configuration storage with Consul KV
type ConsulConfigStorage struct {
	client *api.KV
}

// CreateConsulConfigStorage is to create a ConsulConfigStorage
func CreateConsulConfigStorage(consulConfig *api.Config) (*ConsulConfigStorage, error) {
	client, err := api.NewClient(consulConfig)
	if err != nil {
		return nil, err
	}
	return &ConsulConfigStorage{client.KV()}, nil
}

// Get is to get the configuration
func (storage *ConsulConfigStorage) Get(name string) ([]byte, error) {
	pair, _, err := storage.client.Get(name, nil)
	if err != nil {
		return nil, err
	}
	return pair.Value, nil
}

// Set is to set the configuration
func (storage *ConsulConfigStorage) Set(name string, value []byte) error {
	storage.client.Put(&api.KVPair{Key: name, Value: value}, nil)
	return nil
}
