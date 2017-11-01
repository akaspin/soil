// +build ide test_cluster

package fixture_test

import (
	"github.com/akaspin/soil/fixture"
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestConsulServer(t *testing.T) {
	s := fixture.NewConsulServer(t, nil)
	defer s.Clean()

	var cli *api.Client
	var err error
	var sess1, sess2 string

	t.Run("get api", func(t *testing.T) {
		cli, err = api.NewClient(&api.Config{
			Address:  s.Address(),
			WaitTime: time.Second * 30,
		})
		assert.NoError(t, err)
	})
	t.Run("up", func(t *testing.T) {
		s.Up()
		s.WaitAlive()
	})
	t.Run("get session 1", func(t *testing.T) {
		sess1, _, err = cli.Session().Create(&api.SessionEntry{
			Name: "sess1",
			TTL:  "3m",
		}, nil)
		assert.NoError(t, err)
		res, _, err := cli.Session().Info(sess1, nil)
		assert.NoError(t, err)
		assert.Equal(t, res.Name, "sess1")
	})
	t.Run("get session 2", func(t *testing.T) {
		sess2, _, err = cli.Session().Create(&api.SessionEntry{
			Name: "sess2",
			TTL:  "3m",
		}, nil)
		assert.NoError(t, err)
		res, _, err := cli.Session().Info(sess2, nil)
		assert.NoError(t, err)
		assert.Equal(t, res.Name, "sess2")
	})
	t.Run("acquire with sess1", func(t *testing.T) {
		ok, _, err := cli.KV().Acquire(&api.KVPair{
			Key:     "test",
			Value:   []byte("test"),
			Session: sess1,
		}, nil)
		assert.NoError(t, err)
		assert.True(t, ok)
	})
	t.Run("acquire with sess2", func(t *testing.T) {
		ok, _, err := cli.KV().Acquire(&api.KVPair{
			Key:     "test",
			Session: sess2,
		}, nil)
		assert.NoError(t, err)
		assert.False(t, ok)
	})
	t.Run("update with sess1", func(t *testing.T) {
		ok, _, err := cli.KV().Acquire(&api.KVPair{
			Key:     "test",
			Value:   []byte("test1"),
			Session: sess1,
		}, nil)
		assert.NoError(t, err)
		assert.True(t, ok)
	})
	t.Run("ensure value", func(t *testing.T) {
		res, _, err := cli.KV().Get("test", nil)
		assert.NoError(t, err)
		assert.Equal(t, []byte("test1"), res.Value)
	})
	t.Run("txn acquire", func(t *testing.T) {
		ok, _, _, err := cli.KV().Txn(api.KVTxnOps{
			&api.KVTxnOp{
				Verb:    api.KVLock,
				Key:     "test1",
				Session: sess1,
				Value:   []byte("test"),
			},
		}, nil)
		assert.NoError(t, err)
		assert.True(t, ok)
	})
	t.Run("txn ensure", func(t *testing.T) {
		res, _, err := cli.KV().Get("test1", nil)
		assert.NoError(t, err)
		assert.Equal(t, []byte("test"), res.Value)
	})
	t.Run("release by delete", func(t *testing.T) {
		ok, _, _, err := cli.KV().Txn(api.KVTxnOps{
			&api.KVTxnOp{
				Verb: api.KVDelete,
				Key:  "test1",
			},
		}, nil)
		assert.NoError(t, err)
		assert.True(t, ok)
	})
	t.Run("txn ensure", func(t *testing.T) {
		res, _, err := cli.KV().Get("test1", nil)
		assert.NoError(t, err)
		assert.Nil(t, res)
	})
	t.Run("txn delete non-existent", func(t *testing.T) {
		ok, _, _, err := cli.KV().Txn(api.KVTxnOps{
			&api.KVTxnOp{
				Verb: api.KVDelete,
				Key:  "test1",
			},
		}, nil)
		assert.NoError(t, err)
		assert.True(t, ok)
	})
}
