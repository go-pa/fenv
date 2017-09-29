package fenv

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// Env contains environment variables as key/value map instead of a slice
// like the standard library does it.
type Env map[string]string

// OSEnv returns os.Environ() as a Env.
func OSEnv() Env {
	oe := os.Environ()
	e := make(Env, len(oe))
	err := e.parse(oe)
	if err != nil {
		panic(err)
	}
	return e

}

// Parse parses and sets environment variables in the KEY=VALUE string format
// used by the standard library.
func (e Env) parse(oe []string) error {
	for _, v := range oe {
		kv := strings.SplitN(v, "=", 2)
		if len(kv) < 2 {
			return fmt.Errorf("expected format key=value in '%s'", v)
		}
		e[kv[0]] = kv[1]
	}
	return nil
}

// Slice returns the contents of the env in the format of the environment used
// by the standard library for use with os/exec and similar packages.
func (e Env) slice() []string {
	var res []string
	for k, v := range e {
		res = append(res, k+"="+v)
	}
	sort.Strings(res)
	return res
}

// Update updates e with all entries from o.
func (e Env) update(o Env) {
	for k, v := range o {
		e[k] = v
	}
}

// Set sets all variables in the process environment.
func (e Env) set() error {
	for k, v := range e {
		err := os.Setenv(k, v)
		if err != nil {
			return err
		}
	}
	return nil
}
