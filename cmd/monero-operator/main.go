package main

import (
	"os"

	"github.com/jessevdk/go-flags"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func init() {
	log.SetLogger(zap.New(zap.UseDevMode(true)))
}

type GlobalOptions struct {
	Verbose bool `long:"verbose" short:"v" description:"enable verbose logs output"`
}

var (
	globalOptions GlobalOptions
	parser        = flags.NewParser(&globalOptions, flags.Default)
)

func main() {
	if _, err := parser.Parse(); err != nil {
		switch flagsErr := err.(type) {
		case flags.ErrorType:
			if flagsErr == flags.ErrHelp {
				os.Exit(0)
			}
			os.Exit(1)
		default:
			os.Exit(1)
		}
	}
}
