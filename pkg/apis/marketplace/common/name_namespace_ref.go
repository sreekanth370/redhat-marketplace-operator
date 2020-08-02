package common

import (
	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

// JobStatus represents the current job for the report and it's status.
type NamespacedNameReference struct {

	// Namespace of the resource
	// Required
	UID types.UID `json:"uid"`

	// Namespace of the resource
	// Required
	Namespace string `json:"namespace"`

	// Name of the resource
	// Required
	Name string `json:"name"`

	// GroupVersionKind of the resourse
	// Required
	GroupVersionKind `json:"groupVersionKind"`
}

type GroupVersionKind struct {
	// APIVersion of the CRD
	APIVersion string `json:"apiVersion"`
	// Kind of the CRD
	Kind string `json:"kind"`
}

func NewGroupVersionKind(t interface{}) (GroupVersionKind, error) {
	typeAccessor, err := meta.TypeAccessor(t)
	if err != nil {
		return GroupVersionKind{}, err
	}

	return GroupVersionKind{
		APIVersion: typeAccessor.GetAPIVersion(),
		Kind:       typeAccessor.GetKind(),
	}, nil
}

func (n *NamespacedNameReference) ToTypes() types.NamespacedName {
	return types.NamespacedName{
		Namespace: n.Namespace,
		Name:      n.Name,
	}
}

func NamespacedNameFromMeta(t runtime.Object) *NamespacedNameReference {
	o, ok := t.(v1.Object)
	if !ok {
		return nil
	}

	return &NamespacedNameReference{
		UID:       o.GetUID(),
		Name:      o.GetName(),
		Namespace: o.GetNamespace(),
	}
}
