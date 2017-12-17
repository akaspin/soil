package allocation

import (
	"encoding/json"
	"fmt"
	"github.com/akaspin/soil/manifest"
	"io"
	"strings"
)

const (
	resourceHeaderPrefix   = "### RESOURCE "
	resourceHeaderPrefixV2 = "### RESOURCE.V2 "
)

// Allocation resources
type ResourceSlice []*Resource

func (r *ResourceSlice) Append(v ItemUnmarshaller) {
	*r = append(*r, v.(*Resource))
}

// Allocated resource
type Resource struct {
	// Requested resource
	Request manifest.Resource

	// Allocated values stored in "resource.pod-name.<provider-name>.<resource-name>.__values" environment
	Values manifest.Environment `json:",omitempty"`
}

func (r *Resource) FromManifest(podName string, req manifest.Resource, env manifest.Environment) (err error) {
	r.Request = req

	// try to recover values from env

	return
}

// Returns values bag key without provider
func (r *Resource) ValuesKey() (res string) {
	res = r.Request.Name + "."
	return
}

// Marshal resource line
func (r *Resource) MarshalLine(w io.Writer) (err error) {
	if _, err = fmt.Fprint(w, resourceHeaderPrefixV2); err != nil {
		return
	}
	err = json.NewEncoder(w).Encode(r)
	return
}

func (r *Resource) UnmarshalLine(line string) (err error) {
	// old resources are skipped
	err = json.Unmarshal([]byte(strings.TrimPrefix(line, resourceHeaderPrefixV2)), r)
	return
}

func defaultResource() (r *Resource) {
	r = &Resource{
		Request: manifest.Resource{},
	}
	return
}

func newResource(podName string, request manifest.Resource, env map[string]string) (r *Resource) {
	r = &Resource{
		Request: request,
		Values:  map[string]string{},
	}
	// try to read values from bag
	if bag, ok := env[request.GetValuesKey(podName)]; ok {
		json.NewDecoder(strings.NewReader(bag)).Decode(&r.Values)
	}
	return
}

func (r *Resource) marshalHeader(w io.Writer, encoder *json.Encoder) (err error) {
	if _, err = fmt.Fprintf(w, "### RESOURCE %s %s ", r.Request.Provider, r.Request.Name); err != nil {
		return
	}
	err = encoder.Encode(resourceHeader{
		Request: r.Request.Config,
		Values:  r.Values,
	})
	return
}

func (r *Resource) unmarshalHeader(line string) (err error) {
	if _, err = fmt.Sscanf(line, resourceHeaderPrefix+"%s %s ", &r.Request.Provider, &r.Request.Name); err != nil {
		return
	}
	jsonV := strings.TrimPrefix(line, fmt.Sprintf(resourceHeaderPrefix+"%s %s ", r.Request.Provider, r.Request.Name))
	var receiver resourceHeader
	err = json.NewDecoder(strings.NewReader(jsonV)).Decode(&receiver)
	r.Request.Config = receiver.Request
	r.Values = receiver.Values
	return
}

type resourceHeader struct {
	Request map[string]interface{}
	Values  map[string]string
}
