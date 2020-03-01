package storage

import (
	"github.com/futurehomeno/fimpgo/fimptype/primefimp"
	"testing"
)

func TestVinculumRegistryStore_GetAllDevices(t *testing.T) {
			r := &VinculumRegistryStore{}
			r.vApi = primefimp.NewApiClient("tpflow_reg",nil,false)
			err := r.vApi.LoadVincResponseFromFile("../../testdata/vinfimp/site-response.json")
			if err != nil {
				t.Fatal("Can't load site data from file . Error:",err)
			}
			devices, err := r.GetAllDevices()
			if devices == nil {
				t.Fatalf("No devices found")
			}
			for _,dev := range devices {
				t.Logf( "Device name = %s , type = %s",dev.Alias,dev.Type)
			}
}

func TestVinculumRegistryStore_GetAllThings(t *testing.T) {
	r := &VinculumRegistryStore{}
	r.vApi = primefimp.NewApiClient("tpflow_reg",nil,false)
	err := r.vApi.LoadVincResponseFromFile("../../testdata/vinfimp/site-response.json")
	if err != nil {
		t.Fatal("Can't load site data from file . Error:",err)
	}
	things, err := r.GetAllThings()
	if things == nil {
		t.Fatalf("No things found")
	}
	for _,dev := range things {
		t.Logf( "Thing name = %s , commTech = %s, location = %d, power=%s",dev.Alias,dev.CommTechnology,dev.LocationId,dev.PowerSource)
	}
}

func TestVinculumRegistryStore_GetExtendedDevices(t *testing.T) {
	r := &VinculumRegistryStore{}
	r.vApi = primefimp.NewApiClient("tpflow_reg",nil,false)
	err := r.vApi.LoadVincResponseFromFile("../../testdata/vinfimp/site-response.json")
	if err != nil {
		t.Fatal("Can't load site data from file . Error:",err)
	}
	devices, err := r.GetExtendedDevices()
	if devices == nil {
		t.Fatalf("No devices found")
	}
	for _,dev := range devices {
		t.Logf( "Device name = %s , type = %s,location = %s",dev.Alias,dev.Type,dev.LocationAlias)
	}
}