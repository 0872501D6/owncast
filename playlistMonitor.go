package main

import (
	"io/ioutil"
	"path"
	"path/filepath"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/radovskyb/watcher"
)

type Segment struct {
	VariantIndex       int    // The bitrate variant
	FullDiskPath       string // Where it lives on disk
	RelativeUploadPath string // Path it should have remotely
	RemoteID           string // Used for IPFS
}

type Variant struct {
	VariantIndex int
	Segments     []Segment
}

func (v *Variant) getSegmentForFilename(filename string) *Segment {
	for _, segment := range v.Segments {
		if path.Base(segment.FullDiskPath) == filename {
			return &segment
		}
	}
	return nil
}

func getSegmentFromPath(fullDiskPath string) Segment {
	segment := Segment{}
	segment.FullDiskPath = fullDiskPath
	segment.RelativeUploadPath = getRelativePathFromAbsolutePath(fullDiskPath)
	index, error := strconv.Atoi(segment.RelativeUploadPath[0:1])
	verifyError(error)
	segment.VariantIndex = index

	return segment
}

func getVariantIndexFromPath(fullDiskPath string) int {
	index, error := strconv.Atoi(fullDiskPath[0:1])
	verifyError(error)
	return index
}

var variants []Variant

func monitorVideoContent(pathToMonitor string, configuration Config, storage ChunkStorage) {
	// Create structures to store the segments for the different stream variants
	variants = make([]Variant, len(configuration.VideoSettings.StreamQualities))
	for index := range variants {
		variants[index] = Variant{index, make([]Segment, 0)}
	}

	log.Printf("Using %s for storing files with %d variants...\n", pathToMonitor, len(variants))

	w := watcher.New()

	go func() {
		for {
			select {
			case event := <-w.Event:

				relativePath := getRelativePathFromAbsolutePath(event.Path)

				// Ignore removals
				if event.Op == watcher.Remove {
					continue
				}

				// fmt.Println(event.Op, relativePath)

				// Handle updates to the master playlist by copying it to webroot
				if relativePath == path.Join(configuration.PrivateHLSPath, "stream.m3u8") {

					copy(event.Path, path.Join(configuration.PublicHLSPath, "stream.m3u8"))
					// Handle updates to playlists, but not the master playlist
				} else if filepath.Ext(event.Path) == ".m3u8" {
					variantIndex := getVariantIndexFromPath(relativePath)
					variant := variants[variantIndex]

					playlistBytes, err := ioutil.ReadFile(event.Path)
					verifyError(err)
					playlistString := string(playlistBytes)
					// fmt.Println("Rewriting playlist", relativePath, "to", path.Join(configuration.PublicHLSPath, relativePath))

					playlistString = storage.GenerateRemotePlaylist(playlistString, variant)

					writePlaylist(playlistString, path.Join(configuration.PublicHLSPath, relativePath))
				} else if filepath.Ext(event.Path) == ".ts" {
					segment := getSegmentFromPath(event.Path)

					newObjectPathChannel := make(chan string, 1)
					go func() {
						newObjectPath := storage.Save(path.Join(configuration.PrivateHLSPath, segment.RelativeUploadPath))
						newObjectPathChannel <- newObjectPath
					}()
					newObjectPath := <-newObjectPathChannel
					segment.RemoteID = newObjectPath
					// fmt.Println("Uploaded", segment.RelativeUploadPath, "as", newObjectPath)

					variants[segment.VariantIndex].Segments = append(variants[segment.VariantIndex].Segments, segment)
				}
			case err := <-w.Error:
				log.Fatalln(err)
			case <-w.Closed:
				return
			}
		}
	}()

	// Watch the hls segment storage folder recursively for changes.
	if err := w.AddRecursive(pathToMonitor); err != nil {
		log.Fatalln(err)
	}

	if err := w.Start(time.Millisecond * 100); err != nil {
		log.Fatalln(err)
	}
}
