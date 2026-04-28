package api

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// UpstreamNodes accepts both APISIX upstream node forms:
// {"host:port": weight} and [{"host":"...","port":...,"weight":...}].
type UpstreamNodes map[string]interface{}

func (n *UpstreamNodes) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if string(trimmed) == "null" {
		*n = nil
		return nil
	}

	var keyed map[string]interface{}
	if err := json.Unmarshal(trimmed, &keyed); err == nil {
		*n = keyed
		return nil
	}

	var listed []struct {
		Host   string          `json:"host"`
		Port   int             `json:"port"`
		Weight json.RawMessage `json:"weight"`
	}
	if err := json.Unmarshal(trimmed, &listed); err != nil {
		return err
	}

	nodes := make(UpstreamNodes, len(listed))
	for _, node := range listed {
		key := node.Host
		if node.Port != 0 {
			key = fmt.Sprintf("%s:%d", node.Host, node.Port)
		}

		var weight interface{} = float64(1)
		weightData := bytes.TrimSpace(node.Weight)
		if len(weightData) > 0 && string(weightData) != "null" {
			if err := json.Unmarshal(weightData, &weight); err != nil {
				return err
			}
		}
		nodes[key] = weight
	}

	*n = nodes
	return nil
}
