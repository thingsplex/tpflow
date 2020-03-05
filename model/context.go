package model

import (
	"bytes"
	"encoding/gob"
	log "github.com/sirupsen/logrus"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
	"math"
	"time"
)

type Variable struct {
	Value     interface{}
	ValueType string
}

func (vrbl *Variable) IsNumber() bool {
	if vrbl.ValueType == "int" || vrbl.ValueType == "float" {
		return true
	} else {
		return false
	}
}

func (vrbl *Variable) IsEqual(var2 *Variable) (bool, error) {
	if vrbl.ValueType == var2.ValueType {
		switch vrbl.ValueType {
		case "string":
			v1, ok1 := vrbl.Value.(string)
			v2, ok2 := var2.Value.(string)
			if ok1 && ok2 {
				return v1 == v2, nil
			} else {
				return false, errors.New("Can't cast var to string")
			}
		case "int", "float":
			v1, ok1 := vrbl.ToNumber()
			v2, ok2 := var2.ToNumber()
			if ok1 == nil && ok2 == nil {
				return v1 == v2, nil
			} else {
				return false, errors.New("Can't cast var to number")
			}
		case "bool":
			v1, ok1 := vrbl.Value.(bool)
			v2, ok2 := var2.Value.(bool)
			if ok1 && ok2 {
				return v1 == v2, nil
			} else {
				return false, errors.New("Can't cast var to bool")
			}

		}
	} else {
		return false, errors.New("Types are different")
	}
	return false, nil
}

// Validates actual Variable type against type set in VariableType
func (vrbl *Variable) isTypeValid() bool {
	switch vrbl.ValueType {
	case "int":
		switch vrbl.Value.(type) {
		case int, int8, int16, int32, int64, float32, float64:
			return true
		}
		return false
	case "float":
		switch v := vrbl.Value.(type) {
		case float32:
			return true
		case float64:
			if !math.IsNaN(v) {
				return true
			}
		}
		return false
	case "bool":
		switch vrbl.Value.(type) {
		case bool:
			return true
		}
		return false
	case "string":
		switch vrbl.Value.(type) {
		case string:
			return true
		}
		return false
	}

	return true
}

func (vrbl *Variable) ToNumber() (float64, error) {
	switch v := vrbl.Value.(type) {
	case int:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case float64:
		return float64(v), nil
	default:
		return 0, errors.New("Can't convert into float")
	}
	return 0, errors.New("Not numeric value type")
}

type ContextRecord struct {
	Name        string
	Description string
	UpdatedAt   time.Time
	Variable    Variable
	InMemory    bool
}

type Context struct {
	storageLocation string
	db              *bolt.DB
	inMemoryStore   ContextInMemoryStore
}

func NewContextDB(storageLocation string) (*Context, error) {
	var err error
	gob.Register(map[string]interface{}{})
	ctx := Context{}
	//ctx.inMemoryStore = make(map[string][]ContextRecord)
	ctx.db, err = bolt.Open(storageLocation, 0600, nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	ctx.RegisterFlow("global")
	//ctx.DeleteRecord("weather.temp", "global", false)
	return &ctx, nil
}
func (ctx *Context) Close() {
	ctx.db.Close()
}
func (ctx *Context) RegisterFlow(flowId string) error {
	ctx.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(flowId))
		if err != nil {
			log.Errorf("<ctx> Can't create bucket %s . Error: %s", flowId, err)
			return err
		}
		return nil
	})
	log.Infof("<ctx> Flow %s is registered in store.", flowId)
	return nil
}

func (ctx *Context) UnregisterFlow(flowId string) error {
	ctx.db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket([]byte(flowId))
		if err != nil {
			log.Errorf("<ctx> Can't delete bucket %s . Error: %s", flowId, err)
			return err
		}
		return nil
	})
	ctx.inMemoryStore.DeleteFlow(flowId)
	log.Info("<ctx> Flow %s is deleted .", flowId)
	return nil
}

