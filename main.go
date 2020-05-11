package main

import (
	"archive/zip"
	"context"
	"flag"
	"fmt"
	"io"
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

// Root returns the root inode
func (f *FS) Root() (fs.Node, error) {
	n := &Dir{
		archive: f.archive,
	}
	return n, nil
}

// Dir represents a directory
type Dir struct {
	archive *zip.Reader
	file    *zip.File
}

// Attr return the file attributes
func (d *Dir) Attr() fuse.Attr {
	if d.file == nil {
		return fuse.Attr{Mode: os.ModeDir | 0755}
	}
	return zipAttr(d.file)
}

// Lookup a directory given a ctx
func (d *Dir) Lookup(ctx context.Context, req *fuse.LookupRequest, resp *fuse.LookupResponse) (fs.Node, error) {
	path := req.Name
	if d.file != nil {
		path = d.file.Name + path
	}

	for _, f := range d.archive.File {
		switch {
		case f.Name == path:
			child := &File{file: f}
			return child, nil
		case f.Name[:len(f.Name)-1] == path && f.Name[len(f.Name)-1] == '/':
			child := &Dir{
				archive: d.archive,
				file:    f,
			}
			return child, nil
		}
	}
	return nil, fuse.ENOENT
}

// File that represents a file
type File struct {
	file *zip.File
}

var _ fs.Node = (*File)(nil)

// Attr returns the attributes for file
func (f *File) Attr() fuse.Attr {
	return zipAttr(f.file)
}

var _ = fs.NodeOpener(&File{})

// Open a file and return a handle
func (f *File) Open(req *fuse.OpenRequest, resp *fuse.OpenResponse, intr fs.Intr) (fs.Handle, fuse.Error) {
	r, err := f.file.Open()
	if err != nil {
		return nil, err
	}
	resp.Flags |= fuse.OpenNonSeekable
	return &FileHandle{r: r}, nil
}

// FileHandle for pointers to files
type FileHandle struct {
	r io.ReadCloser
}

var _ fs.Handle = (*FileHandle)(nil)

func zipAttr(f *zip.File) fuse.Attr {
	return fuse.Attr{
		Size:   f.UncompressedSize64,
		Mode:   f.Mode(),
		Mtime:  f.ModTime(),
		Ctime:  f.ModTime(),
		CrTime: f.ModTime(),
	}
}

var _ fs.FS = (*FS)(nil)
var progName = filepath.Base(os.Args[0])
var _ = fs.NodeRequestLookuper(&Dir{})

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
