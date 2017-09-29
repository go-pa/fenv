package fenv_test

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/go-pa/fenv"
)

func ExamplePackage() {
	var s1, s2 string
	flag.StringVar(&s1, "flag.1", "", "") // env var FLAG_1 will be parsed
	flag.StringVar(&s2, "flag2", "", "")
	fenv.Var(&s2, "alt-name", "_")

	os.Setenv("FLAG_1", "v1")
	os.Setenv("ALT_NAME", "v2.alt")
	os.Setenv("FLAG_2", "v2")

	// call fenv.Parse() before flag.Parse()
	if err := fenv.Parse(); err != nil {
		panic(err)
	}
	flag.Parse()

	fmt.Println(s1, s2)
	// output: v1 v2.alt
}

func ExampleFlagSet() {
	var s1, s2 string

	fs := flag.NewFlagSet("example", flag.ContinueOnError)
	es := fenv.NewEnvSet(fs, "my")

	fs.StringVar(&s1, "test1", "", "")
	fs.StringVar(&s2, "test2", "", "")
	es.Var(&s2, "other", "_")

	// es.Parse() or es.ParseEnv() to parse a custom environment
	if err := es.ParseEnv(fenv.Env{
		"MY_TEST1": "v1",
		"MY_TEST2": "v2",
		"OTHER":    "v2.other",
	}); err != nil {
		log.Fatal(err)
	}
	if err := fs.Parse([]string{}); err != nil {
		log.Fatal(err)
	}

	// call fenv.Parse() before flag.Parse()
	if err := fenv.Parse(); err != nil {
		panic(err)
	}
	flag.Parse()

	fmt.Println(s1, s2)
	// output: v1 v2
}
