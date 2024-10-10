package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	clients    = make(map[*websocket.Conn]bool)
	clientsMux sync.Mutex
)

func serveFiles(w http.ResponseWriter, r *http.Request, base string) {
	path := filepath.Join(base, r.URL.Path)
	if r.URL.Path == "/" {
		path = filepath.Join(base, "index.html")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		http.Error(w, "File not found.", http.StatusNotFound)
		return
	}

	contentType := getContentType(path)
	w.Header().Set("Content-Type", contentType)

	if contentType == "text/html" {
		injected := strings.Replace(string(data), "</body>",
			`    <script>
		// live-server injected code
        var ws = new WebSocket("ws://" + location.host + "/ws");
        ws.onmessage = function() { window.location.reload(); };
    </script>
</body>`, 1)
		w.Write([]byte(injected))
	} else {
		w.Write(data)
	}
}

func getContentType(filePath string) string {
	switch strings.ToLower(filepath.Ext(filePath)) {
	case ".html":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	default:
		return "application/octet-stream"
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}
	defer conn.Close()

	clientsMux.Lock()
	clients[conn] = true
	clientsMux.Unlock()

	conn.SetCloseHandler(func(code int, text string) error {
		clientsMux.Lock()
		delete(clients, conn)
		clientsMux.Unlock()
		return nil
	})

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func watchForChanges(watcher *fsnotify.Watcher) {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Printf("File updated: %s\n", event.Name)
				notifyClients()
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("Error:", err)
		}
	}
}

func notifyClients() {
	clientsMux.Lock()
	defer clientsMux.Unlock()

	for client := range clients {
		err := client.WriteMessage(websocket.TextMessage, []byte("reload"))
		if err != nil {
			client.Close()
			delete(clients, client)
		}
	}
}

func main() {
	dir := flag.String("dir", ".", "The directory to serve files from")
	port := flag.String("port", "8080", "Port to serve on")
	flag.Parse()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	err = filepath.Walk(*dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	go watchForChanges(watcher)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveFiles(w, r, *dir)
	})
	http.HandleFunc("/ws", handleWebSocket)

	addr := fmt.Sprintf(":%s", *port)
	fmt.Printf("Serving on http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
