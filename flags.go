package box

import (
	"errors"
	"flag"
)

var (
	logLevel      string
	listenAddress string
	tlsCertFile   string
	tlsKeyFile    string
	cpuMinThreads int
	memLimitRatio float64
)

// MustRegisterFlags registers all Config fields as flags with the flag package and calls flag.Parse.
// Panics if a flag has already been registered.
// The registered flags are:
//
//	-log-level
//	-listen-address
//	-tls-cert-file
//	-tls-key-file
//	-cpu-min-threads
//	-mem-limit-ratio
func MustRegisterFlags() {
	flag.StringVar(&logLevel, "log-level", "", "Log level")
	flag.StringVar(&listenAddress, "listen-address", "", "Webserver listen address")
	flag.StringVar(&tlsCertFile, "tls-cert-file", "", "Webserver TLS certificate file")
	flag.StringVar(&tlsKeyFile, "tls-key-file", "", "Webserver TLS key file")
	flag.IntVar(&cpuMinThreads, "cpu-min-threads", 1, "CPU minimum threads")
	flag.Float64Var(&memLimitRatio, "mem-limit-ratio", 0.8, "Memory limit ratio")
}

// ErrFlagsAlreadyParsed indicates that the flag.Parse function has already been called.
var ErrFlagsAlreadyParsed = errors.New("flag.Parse() has already been called")

// MustRegisterAndParseFlags calls MustRegisterFlags & flag.Parse.
// Panics with ErrFlagsAlreadyParsed if flag.Parse has already been called.
func MustRegisterAndParseFlags() {
	MustRegisterFlags()

	if flag.Parsed() {
		panic(ErrFlagsAlreadyParsed)
	}

	flag.Parse()
}

func setupBoxWithFlags(box *Box) {
	if len(logLevel) > 0 {
		box.Config.LogLevel = logLevel
	}

	if len(listenAddress) > 0 {
		box.Config.ListenAddress = listenAddress
	}

	if len(tlsCertFile) > 0 {
		box.Config.TLSCertFile = tlsCertFile
	}

	if len(tlsKeyFile) > 0 {
		box.Config.TLSKeyFile = tlsKeyFile
	}

	if cpuMinThreads > 1 {
		box.onceCpu.Do(func() {
			box.Config.CpuMinThreads = cpuMinThreads
		})
	}

	if memLimitRatio != 0.8 {
		box.onceMem.Do(func() {
			box.Config.MemLimitRatio = memLimitRatio
		})
	}
}
