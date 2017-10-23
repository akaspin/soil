// +build ide test_unit

package lib_test

import (
	"testing"
	"github.com/akaspin/soil/lib"
	"io/ioutil"
	"github.com/stretchr/testify/assert"
)

func TestStaticBuffers_GetReaders(t *testing.T) {
	var buffers lib.StaticBuffers

	buffers = append(buffers, [][]byte{
		[]byte(`1`),
		[]byte(`2`),
	}...)

	var res []string
	for _, r := range buffers.GetReaders() {
		res1, err := ioutil.ReadAll(r)
		assert.NoError(t, err)
		res = append(res, string(res1))
	}
	assert.Equal(t, []string{"1", "2"}, res)
}

func TestStaticBuffers_ReadFiles(t *testing.T) {
	var buffers lib.StaticBuffers
	err := buffers.ReadFiles(
		"testdata/TestStaticBuffers_ReadFiles_0.txt",
		"testdata/TestStaticBuffers_ReadFiles_1.txt",
	)
	assert.NoError(t, err)
	var res []string
	for _, r := range buffers.GetReaders() {
		res1, _ := ioutil.ReadAll(r)
		res = append(res, string(res1))
	}
	assert.Equal(t, []string{"0", "1"}, res)
}