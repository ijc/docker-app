package templatev1beta2

import (
	"github.com/docker/app/internal/helm/templatetypes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// StackList is a list of stacks
type StackList struct {
	metav1.TypeMeta `yaml:",inline"`
	metav1.ListMeta `yaml:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Items []Stack `yaml:"items" protobuf:"bytes,2,rep,name=items"`
}

// TypeMeta is a rewrite of metav1.TypeMeta which doesn't have yaml annotations
type TypeMeta struct {
	Kind       string `json:"kind,omitempty" yaml:"kind,omitempty"`
	APIVersion string `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
}

// GetObjectKind implements the ObjectKind interface
func (obj *TypeMeta) GetObjectKind() schema.ObjectKind {
	return obj
}

// GroupVersionKind implements the ObjectKind interface
func (obj *TypeMeta) GroupVersionKind() schema.GroupVersionKind {
	return schema.FromAPIVersionAndKind(obj.APIVersion, obj.Kind)
}

// SetGroupVersionKind implements the ObjectKind interface
func (obj *TypeMeta) SetGroupVersionKind(gvk schema.GroupVersionKind) {
	obj.APIVersion, obj.Kind = gvk.ToAPIVersionAndKind()
}

// Stack is v1beta2's representation of a Stack
type Stack struct {
	TypeMeta          `yaml:",inline" json:",inline"`
	metav1.ObjectMeta `yaml:"metadata,omitempty" json:"metadata,omitempty"`

	Spec   *StackSpec   `yaml:"spec,omitempty"`
	Status *StackStatus `yaml:"status,omitempty"`
}

// DeepCopyObject clones the stack
func (s *Stack) DeepCopyObject() runtime.Object {
	return s.clone()
}

// DeepCopyObject clones the stack list
func (s *StackList) DeepCopyObject() runtime.Object {
	if s == nil {
		return nil
	}
	result := new(StackList)
	result.TypeMeta = s.TypeMeta
	result.ListMeta = s.ListMeta
	if s.Items == nil {
		return result
	}
	result.Items = make([]Stack, len(s.Items))
	for ix, s := range s.Items {
		result.Items[ix] = *s.clone()
	}
	return result
}

func (s *Stack) clone() *Stack {
	if s == nil {
		return nil
	}
	result := new(Stack)
	result.TypeMeta = s.TypeMeta
	result.ObjectMeta = s.ObjectMeta
	result.Spec = s.Spec.clone()
	result.Status = s.Status.clone()
	return result
}

// StackSpec defines the desired state of Stack
type StackSpec struct {
	Services []ServiceConfig            `yaml:"services,omitempty"`
	Secrets  map[string]SecretConfig    `yaml:"secrets,omitempty"`
	Configs  map[string]ConfigObjConfig `yaml:"configs,omitempty"`
}

// ServiceConfig is the configuration of one service
type ServiceConfig struct {
	Name string `yaml:"name,omitempty"`

	CapAdd          []templatetypes.StringTemplate                                 `yaml:"cap_add,omitempty"`
	CapDrop         []templatetypes.StringTemplate                                 `yaml:"cap_drop,omitempty"`
	Command         []templatetypes.StringTemplate                                 `yaml:"command,omitempty"`
	Configs         []ServiceConfigObjConfig                                       `yaml:"configs,omitempty"`
	Deploy          DeployConfig                                                   `yaml:"deploy,omitempty"`
	Entrypoint      []templatetypes.StringTemplate                                 `yaml:"entrypoint,omitempty"`
	Environment     map[templatetypes.StringTemplate]*templatetypes.StringTemplate `yaml:"environment,omitempty"`
	ExtraHosts      []templatetypes.StringTemplate                                 `yaml:"extra_hosts,omitempty"`
	Hostname        templatetypes.StringTemplate                                   `yaml:"hostname,omitempty"`
	HealthCheck     *HealthCheckConfig                                             `yaml:"health_check,omitempty"`
	Image           templatetypes.StringTemplate                                   `yaml:"image,omitempty"`
	Ipc             templatetypes.StringTemplate                                   `yaml:"ipc,omitempty"`
	Labels          map[templatetypes.StringTemplate]templatetypes.StringTemplate  `yaml:"labels,omitempty"`
	Pid             templatetypes.StringTemplate                                   `yaml:"pid,omitempty"`
	Ports           []ServicePortConfig                                            `yaml:"ports,omitempty"`
	Privileged      templatetypes.BoolOrTemplate                                   `yaml:"privileged,omitempty" yaml:"privileged,omitempty"`
	ReadOnly        templatetypes.BoolOrTemplate                                   `yaml:"read_only,omitempty"`
	Secrets         []ServiceSecretConfig                                          `yaml:"secrets,omitempty"`
	StdinOpen       templatetypes.BoolOrTemplate                                   `yaml:"stdin_open,omitempty"`
	StopGracePeriod templatetypes.DurationOrTemplate                               `yaml:"stop_grace_period,omitempty"`
	Tmpfs           templatetypes.StringTemplateList                               `yaml:"tmpfs,omitempty"`
	Tty             templatetypes.BoolOrTemplate                                   `yaml:"tty,omitempty"`
	User            *int64                                                         `yaml:"user,omitempty"`
	Volumes         []ServiceVolumeConfig                                          `yaml:"volumes,omitempty"`
	WorkingDir      templatetypes.StringTemplate                                   `yaml:"working_dir,omitempty"`
}

// ServicePortConfig is the port configuration for a service
type ServicePortConfig struct {
	Mode      templatetypes.StringTemplate   `yaml:"mode,omitempty"`
	Target    templatetypes.UInt64OrTemplate `yaml:"target,omitempty"`
	Published templatetypes.UInt64OrTemplate `yaml:"published,omitempty"`
	Protocol  templatetypes.StringTemplate   `yaml:"protocol,omitempty"`
}

// FileObjectConfig is a config type for a file used by a service
type FileObjectConfig struct {
	Name     templatetypes.StringTemplate `yaml:"name,omitempty"`
	File     templatetypes.StringTemplate `yaml:"file,omitempty"`
	External External                     `yaml:"external,omitempty"`
	Labels   map[string]string            `yaml:"labels,omitempty"`
}

// SecretConfig for a secret
type SecretConfig FileObjectConfig

// ConfigObjConfig is the config for the swarm "Config" object
type ConfigObjConfig FileObjectConfig

// External identifies a Volume or Network as a reference to a resource that is
// not managed, and should already exist.
// External.name is deprecated and replaced by Volume.name
type External struct {
	Name     string `yaml:"name,omitempty"`
	External bool   `yaml:"external,omitempty"`
}

// FileReferenceConfig for a reference to a swarm file object
type FileReferenceConfig struct {
	Source templatetypes.StringTemplate   `yaml:"source,omitempty"`
	Target templatetypes.StringTemplate   `yaml:"target,omitempty"`
	UID    templatetypes.StringTemplate   `yaml:"uid,omitempty"`
	GID    templatetypes.StringTemplate   `yaml:"gid,omitempty"`
	Mode   templatetypes.UInt64OrTemplate `yaml:"mode,omitempty"`
}

// ServiceConfigObjConfig is the config obj configuration for a service
type ServiceConfigObjConfig FileReferenceConfig

// ServiceSecretConfig is the secret configuration for a service
type ServiceSecretConfig FileReferenceConfig

// DeployConfig is the deployment configuration for a service
type DeployConfig struct {
	Mode          templatetypes.StringTemplate                                  `yaml:"mode,omitempty"`
	Replicas      templatetypes.UInt64OrTemplate                                `yaml:"replicas,omitempty"`
	Labels        map[templatetypes.StringTemplate]templatetypes.StringTemplate `yaml:"labels,omitempty"`
	UpdateConfig  *UpdateConfig                                                 `yaml:"update_config,omitempty"`
	Resources     Resources                                                     `yaml:"resources,omitempty"`
	RestartPolicy *RestartPolicy                                                `yaml:"restart_policy,omitempty"`
	Placement     Placement                                                     `yaml:"placement,omitempty"`
}

// UpdateConfig is the service update configuration
type UpdateConfig struct {
	Parallelism templatetypes.UInt64OrTemplate `yaml:"paralellism,omitempty"`
}

// Resources the resource limits and reservations
type Resources struct {
	Limits       *Resource `yaml:"limits,omitempty"`
	Reservations *Resource `yaml:"reservations,omitempty"`
}

// Resource is a resource to be limited or reserved
type Resource struct {
	NanoCPUs    templatetypes.StringTemplate      `yaml:"cpus,omitempty"`
	MemoryBytes templatetypes.UnitBytesOrTemplate `yaml:"memory,omitempty"`
}

// RestartPolicy is the service restart policy
type RestartPolicy struct {
	Condition string `yaml:"condition,omitempty"`
}

// Placement constraints for the service
type Placement struct {
	Constraints *Constraints `yaml:"constraints,omitempty"`
}

// Constraints lists constraints that can be set on the service
type Constraints struct {
	OperatingSystem *Constraint
	Architecture    *Constraint
	Hostname        *Constraint
	MatchLabels     map[string]Constraint
}

// Constraint defines a constraint and it's operator (== or !=)
type Constraint struct {
	Value    string
	Operator string
}

// HealthCheckConfig the healthcheck configuration for a service
type HealthCheckConfig struct {
	Test     []string                         `yaml:"test,omitempty"`
	Timeout  templatetypes.DurationOrTemplate `yaml:"timeout,omitempty"`
	Interval templatetypes.DurationOrTemplate `yaml:"interval,omitempty"`
	Retries  templatetypes.UInt64OrTemplate   `yaml:"retries,omitempty"`
}

// ServiceVolumeConfig are references to a volume used by a service
type ServiceVolumeConfig struct {
	Type     string                       `yaml:"type,omitempty"`
	Source   templatetypes.StringTemplate `yaml:"source,omitempty"`
	Target   templatetypes.StringTemplate `yaml:"target,omitempty"`
	ReadOnly templatetypes.BoolOrTemplate `yaml:"read_only,omitempty"`
}

func (s *StackSpec) clone() *StackSpec {
	if s == nil {
		return nil
	}
	result := *s
	return &result
}

// StackPhase is the deployment phase of a stack
type StackPhase string

// These are valid conditions of a stack.
const (
	// StackAvailable means the stack is available.
	StackAvailable StackPhase = "Available"
	// StackProgressing means the deployment is progressing.
	StackProgressing StackPhase = "Progressing"
	// StackFailure is added in a stack when one of its members fails to be created
	// or deleted.
	StackFailure StackPhase = "Failure"
)

// StackStatus defines the observed state of Stack
type StackStatus struct {
	// Current condition of the stack.
	// +optional
	Phase StackPhase `yaml:"phase,omitempty" protobuf:"bytes,1,opt,name=phase,casttype=StackPhase"`
	// A human readable message indicating details about the stack.
	// +optional
	Message string `yaml:"message,omitempty" protobuf:"bytes,5,opt,name=message"`
}

func (s *StackStatus) clone() *StackStatus {
	if s == nil {
		return nil
	}
	result := *s
	return &result
}

// Clone clones a Stack
func (s *Stack) Clone() *Stack {
	return s.clone()
}
