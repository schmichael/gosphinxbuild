// Package gosphinxbuild exports a function to run "make html" when files
// change on a path
package gosphinxbuild

import (
	"github.com/howeyc/fsnotify"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Call in its own goroutine to rebuild docs when buildChan is sent events
func builder(path string, cmd []string, buildChan chan bool) {
	for {
		// Block waiting for a new event
		<-buildChan

		// Pause briefly as editors often emit multiple events at once
		time.Sleep(100 * time.Millisecond)

		// Now just throw away the newest build change event
		select {
		case <-buildChan:
		default:
		}

		// And finally actually build the thing
		var c *exec.Cmd
		if len(cmd) == 1 {
			c = exec.Command(cmd[0])
		} else {
			c = exec.Command(cmd[0], cmd[1:]...)
		}
		out, err := c.CombinedOutput()
		if err != nil {
			log.Fatalf("Error running `%v`: %v\n", cmd, err)
		}
		log.Printf("%v\n%s", cmd, out)
	}
}

// Walk a path and watch non-hidden or build directories
func walkAndWatch(path string, w *fsnotify.Watcher) (watched uint) {
	f := func(path string, fi os.FileInfo, err error) error {
		if fi.IsDir() {
			// Skip hidden directories
			if fi.Name()[0] == '.' {
				return filepath.SkipDir
			}

			// Skip *build* directories
			if strings.Contains(fi.Name(), "build") {
					return filepath.SkipDir
			}

			// Skip *static* directories
			if strings.Contains(fi.Name(), "static") {
					return filepath.SkipDir
			}

			// Watch this path
			err = w.Watch(path)
			watched++
			if err != nil {
				log.Fatalf("Error trying to watch %s:\n%v\n%d paths watched", path, err, watched)
			}
		}
		return nil
	}

	err := filepath.Walk(path, f)
	if err != nil && err != filepath.SkipDir {
		log.Fatal("Error walking tree: %v\n", err)
	}
	return
}

// Starts watching path for changes
func Watch(path string, cmd string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Channel to notify builder goroutine to rebuild
	// Buffered, but it should never have more than 1 item
	buildChan := make(chan bool, 1)

	// Split command to run on spaces
	cmdArgs := strings.Split(cmd, " ")

	// Start builder goroutine
	go builder(path, cmdArgs, buildChan)

	watched := walkAndWatch(path, watcher)
	log.Printf("Watching %d directories\n", watched)

	// Run the command on initial run
	buildChan <- true
	for {
		e := <-watcher.Event

		// Ignore swap files
		switch {
			case strings.HasSuffix(e.Name, ".swp"):
				continue
			case strings.HasSuffix(e.Name, "~"):
				continue
		}

		log.Printf("Event: %v\n", e)

		if e.IsCreate() && e.Name != "" {
			// See if a new directory was created that needs watching
			fi, err := os.Stat(e.Name)
			if err == nil {
				if fi.IsDir() {
					// It's a new directory! Let's walk it
					watched = walkAndWatch(e.Name, watcher)
					log.Printf("Watched %d new directories", watched)
				}
			} else {
				log.Printf("Error Stat()ing %s: %v", e.Name, err)
			}
		}
		// Only signal a change if there's no pending changes
		if len(buildChan) == 0 {
			buildChan <- true
		}
	}
}
