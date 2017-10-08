package allocation

import (
	"encoding/json"
	"fmt"
	"github.com/akaspin/soil/manifest"
	"io"
	"strings"
)

const resourceHeaderPrefix = "### RESOURCE"

// Allocated resource
type Resource struct {

	// Requested resource
	Request *manifest.Resource

	// Allocated values are in "resource.type.pod-name.resource-name._"
	Values map[string]string
}

func defaultResource() (r *Resource) {
	r = &Resource{
		Request: &manifest.Resource{},
	}
	return
}

func newResource(podName string, request *manifest.Resource, env map[string]string) (r *Resource) {
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
	if _, err = fmt.Fprintf(w, "### RESOURCE %s %s ", r.Request.Kind, r.Request.Name); err != nil {
		return
	}
	err = encoder.Encode(resourceHeader{
		Request: r.Request.Config,
		Values:  r.Values,
	})
	return
}

func (r *Resource) unmarshalHeader(line string) (err error) {
	if _, err = fmt.Sscanf(line, resourceHeaderPrefix+" %s %s ", &r.Request.Kind, &r.Request.Name); err != nil {
		return
	}
	jsonV := strings.TrimPrefix(line, fmt.Sprintf(resourceHeaderPrefix+" %s %s ", r.Request.Kind, r.Request.Name))
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
