package gosphinxbuild

import (
	fsnotify "github.com/howeyc/fsnotify"
	"log"
	"os"
	exec "os/exec"
	filepath "path/filepath"
	"time"
)

// Call in its own goroutine to rebuild docs when buildChan is sent events
func builder(path string, buildChan chan bool) {
	for {
		select {
		case <-buildChan:
			log.Printf("Received change\n")
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
		log.Printf("make html >>\n%s", out)
	}
}

// Starts watching path for changes
func Watch(path string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	var walkAndWatch func(path string, fi os.FileInfo, err error) error
	watched := 0
	walkAndWatch = func(path string, fi os.FileInfo, err error) error {
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
			err = watcher.Watch(path)
			watched++
			if err != nil {
				log.Fatal(err)
			}
		}
		return nil
	}

	err = filepath.Walk(path, walkAndWatch)
	if err != nil && err != filepath.SkipDir {
		log.Fatal("Error walking tree: %v\n", err)
	}

	// Channel to notify builder goroutine to rebuild
	// Buffered, but it should never have more than 1 item
	buildChan := make(chan bool, 1)

	// Start builder goroutine
	go builder(path, buildChan)

	log.Printf("Watching %d directories\n", watched)

	for {
		select {
		case e := <-watcher.Event:
			log.Printf("Event: %s %v\n", e.Name, e)
			// Only signal a change if there's no pending changes
			if len(buildChan) == 0 {
				log.Printf("Emitting change event\n")
				buildChan <- true
			}
		}
	}
}
