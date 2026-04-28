package api

import (
	"encoding/json"
	"testing"
)

func TestUpstreamNodesUnmarshalObject(t *testing.T) {
	var nodes UpstreamNodes
	if err := json.Unmarshal([]byte(`{"127.0.0.1:9080": 1}`), &nodes); err != nil {
		t.Fatalf("unmarshal object nodes: %v", err)
	}

	if got := nodes["127.0.0.1:9080"]; got != float64(1) {
		t.Fatalf("unexpected node weight: %#v", got)
	}
}

func TestUpstreamNodesUnmarshalArray(t *testing.T) {
	var nodes UpstreamNodes
	if err := json.Unmarshal([]byte(`[{"host":"127.0.0.1","port":9080,"weight":2}]`), &nodes); err != nil {
		t.Fatalf("unmarshal array nodes: %v", err)
	}

	if got := nodes["127.0.0.1:9080"]; got != float64(2) {
		t.Fatalf("unexpected node weight: %#v", got)
	}
}

func TestRouteWithArrayUpstreamNodesUnmarshals(t *testing.T) {
	var route Route
	err := json.Unmarshal([]byte(`{
		"id": "route-with-array-nodes",
		"uri": "/get",
		"upstream": {
			"type": "roundrobin",
			"nodes": [{"host":"127.0.0.1","port":9080,"weight":1}]
		}
	}`), &route)
	if err != nil {
		t.Fatalf("unmarshal route: %v", err)
	}

	if route.Upstream == nil {
		t.Fatal("expected upstream")
	}
	if got := route.Upstream.Nodes["127.0.0.1:9080"]; got != float64(1) {
		t.Fatalf("unexpected route node weight: %#v", got)
	}
}
