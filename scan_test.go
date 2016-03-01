package gomdies

import (
	"testing"
	"reflect"
)


type tsA struct {
	IntVal int
	StrVal string
	BytVal byte
}

type tsB struct {
	MStrSrt map[string]tsA
	LSrtVal	[]tsA
}

type tsC struct {
	MStrStrSrt map[string]map[string]*tsA
	LLSrtVal [][]*tsA
}

type tsD struct {
	tsB
	tsC
}

type tsE struct {
	TsB tsB
	TsC tsC
}

func MakeTsA() *tsA {
	return &tsA{
		10,
		"tsA",
		0x55,
	}
}

func MakeTsB() *tsB {
	tsB := &tsB{
		make(map[string]tsA),
		make([]tsA, 2, 2),
	}
	tsB.MStrSrt["TsA0"] = *MakeTsA()
	tsB.MStrSrt["TsA1"] = *MakeTsA()
	tsB.LSrtVal[0] = *MakeTsA()
	tsB.LSrtVal[1] = *MakeTsA()
	return tsB
}

func MakeTsC() *tsC {
	tsC := &tsC{
		make(map[string]map[string]*tsA),
		make([][]*tsA, 2),
	}
	tsC.MStrStrSrt["map0"] = make(map[string]*tsA)
	tsC.MStrStrSrt["map0"]["tsA"] = MakeTsA()
	tsC.MStrStrSrt["map1"] = make(map[string]*tsA)
	tsC.MStrStrSrt["map1"]["tsA"] = MakeTsA()

	tsC.LLSrtVal[0] = make([]*tsA, 2)
	tsC.LLSrtVal[0][0] = MakeTsA()
	tsC.LLSrtVal[0][1] = MakeTsA()
	tsC.LLSrtVal[1] = make([]*tsA, 2)
	tsC.LLSrtVal[1][0] = MakeTsA()
	tsC.LLSrtVal[1][1] = MakeTsA()
	return tsC
}

func TestBaseStruct(t *testing.T) {
	specA := structSpecForType(reflect.TypeOf(tsA{}))
	if len(specA.fields) != 3 {
		t.Fail()
	}
	if specA.fields[0].typ != reflect.TypeOf(int(0)) {
		t.Fail()
	}
	if specA.fields[1].typ != reflect.TypeOf(string("")) {
		t.Fail()
	}
	if specA.fields[2].typ != reflect.TypeOf(byte(0)) {
		t.Fail()
	}
}

func TestEmbeddedStruct(t *testing.T) {

}

func TestNestedStruct(t *testing.T) {

}

func TestMultipleEmbeddedStruct(t *testing.T) {

}

func TestMultipleNestedStruct(t *testing.T) {

}


