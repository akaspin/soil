package cluster_test

import (
	"fmt"
	"github.com/hashicorp/memberlist"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
	"github.com/stretchr/testify/require"
	"github.com/davecgh/go-spew/spew"
	"sync"
)

type fakeNode struct {
	*memberlist.Memberlist
	name string
	meta string
	metaMu *sync.Mutex
}

func (d *fakeNode) SetMeta(v string) {
	d.metaMu.Lock()
	defer d.metaMu.Unlock()
	d.meta = v
}

func (d *fakeNode) NotifyAlive(peer *memberlist.Node) (err error) {
	//println(fmt.Sprintf(">>> alive %s %v", d.name, peer.Name))
	return
}

func (d *fakeNode) NotifyMerge(peers []*memberlist.Node) (err error) {
	var names []string
	for _, n := range peers {
		names = append(names, n.Name)
	}
	println(fmt.Sprintf(">>> merge %s %v", d.name, names))
	return
}

func (d *fakeNode) NotifyJoin(node *memberlist.Node) {
	println(fmt.Sprintf(">>> event join %s %s %s", d.name, node.Name, string(node.Meta)))
}

func (d *fakeNode) NotifyLeave(node *memberlist.Node) {
	println(fmt.Sprintf(">>> event leave %s %s %s", d.name, node.Name, string(node.Meta)))
}

func (d *fakeNode) NotifyUpdate(node *memberlist.Node) {
	println(fmt.Sprintf(">>> event update %s %s %s", d.name, node.Name, string(node.Meta)))
}

func (d *fakeNode) NodeMeta(limit int) (res []byte) {
	println(fmt.Sprintf(">>> delegate NodeMeta %s %d", d.name, limit))
	d.metaMu.Lock()
	defer d.metaMu.Unlock()
	res = []byte(d.meta)
	return
}

func (d *fakeNode) NotifyMsg(data []byte) {
	println(fmt.Sprintf(">>> delegate NotifyMsg %s %s", d.name, string(data)))
}

func (d *fakeNode) GetBroadcasts(overhead, limit int) (res [][]byte) {
	//println(fmt.Sprintf(">>> delegate GetBroadcasts %s %d %d", d.name, overhead, limit))
	return
}

func (d *fakeNode) LocalState(join bool) (res []byte) {
	println(fmt.Sprintf(">>> delegate LocalState %s %t", d.name, join))
	d.metaMu.Lock()
	defer d.metaMu.Unlock()
	res = []byte(d.meta)
	return
}

func (d *fakeNode) MergeRemoteState(buf []byte, join bool) {
	println(fmt.Sprintf(">>> delegate MergeRemoteState %s %s %t", d.name, string(buf), join))
	return
}

func createNodes(n, base int) (nodes []*fakeNode, err error) {
	for i := base; i < base + n; i++ {

		config := memberlist.DefaultLocalConfig()
		config.Name = fmt.Sprintf("node-%d", i)
		config.AdvertiseAddr = "127.0.0.1"
		config.BindPort = i
		config.AdvertisePort = i


		fn := &fakeNode{
			name: config.Name,
			meta: config.Name + "-META-1",
			metaMu: &sync.Mutex{},
		}
		config.Alive = fn
		config.Merge = fn
		config.Events = fn
		config.Delegate = fn

		var node *memberlist.Memberlist

		if node, err = memberlist.Create(config); err != nil {
			return
		}
		fn.Memberlist = node
		nodes = append(nodes, fn)
	}
	return
}

func TestMemberlist(t *testing.T) {
	t.Skip()
	lists, err := createNodes(10, 9000)
	require.NoError(t, err)
	for _, list := range lists {
		n, err := list.Join([]string{"127.0.0.1:9000"})
		assert.Equal(t, 1, n)
		assert.NoError(t, err)
	}
	time.Sleep(time.Millisecond * 500)
	for _, list := range lists {
		assert.Equal(t, 10, list.NumMembers())
	}
	for _, member := range lists[3].Members() {
		t.Log(member.Meta)
	}
}

func TestMemberlistJoin(t *testing.T) {
	t.Run("join fail", func(t *testing.T) {
		t.Skip()
		nodes1, err := createNodes(1, 9000)
		require.NoError(t, err)
		// join with fail
		n, err := nodes1[0].Join([]string{"127.0.0.1:9100"})
		assert.Equal(t, n, 0)
		assert.Error(t, err)

		// create second node
		nodes2, err := createNodes(1, 9100)
		require.NoError(t, err)

		// try to join again
		n, err = nodes1[0].Join([]string{"127.0.0.1:9100"})
		assert.Equal(t, n, 1)
		assert.NoError(t, err)

		// stupidly rejoin
		time.Sleep(time.Millisecond * 300)
		n, err = nodes1[0].Join([]string{"127.0.0.1:9100"})
		assert.Equal(t, n, 1)
		assert.NoError(t, err)

		nodes1[0].Shutdown()
		nodes2[0].Shutdown()
	})
	t.Run("rejoin", func(t *testing.T) {
		t.Skip()
		nodes1, err := createNodes(1, 9000)
		require.NoError(t, err)

		nodes2, err := createNodes(1, 9100)
		require.NoError(t, err)

		_, err = nodes1[0].Join([]string{nodes2[0].LocalNode().Address()})
		assert.NoError(t, err)
		time.Sleep(time.Millisecond * 300)

		err = nodes2[0].Shutdown()
		assert.NoError(t, err)

		time.Sleep(time.Second)

		nodes2, err = createNodes(1, 9100)
		require.NoError(t, err)

		time.Sleep(time.Second * 30)

		spew.Dump(nodes1[0].Members())
		spew.Dump(nodes2[0].Members())

		// shutdown again
		err = nodes2[0].Shutdown()
		assert.NoError(t, err)
		time.Sleep(time.Second * 60)

		nodes2, err = createNodes(1, 9100)
		require.NoError(t, err)

		spew.Dump(nodes1[0].Members())
		spew.Dump(nodes2[0].Members())
	})
	t.Run("update", func(t *testing.T) {
		t.Skip()
		nodes, err := createNodes(3, 9000)
		require.NoError(t, err)
		for _, node := range nodes {
			_, err = node.Join([]string{nodes[0].LocalNode().Address()})
			require.NoError(t, err)
		}
		time.Sleep(time.Second)

		nodes[0].SetMeta(nodes[0].name + "-META-2")
		nodes[0].UpdateNode(time.Second * 10)

		time.Sleep(time.Second * 5)
	})
}