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
var stats *Stats

var usingExternalStorage = false

func main() {
	log.Println("Starting up.  Please wait...")
	resetDirectories(configuration)
	checkConfig(configuration)
	stats = getSavedStats()
	stats.Setup()

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
		Online:                stats.IsStreamConnected(),
		ViewerCount:           stats.GetViewerCount(),
		OverallMaxViewerCount: stats.GetOverallMaxViewerCount(),
		SessionMaxViewerCount: stats.GetSessionMaxViewerCount(),
	}
	json.NewEncoder(w).Encode(status)
}

func streamConnected() {
	stats.StreamConnected()

	chunkPath := configuration.PublicHLSPath
	if usingExternalStorage {
		chunkPath = configuration.PrivateHLSPath
	}
	startThumbnailGenerator(chunkPath)
}

func streamDisconnected() {
	stats.StreamDisconnected()
}

func viewerAdded() {
	stats.SetViewerCount(server.ClientCount())
}

func viewerRemoved() {
	stats.SetViewerCount(server.ClientCount())
}
