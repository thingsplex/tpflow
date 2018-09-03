package model

import "testing"

func TestContext_SetVariable(t *testing.T) {
	ctx ,err := NewContextDB("context_test_1.db")
	if err != nil {
		t.Error("Fail to create context ",err)
	}
	ctx.SetVariable("mode","string","away","home mode","global",false)
	err = ctx.SetVariable("temp","float",nil,"","global",false)
	if err == nil {
		t.Error("Type validation doesn't work")
	}else {
		t.Log("Type validation works")
	}
	ctx.Close()

	ctx ,err = NewContextDB("context_test_1.db")
	if err != nil {
		t.Error("Fail to create context ",err)
	}
	rec ,err := ctx.GetRecord("mode","global")
	if err != nil {
		t.Error("Fail to create context ",err)
	}
	ctx.Close()

	if rec.Variable.Value.(string) == "away" {
		t.Log("OK ,value is = ",rec.Variable.Value)
	}else {
		t.Error("Wrong result")
	}
}

func TestContext_GetRecords(t *testing.T) {
	ctx ,err := NewContextDB("context_test_2.db")
	if err != nil {
		t.Error("Fail to create context ",err)
	}
	ctx.SetVariable("temp_1","float",33.5,"","global",false)
	ctx.SetVariable("temp_2","float",44.5,"","global",false)
	ctx.SetVariable("mode","string","away","","global",false)
	ctx.Close()

	ctx ,err = NewContextDB("context_test_2.db")
	if err != nil {
		t.Error("Fail to create context ",err)
	}

	records := ctx.GetRecords("global")
	for _,rec := range records {
		t.Logf(" Record name = %s , type = %s ,value = %s",rec.Name,rec.Variable.ValueType,rec.Variable.Value)
		if rec.Name == "temp_2" {
			if rec.Variable.Value.(float64) != 44.5 {
				t.Error("GetRecords is broken")
			}
		}
	}
	ctx.Close()

}


