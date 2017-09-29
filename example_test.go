package fenv_test

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-pa/fenv"
)

// Package level usage
func Example_package() {
	var s1, s2, s3 string

	flag.StringVar(&s1, "flag.1", "", "") // env var FLAG_1 is automatically added

	flag.StringVar(&s2, "flag2", "", "")
	fenv.Var(&s2, "alt-name", "_") // registers env names ALT_NAME and FLAG2.

	flag.StringVar(&s3, "flag3", "", "")
	fenv.Var(&s3) // excludes the var from being parsed as an enviroment variable.

	os.Setenv("FLAG_1", "v1")
	os.Setenv("ALT_NAME", "v2.alt")
	os.Setenv("FLAG_2", "v2")

	fmt.Println("before Parse()")
	fenv.VisitAll(func(e fenv.EnvFlag) {
		// don't print go test flags in example
		if !strings.HasPrefix(e.Flag.Name, "test.") {
			fmt.Printf("%s:%s\n", e.Flag.Name, e.Name)

		}
	})

	// call fenv.Parse() before flag.Parse()
	if err := fenv.Parse(); err != nil {
		panic(err)
	}
	flag.Parse()

	fmt.Println("after Parse()")
	fenv.VisitAll(func(e fenv.EnvFlag) {
		// don't print go test flags in example
		if !strings.HasPrefix(e.Flag.Name, "test.") {
			fmt.Printf("%s:%s\n", e.Flag.Name, e.Name)
		}
	})

	fmt.Println("values", s1, s2, flag.Parsed())
	// output:
	// before Parse()
	// flag.1:
	// flag2:
	// after Parse()
	// flag.1:FLAG_1
	// flag2:ALT_NAME
	// values v1 v2.alt true

}

// Use with fenv.EnvSet and flag.FlagSet
func Example_flagSet() {
	var s1, s2 string

	fs := flag.NewFlagSet("example", flag.ContinueOnError)
	es := fenv.NewEnvSet(fs, "my")

	fs.StringVar(&s1, "test1", "", "")
	fs.StringVar(&s2, "test2", "", "")
	es.Var(&s2, "other", "_")

	// es.Parse() or es.ParseEnv() to parse a custom environment
	if err := es.ParseEnv(map[string]string{
		"MY_TEST1": "v1",
		"MY_TEST2": "v2",
		"OTHER":    "v2.other",
	}); err != nil {
		log.Fatal(err)
	}
	if err := fs.Parse([]string{}); err != nil {
		log.Fatal(err)
	}

	fmt.Println(s1, s2)
	// output: v1 v2.other

}
