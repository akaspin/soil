package allocation

import (
	"encoding/json"
	"fmt"
	"github.com/akaspin/soil/manifest"
	"github.com/hashicorp/go-multierror"
	"github.com/mitchellh/copystructure"
	"io"
	"strings"
)

const (
	resourceSpecPrefix    = "### RESOURCE "
	ResourceValuesPostfix = "__values"
)

// Allocation resources
type ResourceSlice []*Resource

func (s *ResourceSlice) GetEmpty(paths SystemPaths) (empty Asset) {
	return &Resource{}
}

func (s *ResourceSlice) GetVersionPrefix(v string) (p string) {
	return resourceSpecPrefix
}

func (r *ResourceSlice) FromManifest(m manifest.Pod, env manifest.FlatMap) (err error) {
	err = &multierror.Error{}
	for _, decl := range m.Resources {
		v, _ := copystructure.Copy(decl)
		var resource Resource
		if err1 := (&resource).FromManifest(m.Name, v.(manifest.Resource), env); err1 != nil {
			err = multierror.Append(err, err1)
		}
		*r = append(*r, &resource)
	}
	return err.(*multierror.Error).ErrorOrNil()
}

func (r *ResourceSlice) AppendItem(v Asset) {
	*r = append(*r, v.(*Resource))
}

// Allocated resource
type Resource struct {
	// Requested resource
	Request manifest.Resource

	// Allocated values stored in "resource.pod-name.<provider-name>.<resource-name>.__values" environment
	Values manifest.FlatMap `json:",omitempty"`
}

func (r Resource) String() string {
	return fmt.Sprint(r.Request, r.Values)
}

// Unmarshal resource allocation from manifest
func (r *Resource) FromManifest(podName string, req manifest.Resource, env manifest.FlatMap) (err error) {
	r.Request = req

	// try to recover values from env
	if values, ok := env["resource."+podName+"."+r.ValuesKey()]; ok {
		return json.Unmarshal([]byte(values), &r.Values)
	}
	return nil
}

// Returns values bag key without provider
func (r *Resource) ValuesKey() (res string) {
	return r.Request.Name + "." + ResourceValuesPostfix
}

// Marshal resource line
func (r *Resource) MarshalSpec(w io.Writer) (err error) {
	if _, err = fmt.Fprint(w, resourceSpecPrefix); err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(r)
}

func (r *Resource) UnmarshalSpec(line string, spec Spec, paths SystemPaths) (err error) {
	return json.Unmarshal([]byte(strings.TrimPrefix(line, resourceSpecPrefix)), r)
}

func (r *Resource) Clone() (res *Resource) {
	i, _ := copystructure.Copy(r)
	return i.(*Resource)
}
