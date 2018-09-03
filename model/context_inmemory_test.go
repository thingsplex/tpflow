package model

import (
	"reflect"
	"sync"
	"testing"
)

func TestContextInMemoryStore_Store(t *testing.T) {

	v := Variable{ValueType:"float",Value:nil}
	if v.isTypeValid(){
		t.Error("Wrong type")
		t.Fail()
	}

	rec := ContextRecord{Name:"var1",Variable:Variable{ValueType:"string",Value:"a1"}}
	ctx := ContextInMemoryStore{}
	ctx.Store(rec,"global")
	rec2 := ContextRecord{Name:"var2",Variable:Variable{ValueType:"int",Value:55}}
	ctx.Store(rec2,"global")
	result,_ := ctx.GetRecordsForFlow("global")
	if len(result)!=2 {
		t.Error("Error")
	}
	t.Log(len(result))
	t.Log(ctx.Get("global","var1"))
}

func TestContextInMemoryStore_Get(t *testing.T) {
	type fields struct {
		store sync.Map
	}
	type args struct {
		flowId  string
		varName string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *ContextRecord
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ContextInMemoryStore{
				store: tt.fields.store,
			}
			if got := ctx.Get(tt.args.flowId, tt.args.varName); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ContextInMemoryStore.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContextInMemoryStore_DeleteFlow(t *testing.T) {
	type fields struct {
		store sync.Map
	}
	type args struct {
		flowId string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ContextInMemoryStore{
				store: tt.fields.store,
			}
			ctx.DeleteFlow(tt.args.flowId)
		})
	}
}

func TestContextInMemoryStore_GetRecordsForFlow(t *testing.T) {
	type fields struct {
		store sync.Map
	}
	type args struct {
		flowId string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []ContextRecord
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ContextInMemoryStore{
				store: tt.fields.store,
			}
			got, err := ctx.GetRecordsForFlow(tt.args.flowId)
			if (err != nil) != tt.wantErr {
				t.Errorf("ContextInMemoryStore.GetRecordsForFlow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ContextInMemoryStore.GetRecordsForFlow() = %v, want %v", got, tt.want)
			}
		})
	}
}
