package argstruct

import (
	"encoding/csv"
	"fmt"
	"log"
	"log/slog"
	"os"
	"reflect"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"

	"golang.org/x/exp/maps"
)

// a very simple and opinionated argument parser based around structs and struct tags
// partially compete
// TODO: 	environment variable handling
// 			child of options

const (
	Version = "0.1.3"
)

type ArgStructable interface {
	Run(*ArgStruct) error
}

type HasVersion interface {
	Version() string
}

func Run(a ArgStructable) {
	x := &ArgStruct{
		as:     a,
		tas:    reflect.TypeOf(a),
		args:   make(map[string]*argConfig),
		pos:    make(map[int]*argConfig),
		groups: make(map[string][]*argConfig),
	}
	if err := x.Run(); err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
	os.Exit(0)
}

type argConfig struct {
	ignore        bool
	required      bool
	group         string
	defaultV      string
	help          string
	flag          string
	andShort      bool
	noEnv         bool
	posN          int
	assignedFlags []string

	f      func()
	sField reflect.Value
	name   string
	tag    string
	set    bool
}

type ArgStruct struct {
	AppName    string
	AppVersion string
	AppModule  string
	as         ArgStructable
	tas        reflect.Type
	args       map[string]*argConfig
	pos        map[int]*argConfig
	groups     map[string][]*argConfig
}

func (x *ArgStruct) Dump() {
	for k, v := range x.args {
		fmt.Printf("%s: %+v\n", k, v)
	}
}

func (x *ArgStruct) Run() error {
	x.ParseStruct()
	// x.Dump()
	if err := x.ParseArgv(); err != nil {
		return err
	}
	return x.as.Run(x)
}

func (x *ArgStruct) ParseArgv() error {
	return x.ParseArgs(os.Args)
}

type ArgFeed struct {
	args []string
	pos  int
	now  string
}

func (x *ArgFeed) Next() (string, error) {
	if x.pos >= len(x.args) {
		return "", fmt.Errorf("end of args")
	}
	x.now = x.args[x.pos]
	x.pos++
	return x.now, nil
}

func (x *ArgStruct) ParseArgs(args []string) error {
	// we don't need argv[0], it's just us
	aFeed := &ArgFeed{args: args[1:]}
	// pos starts at 1
	nPos := 1

	for {
		arg, err := aFeed.Next()
		if err != nil {
			// the end?
			break
		}

		if strings.HasPrefix(arg, "-") {
			if x.args[arg] == nil {
				return fmt.Errorf("unknown argument: %s", arg)
			}

			// functions come before setting
			if x.args[arg].f != nil {
				x.args[arg].f()
				continue
			}

			// no function, set the value
			if err := x.SetArg(x.args[arg], aFeed); err != nil {
				return fmt.Errorf("error setting argument %s: %s", arg, err)
			}
			x.args[arg].set = true
			continue
		}

		if x.pos[nPos] != nil {
			if err := x.SetArg(x.pos[nPos], &ArgFeed{args: []string{aFeed.now}}); err != nil {
				return fmt.Errorf("error setting argument: %s", err)
			}
			x.pos[nPos].set = true
			nPos++
			continue
		}

		if x.args[arg] == nil {
			return fmt.Errorf("unknown argument: %s", arg)
		}
	}

	for _, ac := range x.args {
		if ac.required && !ac.set {
			return fmt.Errorf("required argument not set: %s", strings.Join(ac.assignedFlags, " or "))
		}
	}

	for _, ac := range x.groups {
		seen := false
		for _, ac2 := range ac {
			if ac2.set {
				seen = true
				break
			}
		}
		if !seen {
			tags := []string{}
			for _, ac2 := range ac {
				tags = append(tags, strings.Join(ac2.assignedFlags, "|"))
			}
			return fmt.Errorf("at least one option from %s must be set: %s", ac[0].group, strings.Join(tags, " or "))
		}
	}

	return nil
}

func (x *ArgStruct) SetArg(ac *argConfig, a *ArgFeed) error {
	// short circuirt bool, as all others require the next arg
	switch ac.sField.Kind() {
	case reflect.Bool:
		ac.sField.SetBool(true)
		return nil
	}

	n, err := a.Next()
	if err != nil {
		return fmt.Errorf("missing argument for %s", ac.name)
	}

	switch ac.sField.Kind() {
	case reflect.String:
		ac.sField.SetString(n)
	case reflect.Int:
		i, err := strconv.ParseInt(n, 10, 64)
		if err != nil {
			return fmt.Errorf("unable to parse int: %s", err)
		}
		ac.sField.SetInt(i)
	case reflect.Slice:
		r := csv.NewReader(strings.NewReader(n))
		records, err := r.ReadAll()
		if err != nil {
			return fmt.Errorf("unable to parse array: %s", err)
		}
		for _, rec := range records {
			for _, rec2 := range rec {
				switch ac.sField.Type().Elem().Kind() {
				case reflect.String:
					ac.sField.Set(reflect.Append(ac.sField, reflect.ValueOf(rec2)))
				case reflect.Int:
					i, err := strconv.Atoi(rec2)
					if err != nil {
						return fmt.Errorf("unable to parse int: %s", err)
					}
					ac.sField.Set(reflect.Append(ac.sField, reflect.ValueOf(i)))
				}
			}
		}
	}
	return nil
}

