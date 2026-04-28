package api

import (
	"encoding/json"
	"fmt"
)

// UpstreamNodes accepts both APISIX upstream node forms:
// {"host:port": weight} and [{"host":"...","port":...,"weight":...}].
type UpstreamNodes map[string]interface{}

func (n *UpstreamNodes) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*n = nil
		return nil
	}

	var keyed map[string]interface{}
	if err := json.Unmarshal(data, &keyed); err == nil {
		*n = keyed
		return nil
	}

	var listed []struct {
		Host   string          `json:"host"`
		Port   int             `json:"port"`
		Weight json.RawMessage `json:"weight"`
	}
	if err := json.Unmarshal(data, &listed); err != nil {
		return err
	}

	nodes := make(UpstreamNodes, len(listed))
	for _, node := range listed {
		key := node.Host
		if node.Port != 0 {
			key = fmt.Sprintf("%s:%d", node.Host, node.Port)
		}

		var weight interface{} = 1
		if len(node.Weight) > 0 && string(node.Weight) != "null" {
			if err := json.Unmarshal(node.Weight, &weight); err != nil {
				return err
			}
		}
		nodes[key] = weight
	}

	*n = nodes
	return nil
}
