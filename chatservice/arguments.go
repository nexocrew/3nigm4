//
// 3nigm4 chatservice package
// Author: Federico Maggi <federicomaggi92@gmail.com>
// v1.0 23/08/2016
//

package main

// argType identify the available flag types, these types are
// described below.
type argType string

const (
	String      argType = "STRING"      // string flag type;
	StringSlice argType = "STRINGSLICE" // []string flag slice;
	Int         argType = "INT"         // int flag;
	Bool        argType = "BOOL"        // bool flag;
	Uint        argType = "UINT"        // uint flag;
	Duration    argType = "DURATION"    // time.Duration flag.
)

// cliArguments is used to define all available flags with name,
// shorthand, value, usage and kind.
type cliArguments struct {
	name      string
	shorthand string
	value     interface{}
	usage     string
	kind      argType
}

// setArgument invokes setArgumentPFlags before calling Viper config
// manager to integrate values.
func setArgument(command *cobra.Command, key string, destination interface{}) {
	setArgumentPFlags(command, key, destination)
	arg, _ := am[key]
	viper.BindPFlag(arg.name, command.PersistentFlags().Lookup(arg.name))
	viper.SetDefault(arg.name, arg.value)
}
