package v1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// GroupVersion identifies the API group and version for model resources.
var (
	GroupVersion       = schema.GroupVersion{Group: "model.samzong.dev", Version: "v1"}
	SchemeGroupVersion = GroupVersion
	SchemeBuilder      = &scheme.Builder{GroupVersion: GroupVersion}
	AddToScheme        = SchemeBuilder.AddToScheme
)

func init() {
	SchemeBuilder.Register(&Model{}, &ModelList{})
	SchemeBuilder.Register(&ModelSource{}, &ModelSourceList{})
}
