package noowner

import (
	"testing"

	"github.com/FreifunkBremen/yanic/tree/master/data"
	"github.com/FreifunkBremen/yanic/tree/master/runtime"
	"github.com/stretchr/testify/assert"
)

func TestFilter(t *testing.T) {
	assert := assert.New(t)

	// invalid config
	filter, err := build("nope")
	assert.Error(err)

	// delete owner by configuration
	filter, _ = build(true)
	n := filter.Apply(&runtime.Node{Nodeinfo: &data.NodeInfo{
		Owner: &data.Owner{
			Contact: "blub",
		},
	}})

	assert.NotNil(n)
	assert.Nil(n.Nodeinfo.Owner)

	// keep owner configuration
	filter, _ = build(false)
	n = filter.Apply(&runtime.Node{Nodeinfo: &data.NodeInfo{
		Owner: &data.Owner{
			Contact: "blub",
		},
	}})

	assert.NotNil(n)
	assert.NotNil(n.Nodeinfo.Owner)
}
