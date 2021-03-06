package fenv

import (
	"encoding/json"
	"flag"
	"fmt"
	"reflect"
	"testing"
)

func TestParseError(t *testing.T) {
	fs := flag.NewFlagSet("test1", flag.ContinueOnError)
	es := NewEnvSet(fs)
	var v, v2 int
	fs.IntVar(&v, "abc123", 0, "")
	fs.IntVar(&v2, "abc124", 0, "")
	es.Var(&v, "", "t", "foo")
	es.Var(&v2, "", "d")
	err := es.ParseEnv(envmap{
		"T": "NOTINT",
		"D": "NOTINT",
	})
	if err == nil {
		t.Fatal("should have failed")
	}
	_ = err.Error() // test coverage of error string
	fe := err.(FlagError)
	data, err := json.MarshalIndent(fe, "", " ")
	if err != nil {
		t.Fatal(err)
	}
	const expect = `{
 "Flag": {
  "Name": "abc123",
  "Usage": "",
  "Value": 0,
  "DefValue": "0"
 },
 "Name": "T",
 "Value": "NOTINT",
 "Names": [
  "ABC123",
  "T",
  "FOO"
 ],
 "Env": {
  "T": "NOTINT"
 },
 "IsSet": false,
 "IsSelfSet": false,
 "Err": {
  "Func": "ParseInt",
  "Num": "NOTINT",
  "Err": {}
 }
}`

	if string(data) != expect {
		t.Fatalf("\ngot:\n%s\nexpect:\n%s", string(data), expect)
	}
	es.VisitAll(func(e EnvFlag) {
		switch e.Name {
		case "T":
			if e.Err == nil {
				t.Fatal(e.Name)
			}
		case "D":
			if e.Err != nil {
				t.Fatal(e.Name)
			}
		default:
			t.Fatal(e.Name)
		}
	})
}

func TestContinueOnError(t *testing.T) {
	fs := flag.NewFlagSet("test1", flag.ContinueOnError)
	es := NewEnvSet(fs, ContinueOnError())
	var v, v2 int
	fs.IntVar(&v, "abc123", 0, "")
	fs.IntVar(&v2, "abc124", 0, "")
	es.Var(&v, "", "t", "foo")
	es.Var(&v2, "", "d")
	err := es.ParseEnv(envmap{
		"T": "NOTINT",
		"D": "NOTINT",
	})
	if err == nil {
		t.Fatal("should have failed")
	}
	if err != ErrMultipleSet {
		t.Fatal(err)
	}
	es.VisitAll(func(e EnvFlag) {
		if e.Err == nil {
			t.Fatal(e.Name)
		}
	})
}

func TestAlredyParsed(t *testing.T) {
	fs := flag.NewFlagSet("test1", flag.ContinueOnError)
	es := NewEnvSet(fs, Prefix("pre", "fixes"))
	var v1, v2 string
	fs.StringVar(&v1, "t1", "def", "")
	fs.StringVar(&v2, "t2", "def", "")
	es.Var(&v1)
	es.Var(&v2)
	if es.Parsed() {
		t.Fatal()
	}
	if err := fs.Parse([]string{"-t1", "v1"}); err != nil {
		t.Fatal(err)
	}
	if err := es.ParseEnv(envmap{
		"T1": "wrong",
		"T2": "v2",
	}); err != nil {
		t.Fatal(err)
	}
	if v1 != "v1" {
		t.Fatal(v1)
	}
	if v2 != "def" {
		t.Fatal(v2)
	}
	if !es.Parsed() {
		t.Fatal()
	}
	if es.Parse() != ErrAlreadyParsed {
		t.Fatal(es.Parse())
	}
}

func TestPrefix(t *testing.T) {
	fs := flag.NewFlagSet("test1", flag.ContinueOnError)
	es := NewEnvSet(fs, Prefix("pre", "fixes"))
	var v string
	fs.StringVar(&v, "testtest", "", "")
	if err := es.ParseEnv(envmap{
		"PRE_FIXESTESTTEST": "BOO",
	}); err != nil {
		t.Fatal(err)
	}

	if v != "BOO" {
		t.Fatal()
	}
}

func TestFlagByName(t *testing.T) {
	fs := flag.NewFlagSet("test1", flag.ContinueOnError)
	es := NewEnvSet(fs)
	var v string
	fs.StringVar(&v, "testtest", "", "")
	assertPanic(t, func() { // not a registered flag
		es.Flag("notaflag", "boo")
	})
	assertPanic(t, func() { // not a registered flag
		es.Flag("TESTTEST", "boo")
	})
	es.Flag("testtest", "boo")
}

func TestEnvSet(t *testing.T) {
	fs := flag.NewFlagSet("test1", flag.ContinueOnError)
	es := NewEnvSet(fs)
	var v, v2, v3 string
	fs.StringVar(&v, "test-1", "", "test variable")
	fs.StringVar(&v2, "test.2", "", "test2")
	fs.StringVar(&v3, "test3", "default", "test3")
	es.Var(&v, "TEST4", "", "TEST")
	es.Var(&v2, "TEST", "", "TEST")
	es.Var(&v3)
	assertPanic(t, func() { // not a registered flag
		var v string
		es.Var(&v, "", "TESTAGAIN")
	})
	assertPanic(t, func() { // not a pointer to value
		var v string
		es.Var(v, "", "TESTAGAIN")
	})
	assertPanic(t, func() { // not a registered flag
		var v string
		es.Var(&v)
	})
	assertPanic(t, func() { // not a pointer to value
		var v string
		es.Var(v)
	})
	if err := es.ParseEnv(envmap{
		"TEST_1": "BOO",
		"TEST_2": "FOO",
		"TEST3":  "FOO",
		"TEST4":  "V4",
	}); err != nil {
		t.Fatal(err)
	}
	expected := map[string]EnvFlag{
		"test-1": EnvFlag{
			Name:  "TEST4",
			Names: []string{"TEST4", "TEST_1", "TEST"},
			Env: map[string]string{
				"TEST4":  "V4",
				"TEST_1": "BOO",
			},
			IsSet:     true,
			IsSelfSet: true,
			Value:     "V4",
		},
		"test.2": EnvFlag{
			Name:  "TEST_2",
			Names: []string{"TEST", "TEST_2", "TEST"},
			Env: map[string]string{
				"TEST_2": "FOO",
			},
			IsSet:     true,
			IsSelfSet: true,
			Value:     "FOO",
		},
	}
	if err := fs.Parse([]string{}); err != nil {
		t.Fatal(err)
	}
	if v != "V4" {
		t.Fatal(v)
	}
	if v2 != "FOO" {
		t.Fatal(v2)
	}
	if v3 != "default" {
		t.Fatal(v3)
	}
	es.VisitAll(func(e EnvFlag) {
		exp := expected[e.Flag.Name]
		exp.Flag = e.Flag
		if !reflect.DeepEqual(exp, e) {

			t.Fatalf("\nexp:%v\ngot:%v", jsonstr(exp), jsonstr(e))
		}
	})
}

func assertPanic(t *testing.T, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	f()
}

// for debugging: es.VisitAll(visitPrint)
func visitPrint(e EnvFlag) {
	data, err := json.MarshalIndent(&e, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
}

// for debugging: es.VisitAll(visitPrint)
func jsonstr(e EnvFlag) string {
	data, err := json.MarshalIndent(&e, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(data)
}
