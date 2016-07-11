// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package channel

//Options specify configuration to start channel with
type Options struct {
	Host       string
	Port       int
	Debug      bool
	ConfigFile string
}

//DefaultOptions returns local development options for level channel
func DefaultOptions() *Options {
	return &Options{
		Host:       "0.0.0.0",
		Port:       3000,
		Debug:      true,
		ConfigFile: "",
	}
}
