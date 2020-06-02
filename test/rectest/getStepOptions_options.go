package rectest

// Code generated by github.com/launchdarkly/go-options.  DO NOT EDIT.

import (
	"k8s.io/apimachinery/pkg/runtime"
)

type ApplyGetStepOptionFunc func(c *getStepOptions) error

func (f ApplyGetStepOptionFunc) apply(c *getStepOptions) error {
	return f(c)
}

func newGetStepOptions(options ...GetStepOption) (getStepOptions, error) {
	var c getStepOptions
	err := applyGetStepOptionsOptions(&c, options...)
	return c, err
}

func applyGetStepOptionsOptions(c *getStepOptions, options ...GetStepOption) error {
	c.Labels = map[string]string{}
	c.CheckResult = Ignore
	for _, o := range options {
		if err := o.apply(c); err != nil {
			return err
		}
	}
	return nil
}

type GetStepOption interface {
	apply(*getStepOptions) error
}

func GetWithNamespacedName(Name string, Namespace string) ApplyGetStepOptionFunc {
	return func(c *getStepOptions) error {
		c.NamespacedName.Name = Name
		c.NamespacedName.Namespace = Namespace
		return nil
	}
}

func GetWithObj(o runtime.Object) ApplyGetStepOptionFunc {
	return func(c *getStepOptions) error {
		c.Obj = o
		return nil
	}
}

func GetWithLabels(o map[string]string) ApplyGetStepOptionFunc {
	return func(c *getStepOptions) error {
		c.Labels = o
		return nil
	}
}

func GetWithCheckResult(o ReconcilerTestValidationFunc) ApplyGetStepOptionFunc {
	return func(c *getStepOptions) error {
		c.CheckResult = o
		return nil
	}
}
