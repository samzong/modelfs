package v1

// GroupVersion identifies the API group and version for model resources.
type GroupVersion struct {
	Group   string
	Version string
}

// KindDescriptor captures the group-kind tuple for a resource.
type KindDescriptor struct {
	Group string
	Kind  string
}

var groupVersion = GroupVersion{Group: "model.samzong.dev", Version: "v1"}

// Group returns the group name.
func Group() string { return groupVersion.Group }

// Version returns the version name.
func Version() string { return groupVersion.Version }

// Kind returns the group/kind tuple for a resource name.
func Kind(kind string) KindDescriptor {
	return KindDescriptor{Group: groupVersion.Group, Kind: kind}
}
