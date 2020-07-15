// Copyright 2018 orivil.com. All rights reserved.

package service_test

import (
	"encoding/json"
	"fmt"
	"github.com/morgine/service"
)

type Client struct {
	config *Config
}

type Config struct {
	Addr     string `json:"addr"`
	Password string `json:"password"`
}

type ConfigService struct {
	provider service.Provider
}

// Get is a helper function for getting *Config object
func (r *ConfigService) Get(ctn *service.Container) (*Config, error) {
	c, err := ctn.Get(&r.provider)
	if err != nil {
		return nil, err
	}
	return c.(*Config), nil
}

func newConfigService(jsonData []byte) *ConfigService {
	var provider service.ProviderFunc = func(ctn *service.Container) (value interface{}, err error) {
		cfg := &Config{}
		err = json.Unmarshal(jsonData, cfg)
		if err != nil {
			return nil, err
		}
		return cfg, nil
	}
	return &ConfigService{provider: provider}
}

type ClientService struct {
	provider service.Provider
}

// Get is a helper function for getting *Client object
func (r *ClientService) Get(ctn *service.Container) (*Client, error) {
	// Get function will return the singleton *Config object, the provider.New function only execute once.
	// GetNew function will always execute provider.New function and return the result
	c, err := ctn.Get(&r.provider)
	if err != nil {
		return nil, err
	}
	return c.(*Client), nil
}

// ClientService dependent on ConfigService
func newClientService(cs *ConfigService) *ClientService {
	var provider service.ProviderFunc = func(ctn *service.Container) (value interface{}, err error) {
		var cfg *Config
		// Get singleton cfg object
		cfg, err = cs.Get(ctn)
		if err != nil {
			return nil, err
		}
		return &Client{config: cfg}, nil
	}
	return &ClientService{provider: provider}
}

func ExampleContainer() {

	var config = `{
		"addr": "127.0.0.1:5432",
		"password": "secret key"
	}`

	var configService = newConfigService([]byte(config))

	var clientService = newClientService(configService)

	// container contains the singleton objects
	container := service.NewContainer()

	// client will auto inject the config dependency
	client, _ := clientService.Get(container)

	fmt.Println(client.config.Addr)
	fmt.Println(client.config.Password)

	cfg, _ := configService.Get(container)
	cfg.Addr = "localhost:5432"
	fmt.Println(client.config.Addr)

	// Output:
	// 127.0.0.1:5432
	// secret key
	// localhost:5432
}
