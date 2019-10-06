package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ameteiko/log-parser/internal/processor"
	"github.com/ameteiko/log-parser/internal/stats"
)

const (
	inFile  = "file"
	inStdin = "stdin"
)

var (
	errInputTypeInvalid = fmt.Errorf("input type is invalid. Use either %q or %q", inFile, inStdin)
	errInputFileError   = errors.New("cannot open input file for reading")
)

// config aggregates all application configuration settings.
type config struct {
	input *bufio.Scanner
}

func main() {
	cfg, err := parseConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	stat := stats.NewStats()
	p := processor.NewProcessor(cfg.input, stat)
	p.Process()

	println("Done!")
}

// parseConfig parses application configuration value come from the CLI flag parameters, validates these values and
// returns an application config object.
func parseConfig() (*config, error) {
	var (
		inputType string
		input     *bufio.Scanner
	)
	flag.StringVar(&inputType, "in", inStdin, "input type. file for a file name, stdin for standard input.")
	flag.Parse()

	// Parameters normalization.
	inputType = strings.ToLower(strings.ToLower(inputType))
	if inputType != inStdin && inputType != inFile {
		return nil, errInputTypeInvalid
	}

	// Open either stdin or file for reading.
	if inputType == inStdin {
		input = bufio.NewScanner(os.Stdin)
	} else {
		// The first unnamed CLI parameter is the input file name.
		f, err := os.Open(flag.Arg(0))
		if err != nil {
			return nil, errInputFileError
		}
		input = bufio.NewScanner(f)
	}

	return &config{
		input: input,
	}, nil
}
