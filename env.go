package fenv

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// env contains environment variables as key/value map instead of a slice
// like the standard library does it.
type env map[string]string

// OSEnv returns os.Environ() as a env.
func OSEnv() map[string]string {
	oe := os.Environ()
	e := make(env, len(oe))
	err := e.parse(oe)
	if err != nil {
		panic(err)
	}
	return e

}

// Parse parses and sets environment variables in the KEY=VALUE string format
// used by the standard library.
func (e env) parse(oe []string) error {
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
func (e env) slice() []string {
	var res []string
	for k, v := range e {
		res = append(res, k+"="+v)
	}
	sort.Strings(res)
	return res
}

// Update updates e with all entries from o.
func (e env) update(o env) {
	for k, v := range o {
		e[k] = v
	}
}

// Set sets all variables in the process environment.
func (e env) set() error {
	for k, v := range e {
		err := os.Setenv(k, v)
		if err != nil {
			return err
		}
	}
	return nil
}
