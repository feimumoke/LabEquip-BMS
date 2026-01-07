package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"

	"gopkg.in/yaml.v2"
)

type RegisterItem struct {
	Name   string
	Init   func(interface{})
	Config interface{}
}

var RegisterInitMap = make(map[string][]*RegisterItem)
var ConfigMap = make(map[string]interface{})

func RegisterInitWithConfig(name string, init func(interface{}), config interface{}) {
	RegisterInitMap[name] = append(RegisterInitMap[name], &RegisterItem{
		Name:   name,
		Init:   init,
		Config: config,
	})
}

func DoInitWithPath(filePath string) error {
	buf, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	ConfigMap = make(map[string]interface{})
	err = yaml.Unmarshal(buf, ConfigMap)
	if err != nil {
		return err
	}

	fmt.Println(ConfigMap)

	for _, ol := range RegisterInitMap {
		for _, o := range ol {
			if o.Config == nil {
				o.Init(nil)
				continue
			}

			c := reflect.New(reflect.TypeOf(o.Config).Elem()).Interface()
			v, ok := ConfigMap[o.Name]
			if !ok {
				continue
			}
			err = interfaceToObject(v, c)
			if err != nil {
				return err
			}
			o.Init(c)
		}
	}
	return nil
}

func GetConfig(name string, c interface{}) error {
	v, ok := ConfigMap[name]
	if !ok {
		return errors.New(fmt.Sprintf("%s config is not exist", name))
	}
	return interfaceToObject(v, c)
}

func interfaceToObject(i interface{}, o interface{}) error {
	buf, err := yaml.Marshal(i)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(buf, o)
	if err != nil {
		return err
	}
	return nil
}
