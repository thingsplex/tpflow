package storage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type BaseStore struct {
	thingDescPath string
}

func NewBaseStore(thingDescPath string) *BaseStore {
	return &BaseStore{thingDescPath: thingDescPath}
}

func (b *BaseStore) SaveThingDescriptor(address string,inclusionReport interface{})error {

	//address = strings.ReplaceAll(address,"/rn:","")
	//address = strings.ReplaceAll(address,"/ad:","_")

	fullPath := fmt.Sprintf("%s/%s.json",b.thingDescPath,address)
	bReport,err := json.Marshal(inclusionReport)
	if err != nil {
		return err
	}
	ioutil.WriteFile(fullPath,bReport,777)
	return nil
}

func (b *BaseStore) DeleteThingDescriptor(address string)error {
	return nil
}

func (b *BaseStore) GetThingDescriptor(address string)error {
	//fullPath := fmt.Sprintf("%s/%s.json",b.thingDescPath,address)
	//bReport , err  := ioutil.ReadFile(fullPath)
	//if err != nil {
	//	return err
	//}
	return nil
}