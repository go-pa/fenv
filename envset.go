package fenv

import (
	"flag"
	"fmt"
	"reflect"
	"strings"
)

// NewEnvSet returns a *EnvSet tied to the
func NewEnvSet(fs *flag.FlagSet, prefix ...string) *EnvSet {
	return &EnvSet{
		fs:      fs,
		prefix:  strings.ToUpper(strings.Join(prefix, "_")),
		names:   make(map[string][]string),
		exclude: make(map[string]bool),
	}
}

// FlagError
type FlagError struct {
	Flag     *flag.Flag // the associated flag.Flag
	Value    string     // the value which failed to parse
	Name     string     // the environment variable name which failed to parse
	AllNames []string   // all the env variable names mapped associated with the flag
	Err      error      // the actual flag parse error
}

func (f FlagError) Error() string {
	return fmt.Sprintf("failed to set flag %q with value %q", f.Flag.Name, f.Value)
}

// EnvSet adds environment variable support for flag.FlagSet.
type EnvSet struct {
	fs      *flag.FlagSet
	prefix  string
	names   map[string][]string
	exclude map[string]bool
}

// Var enables associattion with environment variable names other than the default auto generated ones
//
// If no name argument is supplied the variable will be excluded from
// environment pasrsing. The special name value "_" will be translated to the
// automatically generated environment variable name.
func (s *EnvSet) Var(v interface{}, names ...string) {
	f, err := s.findFlag(v)
	if err != nil {
		panic(err)
	}
	if f == nil {
		panic(fmt.Sprintf("%T (%v) is not a member of the flagset", v, v))
	}
	if len(names) == 0 {
		s.exclude[f.Name] = true
		delete(s.names, f.Name)
		return
	}

	for i, v := range names {
		names[i] = fmtEnv(v)
	}
	s.names[f.Name] = names
	delete(s.exclude, f.Name)
}

func (s *EnvSet) Parse() error {
	return s.ParseEnv(OSEnv())
}

func (s *EnvSet) ParseEnv(e Env) error {
	actual := make(map[string]bool)
	s.fs.Visit(func(f *flag.Flag) {
		actual[f.Name] = true
	})
	var err error
	s.fs.VisitAll(func(f *flag.Flag) {
		if err != nil {
			return
		}
		if actual[f.Name] || s.exclude[f.Name] {
			return // skip if already set or excluded
		}
		var allNames []string
		if names, ok := s.names[f.Name]; ok {
			for _, name := range names {
				if name == "_" {
					name = fmtEnv(f.Name, s.prefix)
				}
				allNames = append(allNames, name)
			}
		}
		if len(allNames) == 0 {
			allNames = append(allNames, fmtEnv(f.Name, s.prefix))
		}
		for _, name := range allNames {
			v := e[name]
			if v != "" {
				if ferr := f.Value.Set(v); ferr != nil {
					err = FlagError{
						Flag:     f,
						Value:    v,
						Name:     name,
						AllNames: allNames,
						Err:      ferr,
					}
				}
			}
		}
	})
	return err
}

// returns the flag.Flag instace bound to ref or nil if not found
func (s *EnvSet) findFlag(v interface{}) (*flag.Flag, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("not a pointer: %v", v)
	}
	vp := rv.Pointer()
	var flg *flag.Flag
	s.fs.VisitAll(func(f *flag.Flag) {
		p := reflect.ValueOf(f.Value).Pointer()
		if vp == p {
			// todo: find out which PLUGIN_ or other env var we are refering to
			// so that a better message can be printed
			flg = f
		}
	})
	return flg, nil
}

// fmtEnv formats a environment variable name as expected by this package.
func fmtEnv(s string, prefix ...string) string {
	for _, v := range prefix {
		if v != "" {
			s = v + "_" + s
		}
	}
	s = strings.Replace(s, ".", "_", -1)
	s = strings.Replace(s, "-", "_", -1)
	s = strings.ToUpper(s)
	return s
}

func Var(v interface{}, names ...string) {
	commandLine.Var(v, names...)

}

func Parse() error {
	return commandLine.Parse()
}

var commandLine = NewEnvSet(flag.CommandLine, "")
