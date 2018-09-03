package utils

import "testing"

func TestGenerateId(t *testing.T) {
	t.Log(GenerateId(12))
}

func TestRouteIncludesTopic(t *testing.T) {
	route := "pt:j1/mt:evt/rt:dev/rn:zw/#"
	topic := "pt:j1/mt:evt/rt:dev/rn:zw/ad:1/sv:sensor_presence/ad:234"
	if ! RouteIncludesTopic(route,topic) {
		t.Fail()
	}
}
