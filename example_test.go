// Copyright 2018 orivil.com. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found at https://mit-license.org.

package service_test

import (
	"fmt"
	"github.com/orivil/service"
	"unsafe"
)

// 动物
type Animal struct {
	Type string
}

// 狗
type Dog struct {
	Name   string
	Animal *Animal
}

// 猫
type Cat struct {
	Name   string
	Animal *Animal
}

// 动物对象提供器
type AnimalProvider string

// implement service.Provider
func (p AnimalProvider) New(ctn *service.Container) (value interface{}, err error) {
	return &Animal{Type: string(p)}, nil
}

// 狗对象提供器
type DogProvider string

// implement service.Provider
func (p DogProvider) New(ctn *service.Container) (value interface{}, err error) {
	// 获得依赖( Get 方法为单例模式, 如果需要工厂模式则使用 GetNew 方法)
	value, err = ctn.Get(&mammalAnimal)
	if err != nil {
		return nil, err
	}
	dog := &Dog{
		Name:   string(p),
		Animal: value.(*Animal),
	}
	return dog, nil
}

// 猫对象提供器
type CatProvider string

// implement service.Provider
func (p CatProvider) New(ctn *service.Container) (value interface{}, err error) {
	// 获得依赖(单例模式)
	value, err = ctn.Get(&mammalAnimal)
	if err != nil {
		return nil, err
	}
	cat := &Cat{
		Name:   string(p),
		Animal: value.(*Animal),
	}
	return cat, nil
}

// 提供哺乳动物
var mammalAnimal = service.Provider(AnimalProvider("mammal"))

// 提供 tony dog
var dogTony = service.Provider(DogProvider("tony"))

// 提供 kevin cat
var catKevin = service.Provider(CatProvider("kevin"))

func ExampleContainer() {

	// 新建容器
	container := service.NewContainer()

	// 获取 tony 对象(服务)
	tony := container.MustGet(&dogTony).(*Dog)

	// 已注入依赖
	fmt.Printf("dog %s is a %s animal\n", tony.Name, tony.Animal.Type)

	// 获取 kevin 对象(服务)
	kevin := container.MustGet(&catKevin).(*Cat)

	fmt.Printf("cat %s is a %s animal\n", kevin.Name, kevin.Animal.Type)

	// 他们的依赖是同一个 Animal 对象
	fmt.Println(unsafe.Pointer(tony.Animal) == unsafe.Pointer(kevin.Animal)) // true

	// GetNew() 为工厂模式, 每次都调用 New() 方法新建对象
	newKevin := container.MustGetNew(&catKevin).(*Cat)

	fmt.Println(unsafe.Pointer(kevin) == unsafe.Pointer(newKevin)) // false

	// 工厂模式获得的依赖仍然可能是单例模式, 因为获取 Animal 对象的方法是 Get(), 而不是 GetNew()
	fmt.Println(unsafe.Pointer(newKevin.Animal) == unsafe.Pointer(tony.Animal)) // true

	// Output:
	// dog tony is a mammal animal
	// cat kevin is a mammal animal
	// true
	// false
	// true
}
