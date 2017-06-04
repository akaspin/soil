package agent_test

import (
	"fmt"
	"github.com/hashicorp/memberlist"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMemberlist(t *testing.T) {
	var lists []*memberlist.Memberlist
	for i := 9000; i < 9010; i++ {
		config := memberlist.DefaultLocalConfig()
		config.Name = fmt.Sprintf("node-%d", i)
		config.AdvertiseAddr = "127.0.0.1"
		config.BindPort = i
		config.AdvertisePort = i

		list, err := memberlist.Create(config)
		assert.NoError(t, err)
		lists = append(lists, list)
	}
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

func TestMemberlistDelegate(t *testing.T) {

}