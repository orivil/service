// Copyright 2018 morgine.com. All rights reserved.

// Service package provided a dependency injection container
package service

import (
	"sync"
)

// Provider interface provide service, it only execute once
// if in singleton mode
type Provider interface {
	New(ctn *Container) (value interface{}, err error)
}

// ProviderFunc implemented the Provider interface
type ProviderFunc func(ctn *Container) (value interface{}, err error)

func (pf ProviderFunc) New(ctn *Container) (value interface{}, err error) {
	return pf(ctn)
}

// Container is a dependency injection container, it contains
// singleton objects
type Container struct {
	instances   map[*Provider]interface{}
	mus         map[*Provider]*sync.Mutex
	beforeClose []func() error
	mu          sync.Mutex
}

func NewContainer() *Container {
	return &Container{
		instances: make(map[*Provider]interface{}, 10),
		mus:       make(map[*Provider]*sync.Mutex, 10),
	}
}

// OnClose for registering callback functions execute at container Close
func (c *Container) OnClose(callback func() error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.beforeClose = append(c.beforeClose, callback)
}

// Close for triggering the registered OnClose callbacks
func (c *Container) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	var errs Errors
	for _, call := range c.beforeClose {
		err := call()
		if err != nil {
			errs = append(errs)
		}
	}
	return errs
}

func (c *Container) getProviderLocker(p *Provider) *sync.Mutex {
	c.mu.Lock()
	locker := c.mus[p]
	if locker == nil {
		locker = &sync.Mutex{}
		c.mus[p] = locker
	}
	c.mu.Unlock()
	return locker
}

// Get for getting singleton service, p.New function only execute once, and save the result
func (c *Container) Get(p *Provider) (interface{}, error) {
	mu := c.getProviderLocker(p)
	mu.Lock()
	defer mu.Unlock()
	if ins, ok := c.instances[p]; ok {
		return ins, nil
	} else {
		var err error
		ins, err = (*p).New(c)
		if err != nil {
			return nil, err
		} else {
			c.instances[p] = ins
			return ins, nil
		}
	}
}

// GetNew always execute p.New function and return the results
func (c *Container) GetNew(p *Provider) (interface{}, error) {
	return (*p).New(c)
}

// SetGet for setting new cache and return the old cache
func (c *Container) SetGet(p *Provider, new interface{}) (old interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	old = c.instances[p]
	c.instances[p] = new
	return
}

// Flash for delete cache
func (c *Container) Flash(p *Provider) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.instances, p)
}

// HasCache returns whether contains p's cache(whether p.New function exectue)
func (c *Container) HasCache(p *Provider) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.instances[p]
	return ok
}
