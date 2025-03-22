package api

import (
	"fmt"
	"testing"
)

func TestRestGetVip3Level(t *testing.T) {
	l, err := RestGetVip3Level("0x0aB6527027EcFF1144dEc3d78154fce309ac838c")
	if err != nil {
		fmt.Print(err.Error())
		t.FailNow()
	}
	fmt.Print(l)
	l, err = RestGetVip3Level("0xeddcfdf45d384f7b4e6722e55a92acd7c7dd27e1")
	if err != nil {
		fmt.Print(err.Error())
		t.FailNow()
	}
	fmt.Print(l)
}
func TestStrToFeeArray(t *testing.T) {
	v := vip3ToFeeMap("50,75,90,100", 60)
	fmt.Println(v)
	v = vip3ToFeeMap("", 60)
	fmt.Println(v)
}
