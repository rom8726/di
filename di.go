package di

import (
	"fmt"
	"reflect"
	"sync"
)

type Container struct {
	mu        sync.Mutex
	providers []*provider
	instances map[reflect.Type]reflect.Value
}

func New() *Container {
	return &Container{
		instances: make(map[reflect.Type]reflect.Value),
	}
}

func (c *Container) Provide(constructor any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	prvdr := newProvider(constructor)
	for _, pr := range c.providers {
		if pr.returnType == prvdr.returnType {
			panic(fmt.Errorf("duplicate provider %v", pr.returnType))
		}
	}

	c.providers = append(c.providers, prvdr)
}

func (c *Container) Resolve(target any) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	ptrVal := reflect.ValueOf(target)
	if ptrVal.Kind() != reflect.Ptr || ptrVal.Elem().Kind() != reflect.Interface {
		return fmt.Errorf("expected pointer to interface")
	}

	ifaceType := ptrVal.Elem().Type()

	impl, err := c.getInstanceByInterface(ifaceType)
	if err != nil {
		return err
	}

	ptrVal.Elem().Set(impl)

	return nil
}

func (c *Container) getInstanceByInterface(ifaceType reflect.Type) (reflect.Value, error) {
	if val, ok := c.instances[ifaceType]; ok {
		return val, nil
	}

	for _, prov := range c.providers {
		if prov.returnType.Implements(ifaceType) {
			inst, err := c.buildInstance(prov)
			if err != nil {
				return reflect.Value{}, err
			}

			c.instances[ifaceType] = inst

			return inst, nil
		}
	}

	return reflect.Value{}, fmt.Errorf("no provider found for interface %v", ifaceType)
}

func (c *Container) buildInstance(p *provider) (reflect.Value, error) {
	args := make([]any, len(p.paramTypes))
	for i, pt := range p.paramTypes {
		arg, err := c.getInstanceByType(pt)
		if err != nil {
			return reflect.Value{}, err
		}

		args[i] = arg
	}

	result, err := p.initFunc(args)
	if err != nil {
		return reflect.Value{}, err
	}

	return reflect.ValueOf(result), nil
}

func (c *Container) getInstanceByType(t reflect.Type) (any, error) {
	if val, ok := c.instances[t]; ok {
		return val.Interface(), nil
	}

	for _, prov := range c.providers {
		if prov.returnType.AssignableTo(t) || prov.returnType.Implements(t) {
			inst, err := c.buildInstance(prov)
			if err != nil {
				return reflect.Value{}, err
			}

			c.instances[t] = inst

			return inst.Interface(), nil
		}
	}

	return reflect.Value{}, fmt.Errorf("no provider found for type %v", t)
}
