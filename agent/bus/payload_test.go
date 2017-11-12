// +build ide test_unit

package bus_test

import (
	"encoding/json"
	"github.com/akaspin/soil/agent/bus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFlatMapPayload(t *testing.T) {
	payload := bus.NewFlatMapPayload(map[string]string{"1": "1"})
	t.Run("unmarshal", func(t *testing.T) {
		var v map[string]string
		err := payload.Unmarshal(&v)
		assert.NoError(t, err)
		assert.Equal(t, map[string]string{"1": "1"}, v)
	})
	t.Run("clone", func(t *testing.T) {
		res := payload.Clone()
		assert.Equal(t, res.Hash(), payload.Hash())
	})
	t.Run("unmarshal empty", func(t *testing.T) {
		payload1 := bus.NewFlatMapPayload(nil)
		var v map[string]string
		err := payload1.Unmarshal(&v)
		assert.NoError(t, err)
		assert.NotNil(t, v)
		assert.Equal(t, map[string]string{}, v)
	})
	t.Run("json", func(t *testing.T) {
		res, err := payload.JSON()
		assert.NoError(t, err)
		assert.Equal(t, `{"1":"1"}`, string(res))
	})
}

func TestJSONPayload(t *testing.T) {
	type dummy struct {
		A string
	}
	t.Run("unmarshal", func(t *testing.T) {
		expect := dummy{"a"}
		data, jErr := json.Marshal(expect)
		assert.NoError(t, jErr)
		payload := bus.NewJSONPayload(data)
		var v dummy
		err := payload.Unmarshal(&v)
		assert.NoError(t, err)
		assert.Equal(t, expect, v)
	})
	t.Run("slice", func(t *testing.T) {
		ingest := []interface{}{
			map[string]string{
				"A": "1",
			},
			map[string]string{
				"A": "2",
			},
		}
		data, jErr := json.Marshal(ingest)
		assert.NoError(t, jErr)
		payload := bus.NewJSONPayload(data)
		var v []dummy
		assert.NoError(t, payload.Unmarshal(&v))
		assert.Equal(t, []dummy{
			{A: "1"},
			{A: "2"},
		}, v)
	})
	t.Run("map", func(t *testing.T) {
		ingest := map[string]interface{}{
			"A": map[string]string{
				"A": "1",
			},
			"B": map[string]string{
				"A": "2",
			},
		}
		data, jErr := json.Marshal(ingest)
		assert.NoError(t, jErr)
		payload := bus.NewJSONPayload(data)
		var v map[string]dummy
		assert.NoError(t, payload.Unmarshal(&v))
		assert.Equal(t, map[string]dummy{
			"A": {A: "1"},
			"B": {A: "2"},
		}, v)
	})
}

func TestNewPayload(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		p := bus.NewPayload(nil)
		_, ok := p.(bus.FlatMapPayload)
		assert.True(t, ok)
		assert.True(t, p.IsEmpty())
	})
	t.Run("flat map", func(t *testing.T) {
		p := bus.NewPayload(map[string]string{
			"1": "1",
		})
		_, ok := p.(bus.FlatMapPayload)
		assert.True(t, ok)
	})
	t.Run("json", func(t *testing.T) {
		p := bus.NewPayload(map[string]interface{}{
			"1": 1,
		})
		_, ok := p.(bus.JSONPayload)
		assert.True(t, ok)
	})
	t.Run("string", func(t *testing.T) {
		p := bus.NewPayload("test")
		var v string
		assert.NoError(t, p.Unmarshal(&v))
		assert.Equal(t, "test", v)
	})

}
