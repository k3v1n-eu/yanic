package blacklist

import (
	"errors"

	"github.com/FreifunkBremen/yanic/tree/master/output/filter"
	"github.com/FreifunkBremen/yanic/tree/master/runtime"
)

type blacklist map[string]interface{}

func init() {
	filter.Register("blacklist", build)
}

func build(config interface{}) (filter.Filter, error) {
	values, ok := config.([]interface{})
	if !ok {
		return nil, errors.New("invalid configuration, array (of strings) expected")
	}

	list := make(blacklist)
	for _, value := range values {
		if nodeid, ok := value.(string); ok {
			list[nodeid] = struct{}{}
		} else {
			return nil, errors.New("invalid configuration, array of strings expected")
		}
	}
	return &list, nil
}

func (list blacklist) Apply(node *runtime.Node) *runtime.Node {
	if nodeinfo := node.Nodeinfo; nodeinfo != nil {
		if _, ok := list[nodeinfo.NodeID]; ok {
			return nil
		}
	}
	return node
}