func (x *ArgStruct) fillArg(ac *argConfig) {
	for _, te := range strings.Split(ac.tag, ",") {
		kv := strings.Split(te, "=")
		switch strings.ToLower(kv[0]) {
		case "-":
			ac.ignore = true
			return
		case "required":
			ac.required = true
		case "group":
			ac.group = kv[1]
		case "default":
			ac.defaultV = kv[1]
			if err := x.SetArg(ac, &ArgFeed{args: []string{kv[1]}}); err != nil {
				log.Fatalf("error setting default: %s", err)
			}
		case "help":
			ac.help = kv[1]
		case "flag":
			ac.flag = kv[1]
		case "andshort":
			ac.andShort = true
		case "env":
			ac.noEnv = true // TODO
		case "pos":
			i, _ := strconv.Atoi(kv[1])
			ac.posN = i
		}
	}

	if len(ac.name) == 1 {
		x.args["-"+ac.name] = ac
		ac.assignedFlags = append(ac.assignedFlags, "-"+ac.name)
	} else {
		x.args["--"+ac.name] = ac
		ac.assignedFlags = append(ac.assignedFlags, "--"+ac.name)
		if ac.andShort {
			f := "-" + ac.name[:1]
			x.args[f] = ac
			ac.assignedFlags = append(ac.assignedFlags, f)
		}
	}

	if ac.group != "" {
		x.groups[ac.group] = append(x.groups[ac.group], ac)
	}

	if ac.posN != 0 {
		x.pos[ac.posN] = ac
	}
}

func (x *ArgStruct) ParseStruct() {
	// deref through the pointer
	as := reflect.ValueOf(x.as).Elem()

	// fill app name, version and module
	x.AppName = strings.ToLower(as.Type().Name())
	if _, ok := x.as.(HasVersion); ok {
		x.AppVersion = x.as.(HasVersion).Version()
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		x.AppModule = info.Path
	}

	for i := 0; i < as.NumField(); i++ {
		field := as.Field(i)

		// skip unexported fields
		if !field.CanSet() {
			continue
		}

		f := as.Type().Field(i)
		ac := &argConfig{
			sField: field,
			name:   strings.ToLower(f.Name),
			tag:    f.Tag.Get("argstruct"),
		}

		x.fillArg(ac)
	}

	// auto version
	if !slices.Contains(maps.Keys(x.args), "--version") {
		if _, ok := x.as.(HasVersion); ok {
			ac := &argConfig{
				name: "version",
				tag:  "help=shows version",
				f: func() {
					x.PrintVersion()
					os.Exit(0)
				},
				sField: reflect.ValueOf(false),
			}
			x.fillArg(ac)
		}
	}

	if !slices.Contains(maps.Keys(x.args), "--versionlong") {
		if _, ok := x.as.(HasVersion); ok {
			ac := &argConfig{
				name: "versionlong",
				tag:  "help=shows long version",
				f: func() {
					x.PrintVersionLong()
					os.Exit(0)
				},
				sField: reflect.ValueOf(false),
			}
			x.fillArg(ac)
		}
	}

	// auto help
	if !slices.Contains(maps.Keys(x.args), "-h") && !slices.Contains(maps.Keys(x.args), "--help") {
		ac := &argConfig{
			name:   "help",
			tag:    "help=show this help message,andShort",
			f:      x.PrintHelp,
			sField: reflect.ValueOf(false),
		}
		x.fillArg(ac)
	}

	// auto debug
	if !slices.Contains(maps.Keys(x.args), "--debug") {
		ac := &argConfig{
			name:   "debug",
			tag:    "help=enable debug logging",
			f:      x.EnableDebug,
			sField: reflect.ValueOf(false),
		}
		x.fillArg(ac)
	}

}

func (x *ArgStruct) PrintHelp() {
	fmt.Println(x.AppName, "usage:")

	seen := []string{}

	flags := maps.Keys(x.args)
	slices.Sort(flags)

	for _, f := range flags {
		ac := x.args[f]
		if ac.ignore {
			continue
		}
		slices.Sort(ac.assignedFlags)

		flagStr := strings.Join(ac.assignedFlags, ", ")
		if slices.Contains(seen, flagStr) {
			continue
		}
		seen = append(seen, flagStr)

		fmt.Printf("  %-15s", flagStr)
		if ac.flag != "" {
			fmt.Printf(", %s", ac.flag)
		}

		fmt.Printf(" %s", ac.name)
		if ac.help != "" {
			fmt.Printf(" - %s", ac.help)
		}

		fmt.Println("")
		if ac.required && ac.group != "" {
			fmt.Printf("  %-17s required: %v", "", ac.required)
		}
		if ac.group != "" {
			fmt.Printf("  %-17s required: true", "")
		}

		if ac.defaultV != "" {
			fmt.Printf(" %-17s default: %s", " ", ac.defaultV)
		}
		if ac.group != "" {
			fmt.Printf(" group: %s", ac.group)
		}
		// TODO ENV
		// if !ac.noEnv {
		// 	fmt.Printf(" var: %s_%s (TODO)", strings.ToUpper(x.appName), strings.ToUpper(ac.name))
		// }
		fmt.Println()
	}

	if len(x.groups) != 0 {
		fmt.Println()
		fmt.Println("groups: at least one must be set in this group")
		for _, g := range x.groups {
			fmt.Printf("  %-15s", g[0].group)
			for i, ac := range g {
				if i != 0 {
					fmt.Printf(" or ")
				}
				fmt.Printf("%s", strings.Join(ac.assignedFlags, "|"))
			}
			fmt.Println()
		}
		fmt.Println()
	}

	x.PrintVersionLong()
	os.Exit(0)
}

func (x *ArgStruct) EnableDebug() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
}

func (x *ArgStruct) PrintVersion() {
	fmt.Println(x.AppVersion)
}
func (x *ArgStruct) PrintVersionLong() {
	fmt.Println("app info:")
	fmt.Printf("  %-17s %s\n", "name", x.AppName)
	fmt.Printf("  %-17s %s\n", "version", x.AppVersion)
	fmt.Printf("  %-17s %s\n", "module", x.AppModule)
}
