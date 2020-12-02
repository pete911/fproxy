package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type Flags struct {
	Port       int
	Verbose    bool
	OutputFile string
	TLSCrt     string
	TLSKey     string
}

func (f Flags) String() string {

	return fmt.Sprintf("verbose: %t output-file: %q tls-crt: %q tls-key: %q",
		f.Verbose, f.OutputFile, f.TLSCrt, f.TLSKey)
}

func ParseFlags() (Flags, error) {

	f := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	port := f.Int("port", getIntEnv("FP_PORT", 8080),
		"port on which to run proxy server")
	verbose := f.Bool("verbose", getBoolEnv("FP_VERBOSE", false),
		"verbose, ")
	outputFile := f.String("output-file", getStringEnv("FP_OUTPUT_FILE", ""),
		"file where to save requests and responses, if flag is not set, output is sent to std out")
	tlsCrt := f.String("tls-crt", getStringEnv("FP_TLS_CRT", ""),
		"path to tls certificate to be used by this service")
	tlsKey := f.String("tls-key", getStringEnv("FP_TLS_KEY", ""),
		"path to tls key to be used by this service")
	if err := f.Parse(os.Args[1:]); err != nil {
		return Flags{}, err
	}

	flags := Flags{
		Port:       intValue(port),
		Verbose:    boolValue(verbose),
		OutputFile: stringValue(outputFile),
		TLSCrt:     stringValue(tlsCrt),
		TLSKey:     stringValue(tlsKey),
	}

	err := flags.validate()
	return flags, err
}

func (f Flags) validate() error {

	if f.Port < 1 || f.Port > 65535 {
		return fmt.Errorf("invalid port %d", f.Port)
	}
	// TODO f.TLSKey and f.TLSCrt needs to be set or unset, not one or the other
	return nil
}

func getStringEnv(envName string, defaultValue string) string {

	env, ok := os.LookupEnv(envName)
	if !ok {
		return defaultValue
	}
	return env
}

func stringValue(v *string) string {

	if v == nil {
		return ""
	}
	return *v
}

func getIntEnv(envName string, defaultValue int) int {

	env, ok := os.LookupEnv(envName)
	if !ok {
		return defaultValue
	}

	if intValue, err := strconv.Atoi(env); err == nil {
		return intValue
	}
	return defaultValue
}

func intValue(v *int) int {

	if v == nil {
		return 0
	}
	return *v
}

func getBoolEnv(envName string, defaultValue bool) bool {

	env, ok := os.LookupEnv(envName)
	if !ok {
		return defaultValue
	}

	if intValue, err := strconv.ParseBool(env); err == nil {
		return intValue
	}
	return defaultValue
}

func boolValue(v *bool) bool {

	if v == nil {
		return false
	}
	return *v
}
