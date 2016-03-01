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
	LIntVal []int
	LStrVal []string
	LSrtVal []tsA
}

type tsC struct {
	MIntStr map[int]string
	MStrSrt map[string]tsA
}

type tsD struct {
	tsA
	tsB
	tsC
}

type tsE struct{
	SrtA tsA
	SrtB tsB
	SrtC tsC
}

type tsF struct {
	tsD
	tsE
}

type tsG struct {
	SrtD tsD
	SrtE tsE
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


