package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"
)

var storage ChunkStorage
var configuration = getConfig()
var server *Server

var online = false
var usingExternalStorage = false

func main() {
	log.Println("Starting up.  Please wait...")
	resetDirectories(configuration)
	checkConfig(configuration)

	if configuration.IPFS.Enabled {
		storage = &IPFSStorage{}
		usingExternalStorage = true
	} else if configuration.S3.Enabled {
		storage = &S3Storage{}
		usingExternalStorage = true
	}

	if usingExternalStorage {
		storage.Setup(configuration)
		// hlsDirectoryPath = configuration.PrivateHLSPath
		go monitorVideoContent(configuration.PrivateHLSPath, configuration, storage)
	}

	go startChatServer()

	startRTMPService()
}

func startChatServer() {
	// log.SetFlags(log.Lshortfile)

	// websocket server
	server = NewServer("/entry")
	go server.Listen()

	// static files
	http.Handle("/", http.FileServer(http.Dir("webroot")))
	http.HandleFunc("/status", getStatus)

	log.Printf("Starting public web server on port %d", configuration.WebServerPort)

	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(configuration.WebServerPort), nil))
}

func getStatus(w http.ResponseWriter, r *http.Request) {
	status := Status{
		Online:      online,
		ViewerCount: server.ClientCount(),
	}
	json.NewEncoder(w).Encode(status)
}

func streamConnected() {
	online = true
	chunkPath := configuration.PublicHLSPath
	if usingExternalStorage {
		chunkPath = configuration.PrivateHLSPath
	}
	startThumbnailGenerator(chunkPath)
}

func streamDisconnected() {
	online = false
}

func viewerAdded() {
}

func viewerRemoved() {
}
