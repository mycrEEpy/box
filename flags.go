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
)

// MustRegisterFlags registers all Config fields as flags with the flag package and calls flag.Parse.
// Panics if a flag has already been registered.
// The registered flags are:
//
//	-log-level
//	-listen-address
//	-tls-cert-file
//	-tls-key-file
func MustRegisterFlags() {
	flag.StringVar(&logLevel, "log-level", logLevel, "Log level")
	flag.StringVar(&listenAddress, "listen-address", listenAddress, "Webserver listen address")
	flag.StringVar(&tlsCertFile, "tls-cert-file", tlsCertFile, "Webserver TLS certificate file")
	flag.StringVar(&tlsKeyFile, "tls-key-file", tlsKeyFile, "Webserver TLS key file")
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
}
