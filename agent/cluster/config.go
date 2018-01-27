package cluster

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/mitchellh/hashstructure"
	"github.com/mitchellh/mapstructure"
	"io"
	"time"
)

// Cluster config
type Config struct {
	NodeID        string        `mapstructure:"node_id"`
	BackendURL    string        `mapstructure:"backend"` // kind://address/chroot
	Advertise     string        `mapstructure:"advertise"`
	TTL           time.Duration `mapstructure:"ttl"`
	RetryInterval time.Duration `mapstructure:"retry"`
}

func DefaultConfig() (c Config) {
	return Config{
		BackendURL:    "local://localhost/soil",
		Advertise:     "localhost:7654",
		NodeID:        "",
		TTL:           time.Minute * 3,
		RetryInterval: time.Second * 30,
	}
}

func (c Config) IsEqual(config Config) (res bool) {
	left, _ := hashstructure.Hash(c, nil)
	right, _ := hashstructure.Hash(config, nil)
	return left == right
}

func (c *Config) Unmarshal(readers ...io.Reader) (err error) {
	var failures []error
	for _, reader := range readers {
		if failure := c.unmarshal(reader); failure != nil {
			failures = append(failures, failure)
		}
	}
	if len(failures) > 0 {
		return fmt.Errorf("%v", failures)
	}
	return nil
}

func (c *Config) unmarshal(r io.Reader) (err error) {
	var buf bytes.Buffer
	if _, err = io.Copy(&buf, r); err != nil {
		return err
	}
	root, err := hcl.Parse(buf.String())
	if err != nil {
		return fmt.Errorf("error parsing: %s", err)
	}
	buf.Reset()

	list, ok := root.Node.(*ast.ObjectList)
	if !ok {
		return fmt.Errorf("error parsing: %s", fmt.Errorf("error parsing: root should be an object"))
	}
	matches := list.Filter("cluster")

	var failures []error
	for _, m := range matches.Items {
		var failure error
		var values map[string]interface{}
		if failure = hcl.DecodeObject(&values, m.Val); failure != nil {
			failures = append(failures, failure)
			continue
		}
		// try to parse TTL and retry
		config := &mapstructure.DecoderConfig{
			DecodeHook:       mapstructure.StringToTimeDurationHookFunc(),
			Metadata:         nil,
			Result:           c,
			WeaklyTypedInput: true,
		}
		var dec *mapstructure.Decoder
		if dec, failure = mapstructure.NewDecoder(config); failure != nil {
			failures = append(failures, failure)
			continue
		}
		if failure = dec.Decode(values); failure != nil {
			failures = append(failures, failure)
			continue
		}
	}
	if len(failures) > 0 {
		return fmt.Errorf("%v", failures)
	}
	return nil
}
