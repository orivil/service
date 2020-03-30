// Copyright 2018 orivil.com. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found at https://mit-license.org.

// service 包是 golang 版的对象容器, 可实现对象依赖自动注入
package service

import (
	"errors"
	"sync"
)

var ErrProviderNotExist = errors.New("provider not exist")

// Provider 为服务提供者, 其功能相当于 new() 方法, 用于创建对象.
type Provider func(c *Container) (value interface{}, err error)

func NewServiceProvider(provider Provider) *Provider {
	return &provider
}

// ProviderContainer 主要用于给服务提供者命名
type ProviderContainer struct {
	providers map[string]*Provider
	mu        sync.RWMutex
}

func NewProviderContainer() *ProviderContainer {
	return &ProviderContainer{
		providers: make(map[string]*Provider, 5),
	}
}

func (pc *ProviderContainer) FlashCache(c *Container, providerName string) {
	provider := pc.GetProvider(providerName)
	if provider != nil {
		c.FlashCache(provider)
	}
}

func (pc *ProviderContainer) SetProvider(name string, provider *Provider) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.providers[name] = provider
}

func (pc *ProviderContainer) GetProvider(name string) (provider *Provider) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return pc.providers[name]
}

func (pc *ProviderContainer) MustGet(c *Container, providerName string) (v interface{}) {
	v, err := pc.Get(c, providerName)
	if err != nil {
		panic(err)
	}
	return v
}

func (pc *ProviderContainer) Get(c *Container, providerName string) (v interface{}, err error) {
	provider := pc.GetProvider(providerName)
	if provider != nil {
		return c.Get(provider)
	} else {
		return nil, ErrProviderNotExist
	}
}

func (pc *ProviderContainer) MustGetNew(c *Container, providerName string) (v interface{}) {
	v, err := pc.GetNew(c, providerName)
	if err != nil {
		panic(err)
	}
	return v
}

func (pc *ProviderContainer) GetNew(c *Container, providerName string) (v interface{}, err error) {
	provider := pc.GetProvider(providerName)
	if provider != nil {
		return c.GetNew(provider)
	} else {
		return nil, ErrProviderNotExist
	}
}

func (pc *ProviderContainer) SetCache(c *Container, providerName string, new interface{}) (old interface{}) {
	provider := pc.GetProvider(providerName)
	if provider != nil {
		return c.SetCache(provider, new)
	} else {
		return nil
	}
}

func (pc *ProviderContainer) HasCache(c *Container, providerName string) bool {
	provider := pc.GetProvider(providerName)
	if provider != nil {
		return c.HasCache(provider)
	} else {
		return false
	}
}

// Container 作为对象容器(服务容器), 被所有服务所依赖. Container 分为单线程模式及多线程模式.
type Container struct {
	instances map[*Provider]interface{}
	mu        *sync.RWMutex
}

// safe 为 false 时为单线程模式(即只在单线程中运行), 不需要加锁, 因此不会发生阻塞, 单线程模式
// 涵盖了大多数应用场景.
// safe 为 true 时为多线程模式, 线程安全, 但有可能因为某一个服务阻塞而造成 Container 内所
// 有服务都成为阻塞状态, 因此多线程模式下最好将不同功能的服务放在不同的容器中.
func NewContainer(safe bool) *Container {
	c := &Container{
		instances: make(map[*Provider]interface{}, 10),
	}
	if safe {
		c.mu = &sync.RWMutex{}
	}
	return c
}

func (c *Container) MustGet(p *Provider) (value interface{}) {
	value, err := c.Get(p)
	if err != nil {
		panic(err)
	}
	return value
}

// Get 用于获取服务的单例对象. 无论 Get 被调用多少次, p 只执行一次.
func (c *Container) Get(p *Provider) (value interface{}, err error) {
	if c.mu != nil {
		c.mu.RLock()
		defer c.mu.RUnlock()
	}
	if ins, ok := c.instances[p]; ok {
		return ins, nil
	} else {
		ins, err = (*p)(c)
		if err != nil {
			return nil, err
		} else {
			c.instances[p] = ins
			return ins, nil
		}
	}
}

func (c *Container) MustGetNew(p *Provider) (value interface{}) {
	value, err := c.GetNew(p)
	if err != nil {
		panic(err)
	} else {
		return value
	}
}

// GetNew 总是返回新的实例. 每次调用 GetNew 都会执行 p.
func (c *Container) GetNew(p *Provider) (value interface{}, err error) {
	return (*p)(c)
}

// 设置新的缓存并返回旧的缓存.
func (c *Container) SetCache(p *Provider, new interface{}) (old interface{}) {
	if c.mu != nil {
		c.mu.Lock()
		defer c.mu.Unlock()
	}
	old = c.instances[p]
	c.instances[p] = new
	return
}

func (c *Container) FlashCache(p *Provider) {
	if c.mu != nil {
		c.mu.Lock()
		defer c.mu.Unlock()
	}
	delete(c.instances, p)
}

// 判断 p 是否有缓存实列
func (c *Container) HasCache(p *Provider) bool {
	if c.mu != nil {
		c.mu.Lock()
		defer c.mu.Unlock()
	}
	_, ok := c.instances[p]
	return ok
}
