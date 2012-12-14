package main

import (
	fsnotify "github.com/howeyc/fsnotify"
	"log"
	"os/exec"
)

var changeChan = make(chan bool, 1)


func builder(path string) {
	for {
		select {
		case <-changeChan:
			log.Printf("Received change\n")
			cmd := exec.Command("make", "html")
			out, err := cmd.Output()
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("%s", out)
		}
	}
}


func main() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	err = watcher.Watch(".")
	if err != nil {
		log.Fatal(err)
	}

	go builder(".")

	log.Printf("Watching\n")

	for {
		select {
		case e := <-watcher.Event:
			log.Printf("Event: %s %v\n", e.Name, e)
			// Only signal a change if there's no pending changes
			if len(changeChan) == 0 {
				log.Printf("Emitting change event\n")
				changeChan <- true
			}
		}
	}
}
