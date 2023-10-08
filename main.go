package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
)

const port = 8080
const counterPath = "data/counter.json"

type Counter struct {
	Count int `json:"count"`
}

func main() {
	hostname, hostnameError := os.Hostname()
	if hostnameError != nil {
		panic(hostnameError)
	}
	user, userError := user.Current()
	if userError != nil {
		panic(hostnameError)
	}
	t, templateErr := template.ParseFiles("web/template/index.html")
	if templateErr != nil {
		panic(templateErr)
	}

	http.HandleFunc("/crash", handleCrash)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { handleRoot(w, r, t) })
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		// fmt.Println("Favicon requested")
	})

	fmt.Println("Welcome to Containers.")
	fmt.Printf("Service started by: %s\n", user.Username)
	fmt.Printf("Running on host: %s\n", hostname)
	fmt.Printf("Server listening on http://localhost:%d...\n", port)

	httpError := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	fmt.Printf("Fatal error: %v\n", httpError)
}

func handleCrash(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(os.Stderr, "Crash endpoint was hit. Crashing...")
	os.Exit(1)
}

func handleRoot(w http.ResponseWriter, r *http.Request, t *template.Template) {
	count, readErr := readPersistentCounter()
	if readErr != nil {
		fmt.Fprintf(os.Stderr, "Error while reading count: %v\n", readErr)
		fmt.Fprintln(w, "Error")
		return
	}

	updated := count + 1

	t.Execute(w, updated)

	writeErr := updatePersistentCounter(updated)
	if writeErr != nil {
		fmt.Fprintf(os.Stderr, "Error while updating count: %v\n", readErr)
		fmt.Fprintln(w, "Error")
	}
}

func preparePersistentCounter() {
	_, err := os.Stat(counterPath)
	if os.IsNotExist((err)) {
		if mkDirErr := os.MkdirAll(filepath.Dir(counterPath), 0755); mkDirErr != nil {
			panic(mkDirErr)
		}

		return
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while preparing persistent counter: %v\n", err)
	}
}

func readPersistentCounter() (int, error) {
	preparePersistentCounter()
	file, fileErr := os.OpenFile(counterPath, os.O_RDONLY|os.O_CREATE, 0644)
	if fileErr != nil {
		return 0, fileErr
	}

	bytes, readErr := io.ReadAll(file)
	if readErr != nil {
		return 0, readErr
	}

	closeErr := file.Close()
	if closeErr != nil {
		return 0, closeErr
	}

	var parsed Counter
	jsonErr := json.Unmarshal(bytes, &parsed)
	if jsonErr != nil {
		return 0, nil
	}

	return parsed.Count, nil
}

func updatePersistentCounter(newCount int) error {
	updated := Counter{newCount}
	bytes, jsonErr := json.Marshal(updated)
	if jsonErr != nil {
		return jsonErr
	}

	file, fileErr := os.OpenFile(counterPath, os.O_WRONLY|os.O_CREATE, 0644)
	if fileErr != nil {
		return fileErr
	}

	_, writeErr := file.Write(bytes)
	if writeErr != nil {
		return writeErr
	}

	return nil
}
