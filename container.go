// Copyright 2018 orivil.com. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found at https://mit-license.org.

// service 包是 golang 版的对象容器, 可实现对象依赖自动注入
package service

import (
	"sync"
)

// Provider 为服务提供者, 其功能相当于 new() 方法, 用于创建对象.
type Provider interface {
	New(ctn *Container) (value interface{}, err error)
}

type ProviderFunc func(ctn *Container) (value interface{}, err error)

func (pf ProviderFunc) New(ctn *Container) (value interface{}, err error) {
	return pf(ctn)
}

func NewServiceProvider(provider Provider) *Provider {
	return &provider
}

// Container 是依赖容器, 用于管理服务的依赖
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

func (c *Container) OnClose(callback func() error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.beforeClose = append(c.beforeClose, callback)
}

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

func (c *Container) MustGet(p *Provider) (value interface{}) {
	var err error
	value, err = c.Get(p)
	if err != nil {
		panic(err)
	}
	return value
}

// Get 用于获取服务的单例对象. p.New() 的返回结果将会被保存, 出现错误除外
func (c *Container) Get(p *Provider) (value interface{}, err error) {
	mu := c.getProviderLocker(p)
	mu.Lock()
	defer mu.Unlock()
	if ins, ok := c.instances[p]; ok {
		return ins, nil
	} else {
		ins, err = (*p).New(c)
		if err != nil {
			return nil, err
		} else {
			c.instances[p] = ins
			return ins, nil
		}
	}
}

func (c *Container) MustGetNew(p *Provider) (value interface{}) {
	var err error
	value, err = c.GetNew(p)
	if err != nil {
		panic(err)
	} else {
		return value
	}
}

// GetNew 总是返回新的实例. 每次调用 GetNew 都会执行 p.New()
func (c *Container) GetNew(p *Provider) (value interface{}, err error) {
	return (*p).New(c)
}

// 设置新的缓存并返回旧的缓存.
func (c *Container) SetGet(p *Provider, new interface{}) (old interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	old = c.instances[p]
	c.instances[p] = new
	return
}

// 删除缓存
func (c *Container) Flash(p *Provider) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.instances, p)
}

// 判断 p 是否有缓存
func (c *Container) HasCache(p *Provider) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.instances[p]
	return ok
}