func (ctx *Context) SetVariable(name string, valueType string, value interface{}, description string, flowId string, inMemory bool) error {
	rec := ContextRecord{Name: name, UpdatedAt: time.Now(), Description: description, Variable: Variable{ValueType: valueType, Value: value}}
	return ctx.PutRecord(&rec, flowId, inMemory)
}

func (ctx *Context) PutRecord(rec *ContextRecord, flowId string, inMemory bool) error {
	//log.Infof("<ctx> Saving variable ,in flow %s , type = %s , ", flowId, rec.Variable.ValueType, rec.Variable.Value)
	if !rec.Variable.isTypeValid() {
		log.Info("<ctx> Incompatible type , varname = ", rec.Name)
		return errors.New("Incompatible type")
	}
	if inMemory {
		ctx.inMemoryStore.Store(*rec, flowId)
	} else {
		err := ctx.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(flowId))
			data, err := ctx.encodeRecord(rec)
			if err != nil {
				return err
			}
			err = b.Put([]byte(rec.Name), data)
			return err
		})
		return err
	}
	return nil
}

func (ctx *Context) DeleteRecord(name string, flowId string, inMemory bool) error {
	if inMemory {

	} else {
		log.Infof("<ctx> Deleting variable %s from flow %s", name, flowId)
		err := ctx.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(flowId))
			return b.Delete([]byte(name))
		})
		return err
	}
	return nil
}

func (ctx *Context) GetVariable(name string, flowId string) (Variable, error) {
	rec, err := ctx.GetRecord(name, flowId)
	if err == nil {
		return rec.Variable, err
	} else {
		return Variable{}, err
	}

}

func (ctx *Context) GetVariableType(name string, flowId string) (string, error) {
	varb, err := ctx.GetVariable(name, flowId)
	if err == nil {
		return varb.ValueType, err
	} else {
		return "", err
	}
}

func (ctx *Context) GetRecord(name string, flowId string) (*ContextRecord, error) {
	// check memmory first
	rec := ctx.inMemoryStore.Get(flowId, name)
	if rec != nil {
		rec.InMemory = true
		return rec, nil
	}

	var ctxRec *ContextRecord
	var err error
	ctx.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(flowId))
		if b == nil {
			err = errors.New("Flow doesn't exist")
			return nil
		}
		data := b.Get([]byte(name))
		if data == nil {
			err = errors.New("Not Found")
		} else {
			ctxRec, err = ctx.decodeRecord(data)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ctxRec, err
}

func (ctx *Context) GetRecords(flowId string) []ContextRecord {
	result := []ContextRecord{}
	log.Info("<ctx> GEtting records")
	ctx.db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(flowId))
		if b == nil {

			return nil
		}
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			rec, err := ctx.decodeRecord(v)
			if err == nil {
				result = append(result, *rec)
			} else {
				log.Errorf("Can't decode record = %s , %s", k, err)
			}

		}

		return nil
	})
	//log.Info("<ctx> DONE ,",len(result))
	//log.Info("<ctx> DONE ,",result)
	//
	memResult, err := ctx.inMemoryStore.GetRecordsForFlow(flowId)
	if err == nil {
		for i := range memResult {
			memResult[i].InMemory = true
		}
		result = append(result, memResult...)
	} else {
		log.Error("Can't get records from memory store , err :", err)
	}
	return result
}

func (ctx *Context) encodeRecord(rec *ContextRecord) ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(rec)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (ctx *Context) decodeRecord(data []byte) (*ContextRecord, error) {
	ctxRec := ContextRecord{}
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&ctxRec)
	//log.Infof("Decoding variable %s",ctxRec.Name)
	//if ctxRec.Name == "weather.temp"{
	//	return nil,errors.New("WrongType")
	//
	//}

	return &ctxRec, err
}

//func (ctx *Context) GetRecord(name string) (*ContextRecord,error) {
//	rec , ok := ctx.records[name]
//	if ok {
//		return &rec,nil
//	}
//	return &ContextRecord{},errors.New("Variable doesn't exist")
//}
//
