package fenv

import (
	"errors"
	"flag"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// NewEnvSet returns a *EnvSet
func NewEnvSet(fs *flag.FlagSet, opt ...Option) *EnvSet {
	es := &EnvSet{
		fs:      fs,
		names:   make(map[string][]string),
		exclude: make(map[string]bool),
		applied: make(map[string]string),
		env:     make(map[string]string),
		errs:    make(map[string]error),
	}
	for _, o := range opt {
		o(es)
	}
	return es
}

// EnvFlag is used bu the EnvSet.Visit* funcs.
type EnvFlag struct {
	// the associated flag.Flag
	Flag *flag.Flag
	// the environment variable name which the value for flag parsing was
	// extracted from. This field is set regardless of if the flag.Set
	// succeeds or fails.
	Name string
	// the value of the Name environment variable
	Value string
	// all the env variable names mapped associated with the flag.
	AllNames []string
	// true when the flag has been sucessfully set by the envsdert
	IsSet bool
	// error caused by flag.Set
	Err error
}

// ErrAlreadyParsed is returned by EnvSet.Parse() if the EnvSet already was parsed.
var ErrAlreadyParsed = errors.New("the envset is already parsed")

// ErrMultipleSet is returned by EnvSet.Parse() if the ContinueOnError is enabled and more than one flag failed to be set.
var ErrMultipleSet = errors.New("multiple errors encountered when calling flag.Set()")

// FlagError
type FlagError struct {
	// the associated flag.Flag
	Flag *flag.Flag
	// the value which failed to parse
	Value string
	// the environment variable name which failed to parse
	Name string
	// all the env variable names mapped associated with the flag
	AllNames []string
	// the actual flag parse error
	Err error
}

func (f FlagError) Error() string {
	return fmt.Sprintf("failed to set flag %q with value %q", f.Flag.Name, f.Value)
}

type Option func(e *EnvSet)

func Prefix(prefix ...string) Option {
	return func(e *EnvSet) {
		e.prefix = strings.ToUpper(strings.Join(prefix, "_"))
	}
}

func ContinueOnError() Option {
	return func(e *EnvSet) {
		e.continueOnError = true
	}
}

// EnvSet adds environment variable support for flag.FlagSet.
type EnvSet struct {
	fs              *flag.FlagSet
	prefix          string
	continueOnError bool // parses all flags even if one fails

	// the key for all these maps are based on the flag.Flag.Name
	names   map[string][]string // all en var names
	exclude map[string]bool     // excluded from env vars
	applied map[string]string   // flags which value was set by a specific env var
	env     map[string]string   // the environment when parse was run
	errs    map[string]error    // errors which occured during flag.Set()
	parsed  bool                // true after Parse() has been run
}

// Var enables associattion with environment variable names other than the default auto generated ones
//
// If no name argument is supplied the variable will be excluded from
// environment pasrsing. The special name value emtpy string "" will be
// translated to the automatically generated environment variable name.
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

func (s *EnvSet) Flag(flagName string, names ...string) {
	var flg *flag.Flag
	s.fs.VisitAll(func(f *flag.Flag) {
		if flg != nil {
			return
		}
		if f.Name == flagName {
			flg = f
		}
	})
	if flg == nil {
		panic(fmt.Sprintf("%s is not a registed flag in the flagset", flagName))
	}
	s.Var(flg.Value, names...)
}

// Parsed reports whether s.Parse has been called.
func (s *EnvSet) Parsed() bool {
	return s.parsed
}

func (s *EnvSet) Parse() error {
	return s.ParseEnv(OSEnv())
}

func (s *EnvSet) ParseEnv(e map[string]string) error {
	if s.parsed {
		return ErrAlreadyParsed
	}
	s.parsed = true
	actual := make(map[string]bool)
	s.fs.Visit(func(f *flag.Flag) {
		actual[f.Name] = true
	})
	var (
		err  error
		nerr int
	)
	s.fs.VisitAll(func(f *flag.Flag) {
		if s.exclude[f.Name] {
			return
		}
		allNames := s.allNames(f)
		if actual[f.Name] {
			return // skip if already set
		}
	eachName:
		for _, name := range allNames {
			v := e[name]
			if v != "" {
				s.applied[f.Name] = name
				s.env[name] = v
				if s.continueOnError || err == nil {
					if ferr := f.Value.Set(v); ferr != nil {
						nerr++
						s.errs[f.Name] = ferr
						err = FlagError{
							Flag:     f,
							Value:    v,
							Name:     name,
							AllNames: allNames,
							Err:      ferr,
						}
					}
				}
				break eachName
			}
		}
	})
	if nerr > 1 {
		return ErrMultipleSet
	}
	return err
}

// Visit visits all non exluded EnvFlags in the flagset
func (s *EnvSet) VisitAll(fn func(e EnvFlag)) {
	actual := make(map[string]bool)
	s.fs.Visit(func(f *flag.Flag) {
		actual[f.Name] = true
	})
	s.fs.VisitAll(func(f *flag.Flag) {
		n := f.Name
		if s.exclude[n] {
			return
		}
		fn(EnvFlag{
			Flag:     f,
			Name:     s.applied[n],
			AllNames: s.allNames(f),
			IsSet:    actual[f.Name],
			Value:    s.env[s.applied[n]],
			Err:      s.errs[f.Name],
		})
	})
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

// allNames return all environment namesg for a given flag
func (s *EnvSet) allNames(f *flag.Flag) []string {
	var allNames []string
	if names, ok := s.names[f.Name]; ok {
		for _, name := range names {
			if name == "" {
				name = fmtEnv(f.Name, s.prefix)
			}
			allNames = append(allNames, name)
		}
	}
	if len(allNames) == 0 {
		allNames = append(allNames, fmtEnv(f.Name, s.prefix))
	}
	return allNames
}

// fmtEnv formats a environment variable name as expected by this package.
func fmtEnv(s string, prefix ...string) string {
	s = strings.Join(prefix, "_") + s
	s = strings.Replace(s, ".", "_", -1)
	s = strings.Replace(s, "-", "_", -1)
	s = strings.ToUpper(s)
	return s
}

func Var(v interface{}, names ...string) {
	commandLineMu.Lock()
	defer commandLineMu.Unlock()
	commandLine.Var(v, names...)
}

func Parse() error {
	commandLineMu.Lock()
	defer commandLineMu.Unlock()
	return commandLine.Parse()
}

func Parsed() bool {
	commandLineMu.Lock()
	defer commandLineMu.Unlock()
	return commandLine.Parsed()
}

func VisitAll(fn func(e EnvFlag)) {
	commandLineMu.Lock()
	defer commandLineMu.Unlock()
	commandLine.VisitAll(fn)
}

var (
	commandLineMu sync.Mutex
	commandLine   = NewEnvSet(flag.CommandLine)
)

// CommandLinePrefix sets the prefix used by the package level env set functions.
func CommandLinePrefix(prefix ...string) {
	commandLineMu.Lock()
	defer commandLineMu.Unlock()
	if commandLine.prefix != "" {
		panic("prefix already set: " + commandLine.prefix)
	}
	if Parsed() {
		panic("default commandline envset already parsed")
	}
	commandLine.prefix = strings.ToUpper(strings.Join(prefix, "_"))
}
