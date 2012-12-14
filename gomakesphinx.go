package gosphinxbuild

import (
	"github.com/howeyc/fsnotify"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Call in its own goroutine to rebuild docs when buildChan is sent events
func builder(path string, buildChan chan bool) {
	for {
		// Block waiting for a new event
		select {
		case <-buildChan:
		}

		// Pause briefly as editors often emit multiple events at once
		time.Sleep(100 * time.Millisecond)

		// Now just throw away the newest build change event
		select {
		case <-buildChan:
		default:
		}

		// And finally actually build the thing
		cmd := exec.Command("make", "html")
		out, err := cmd.Output()
		if err != nil {
			log.Fatalf("Error running `make html`: %v\n", err)
		}
		log.Printf("make html\n%s", out)
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

			// Skip _build direcotry
			if fi.Name() == "_build" {
					return filepath.SkipDir
			}

			// Watch this path
			err = w.Watch(path)
			watched++
			if err != nil {
				log.Fatal(err)
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
func Watch(path string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Channel to notify builder goroutine to rebuild
	// Buffered, but it should never have more than 1 item
	buildChan := make(chan bool, 1)

	// Start builder goroutine
	go builder(path, buildChan)

	watched := walkAndWatch(path, watcher)
	log.Printf("Watching %d directories\n", watched)

	for {
		select {
		case e := <-watcher.Event:
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
				}
			}
			// Only signal a change if there's no pending changes
			if len(buildChan) == 0 {
				buildChan <- true
			}
		}
	}
}
