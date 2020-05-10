package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

// FS fuse struct that holds a zip reader
type FS struct {
	archive *zip.Reader
}

var _ fs.FS = (*FS)(nil)
var progName = filepath.Base(os.Args[0])

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", progName)
	fmt.Fprintf(os.Stderr, "%s ZIP MOUNTPOINT\n", progName)
	flag.PrintDefaults()
}

func main() {
	log.SetFlags(0)
	log.SetPrefix(progName + ": ")

	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 2 {
		usage()
		os.Exit(2)
	}

	path := flag.Arg(0)
	mountpoint := flag.Arg(1)
	if err := mount(path, mountpoint); err != nil {
		log.Fatal(err)
	}
}

func mount(path, mountpoint string) error {
	archive, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	defer archive.Close()

	c, err := fuse.Mount(mountpoint)
	if err != nil {
		return err
	}
	c.Close()

	filesys := &FS{
		archive: &archive.Reader,
	}
	if err := fs.Serve(c, filesys); err != nil {
		return err
	}

	<-c.Ready
	if err := c.MountError; err != nil {
		return err
	}

	return nil
}
