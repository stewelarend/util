package util_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/stewelarend/util"
)

func TestDecoder(t *testing.T) {
	a := MyStruct{
		Name: "Jan",
		Age:  15,
	}
	data := a.Encode()

	_b, err := util.StructDecode(&MyStruct{}, data)
	if err != nil {
		t.Fatalf("failed to decode")
	}
	b := _b.(*MyStruct)
	if b.Name != a.Name || b.Age != a.Age {
		t.Fatalf("%+v != %+v", a, b)
	}
	t.Logf("Success")
}

type MyStruct struct {
	Name string
	Age  int
}

func (d *MyStruct) Decode(data []byte) error {
	part := strings.SplitN(string(data), ",", 2)
	d.Name = part[0]
	age, _ := strconv.ParseInt(part[1], 10, 64)
	d.Age = int(age)
	return nil
}

func (d MyStruct) Encode() []byte {
	return []byte(fmt.Sprintf("%s,%d", d.Name, d.Age))
}
