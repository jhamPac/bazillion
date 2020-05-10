package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

var progName = filepath.Base(os.Args[0])

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", progName)
	fmt.Fprintf(os.Stderr, "%s ZIP MOUNTPOINT\n", progName)
	flag.PrintDefaults()
}

func main() {

}
