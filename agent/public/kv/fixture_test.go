// +build ide test_cluster

package kv_test

import (
	"fmt"
	"github.com/akaspin/soil/agent/metadata"
	"github.com/hashicorp/consul/testutil"
	"github.com/nu7hatch/gouuid"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

type consulFixture struct {
	Dir      string
	Name     string
	Ports    *testutil.TestPortConfig
	Server   *testutil.TestServer
	t        *testing.T
	LogLevel string
}

func newConsulFixture(t *testing.T) (f *consulFixture) {
	t.Helper()
	f = &consulFixture{
		Dir:      "testdata/.consul_data_" + strings.Replace(t.Name(), "/", "_", -1),
		Name:     strings.Replace(t.Name(), "/", "-", -1),
		t:        t,
		LogLevel: "ERR",
	}
	os.RemoveAll(f.Dir)
	f.Ports = &testutil.TestPortConfig{
		DNS:     randomPort(),
		HTTP:    randomPort(),
		HTTPS:   randomPort(),
		SerfLan: randomPort(),
		SerfWan: randomPort(),
		Server:  randomPort(),
		RPC:     randomPort(),
	}
	os.MkdirAll(f.Dir, 0777)
	ioutil.WriteFile(f.Dir+"/ttl.json", []byte(`{"session_ttl_min":".5s"}`), 0777)
	f.Start()
	return
}

func (f *consulFixture) Start() (err error) {
	f.t.Helper()
	if f.Server != nil {
		return
	}
	f.Server, err = testutil.NewTestServerConfigT(f.t, func(c *testutil.TestServerConfig) {
		c.Ports = f.Ports
		c.DataDir = f.Dir
		c.NodeName = f.Name
		sUUID, _ := uuid.NewV5(uuid.NamespaceOID, []byte(c.NodeName))
		c.NodeID = sUUID.String()
		c.LogLevel = f.LogLevel
		c.Stdout = &nullWriter{}
		c.Stderr = &nullWriter{}
		c.Args = []string{fmt.Sprintf("-config-file=%s/ttl.json", f.Dir)}
	})
	if err != nil {
		f.t.FailNow()
	}
	return
}

func (f *consulFixture) Stop() {
	if f.Server == nil {
		return
	}
	f.Server.Stop()
	f.Server = nil
}

func (f *consulFixture) Restart(wrap, between time.Duration) {
	time.Sleep(wrap)
	f.Stop()
	time.Sleep(between)
	f.Start()
	time.Sleep(wrap)
}

func randomPort() int {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

type nullWriter struct{}

func (*nullWriter) Write(p []byte) (n int, err error) {
	n = len(p)
	return
}

type dummyConsumer struct {
	changes int
	res     []map[string]string
	//states  []bool
	mu *sync.Mutex
}

func newDummyConsumer() (c *dummyConsumer) {
	c = &dummyConsumer{
		mu: &sync.Mutex{},
	}
	return
}

func (c *dummyConsumer) ConsumeMessage(message metadata.Message) {
	c.mu.Lock()
	defer c.mu.Unlock()
	//c.states = append(c.states, message.Clean)
	c.res = append(c.res, message.Data)
}
