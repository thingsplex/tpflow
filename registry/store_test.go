package registry

import (
	"testing"
	"os"
)

func TestThingRegistryStore_UpsertThing(t *testing.T) {
	dbFileName := "testStore1.db"
	thing := Thing{Alias:"Super mega thing",Address:"1234"}
	service := Service{Address:"service/1",Name:"dev_sys"}

	st := NewThingRegistryStore(dbFileName)
	newID ,err := st.UpsertThing(&thing)
	if err != nil {
		t.Error("Can't upsert Thing. Error:",err)
		t.Fail()
	}
	t.Log("New id = ",newID)

	service.ParentContainerId = newID
	service.ParentContainerType = ThingContainer

	newServiceID ,err := st.UpsertService(&service)
	if err != nil {
		t.Error("Can't upsert Service. Error:",err)
		t.Fail()
	}
	t.Log("New id = ",newServiceID)


	if err != nil {
		t.Error("Can't get Thing. Error:",err)
		t.Fail()
	}
	thing2 , err := st.GetThingExtendedViewById(newID)
	if thing2.Services[0].Name != "dev_sys" {
		t.Error("Wrong value")
		t.Fail()
	}

	newID ,err = st.UpsertThing(&thing)
	if err != nil {
		t.Error("Can't upsert Thing. Error:",err)
		t.Fail()
	}
	t.Log(" Second update New id = ",newID)
	st.Disconnect()
	os.Remove(dbFileName)

}

