package fenv

import (
	"encoding/json"
	"flag"
	"testing"
)

func TestCommandline(t *testing.T) {
	var v1, v2 string
	flag.StringVar(&v1, "tetete1", "", "")
	flag.StringVar(&v2, "tetete2", "", "")
	Var(&v1, "hello")
	Var(&v2) // excludes v2 from envset
	if err := Parse(); err != nil {
		t.Fatal(err)
	}
}

func TestParseError(t *testing.T) {
	fs := flag.NewFlagSet("test1", flag.ContinueOnError)
	es := NewEnvSet(fs)
	var v int
	fs.IntVar(&v, "abc123", 0, "")
	es.Var(&v, "_", "t", "foo")
	err := es.ParseEnv(Env{
		"T": "NOTINT",
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
 "Value": "NOTINT",
 "Name": "T",
 "AllNames": [
  "ABC123",
  "T",
  "FOO"
 ],
 "Err": {
  "Func": "ParseInt",
  "Num": "NOTINT",
  "Err": {}
 }
}`

	if string(data) != expect {
		t.Fatalf("\ngot:\n%s\nexpect:\n%s", string(data), expect)
	}
}

func TestAlredyParsed(t *testing.T) {
	fs := flag.NewFlagSet("test1", flag.ContinueOnError)
	es := NewEnvSet(fs, "pre", "fixes")
	var v1, v2 string
	fs.StringVar(&v1, "t1", "def", "")
	fs.StringVar(&v2, "t2", "def", "")
	es.Var(&v1)
	es.Var(&v2)
	if err := fs.Parse([]string{"-t1", "v1"}); err != nil {
		t.Fatal(err)
	}
	if err := es.ParseEnv(Env{
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
}

func TestPrefix(t *testing.T) {
	fs := flag.NewFlagSet("test1", flag.ContinueOnError)
	es := NewEnvSet(fs, "pre", "fixes")
	var v string
	fs.StringVar(&v, "testtest", "", "")
	es.Var(&v)
	if err := es.ParseEnv(Env{
		"PRE_FIXES_TEST1": "BOO",
	}); err != nil {
		t.Fatal(err)
	}
}

func TestEnvSet(t *testing.T) {
	fs := flag.NewFlagSet("test1", flag.ContinueOnError)
	es := NewEnvSet(fs)
	var v, v2, v3 string
	fs.StringVar(&v, "test-1", "", "test variable")
	fs.StringVar(&v2, "test.2", "", "test2")
	fs.StringVar(&v3, "test3", "default", "test3")
	es.Var(&v, "TEST", "_", "TEST")
	es.Var(&v2, "TEST", "_", "TEST")
	es.Var(&v3)
	assertPanic(t, func() { // not a registered flag
		var v string
		es.Var(&v, "_", "TESTAGAIN")
	})
	assertPanic(t, func() { // not a pointer to value
		var v string
		es.Var(v, "_", "TESTAGAIN")
	})
	assertPanic(t, func() { // not a registered flag
		var v string
		es.Var(&v)
	})
	assertPanic(t, func() { // not a pointer to value
		var v string
		es.Var(v)
	})
	if err := es.ParseEnv(Env{
		"TEST_1": "BOO",
		"TEST_2": "FOO",
		"TEST3":  "FOO",
	}); err != nil {
		t.Fatal(err)
	}
	if err := fs.Parse([]string{}); err != nil {
		t.Fatal(err)
	}
	if v != "BOO" {
		t.Fatal(v)
	}
	if v2 != "FOO" {
		t.Fatal(v2)
	}
	if v3 != "default" {
		t.Fatal(v3)
	}
}

func assertPanic(t *testing.T, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	f()
}
