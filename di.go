package di

import (
	"fmt"
	"reflect"
	"sync"
)

type Container struct {
	mu        sync.Mutex
	providers []*Provider
	instances map[reflect.Type]reflect.Value

	instancesList []any
	resolvedMap   map[reflect.Type]struct{}
}

func New() *Container {
	return &Container{
		instances:   make(map[reflect.Type]reflect.Value),
		resolvedMap: make(map[reflect.Type]struct{}),
	}
}

func (c *Container) Provide(constructor any) *Provider {
	c.mu.Lock()
	defer c.mu.Unlock()

	prvdr := newProvider(constructor)
	for _, pr := range c.providers {
		if pr.returnType == prvdr.returnType {
			panic(fmt.Errorf("duplicate provider %v", pr.returnType))
		}
	}

	c.providers = append(c.providers, prvdr)

	return prvdr
}

func (c *Container) Resolve(target any) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	ptrVal := reflect.ValueOf(target)
	if ptrVal.Kind() != reflect.Ptr {
		return fmt.Errorf("expected a pointer")
	}

	elemType := ptrVal.Elem().Type()

	if ptrVal.Elem().Kind() == reflect.Interface {
		impl, err := c.getInstanceByInterface(elemType)
		if err != nil {
			return fmt.Errorf("%w [target=%s]", err, getFuncName(ptrVal))
		}

		ptrVal.Elem().Set(impl)

		return nil
	}

	inst, err := c.getInstanceByType(elemType)
	if err != nil {
		return err
	}

	ptrVal.Elem().Set(reflect.ValueOf(inst))

	return nil
}

func (c *Container) ResolveToStruct(target any) error {
	ptrVal := reflect.ValueOf(target)

	if ptrVal.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer")
	}

	elemVal := ptrVal.Elem()

	if elemVal.Kind() != reflect.Struct {
		return fmt.Errorf("target must point to a struct")
	}

	elemType := elemVal.Type()
	for i := 0; i < elemType.NumField(); i++ {
		field := elemType.Field(i)
		fieldValue := elemVal.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		fieldValuePtr := fieldValue.Addr().Interface()

		if err := c.Resolve(fieldValuePtr); err != nil {
			return fmt.Errorf("failed to resolve field %q: %w", field.Name, err)
		}
	}

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

func (c *Container) buildInstance(p *Provider) (reflect.Value, error) {
	if _, ok := c.resolvedMap[p.returnType]; ok {
		return reflect.Value{}, fmt.Errorf("circular dependency detected: %v", p.name)
	}

	c.resolvedMap[p.returnType] = struct{}{}
	defer delete(c.resolvedMap, p.returnType)

	args := make([]any, len(p.paramTypes))
	for i, pt := range p.paramTypes {
		if arg, ok := p.args[pt]; ok {
			args[i] = arg.Interface()

			continue
		}

		arg, err := c.getInstanceByType(pt)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("%w [constructor: %s]", err, p.name)
		}

		args[i] = arg
	}

	result, err := p.initFunc(args)
	if err != nil {
		return reflect.Value{}, err
	}

	c.instancesList = append(c.instancesList, result)

	return reflect.ValueOf(result), nil
}

func (c *Container) getInstanceByType(t reflect.Type) (any, error) {
	if val, ok := c.instances[t]; ok {
		return val.Interface(), nil
	}

	for _, prov := range c.providers {
		if prov.returnType.AssignableTo(t) ||
			(prov.returnType.Kind() == reflect.Interface && prov.returnType.Implements(t)) {
			inst, err := c.buildInstance(prov)
			if err != nil {
				return reflect.Value{}, err
			}

			c.instances[t] = inst

			return inst.Interface(), nil
		}
	}

	return reflect.Value{}, fmt.Errorf("no provider\\arg found for type %v", t)
}
