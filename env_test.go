package fenv

import (
	"encoding/json"
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	e := make(envmap)
	err := e.parse([]string{"A=3", "A=B", "C=D"})
	if err != nil {
		t.Fatal(err)
	}
	if e["A"] != "B" {
		t.Fatal()
	}
	if _, ok := e["B"]; ok {
		t.Fatal()
	}
	err = e.parse([]string{"B=3", "2=3", "wrong", "4=2"})
	if err.Error() != "expected format key=value in 'wrong'" {
		t.Fatal(err)
	}
}

func TestOSenv(t *testing.T) {
	if err := os.Setenv("EMPTY_VALUE", ""); err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv("TEST_test", "12345"); err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv("UNSET", "asd"); err != nil {
		t.Fatal(err)
	}
	if err := os.Unsetenv("UNSET"); err != nil {
		t.Fatal(err)
	}
	e := OSEnv()
	if e["TEST_test"] != "12345" {
		t.Fatal(e["TEST_test"])
	}
	if _, ok := e["UNSET"]; ok {
		t.Fatal()
	}
}

func TestSlice(t *testing.T) {
	e := envmap{
		"G": "H",
		"A": "B",
		"E": "F",
		"C": "D",
	}
	data, err := json.Marshal(e.slice())
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `["A=B","C=D","E=F","G=H"]` {
		t.Fatal(string(data))
	}
}

func TestUpdate(t *testing.T) {
	e1 := envmap{
		"A": "1",
		"B": "1",
	}
	e2 := envmap{
		"A": "2",
		"C": "2",
	}
	e1.update(e2)

	for k, v := range map[string]string{
		"A": "2",
		"B": "1",
		"C": "2",
	} {
		if e1[k] != v {
			t.Fatal(e1[k], k, v)
		}

	}
}

func TestSet(t *testing.T) {
	if err := (envmap{"BVOPEA": "!@V#@JDF"}.set()); err != nil {
		t.Fatal(err)
	}
	if os.Getenv("BVOPEA") != "!@V#@JDF" {
		t.Fatal(os.Getenv("BVOPEA"))
	}

}
