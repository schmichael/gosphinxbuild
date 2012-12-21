package main

import (
	"path/filepath"
	"flag"
	"github.com/schmichael/gosphinxbuild"
	"log"
	"os"
)

var path = flag.String("path", ".", "path containing a sphinx Makefile")
var cmd = flag.String("cmd", "make html", "command to run when files change")

func main() {
	flag.Parse()

	// Only sanity check *path if it's not cwd
	if *path != "." {
		fi, err := os.Stat(*path)
		if err != nil {
			log.Fatalf("Could not stat %v: %v\n", *path, err)
		}
		if !fi.IsDir() {
			log.Fatalf("Path must be a directory. %s is not.\n", *path)
		}
	}

	ap, err := filepath.Abs(*path)
	if err != nil {
		log.Fatalf("Could not resolve path %s: %v\n", *path, err)
	}

	gosphinxbuild.Watch(ap, *cmd)
}
