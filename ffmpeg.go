package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
)

func startFfmpeg(configuration Config) {
	var outputDir = configuration.PublicHLSPath
	var hlsPlaylistName = path.Join(configuration.PublicHLSPath, "stream.m3u8")

	if configuration.IPFS.Enabled || configuration.S3.Enabled {
		outputDir = configuration.PrivateHLSPath
		hlsPlaylistName = path.Join(outputDir, "temp.m3u8")
	}

	log.Printf("Starting transcoder saving to /%s.", outputDir)
	pipePath := getTempPipePath()

	ffmpegCmd := "cat " + pipePath + " | " + configuration.FFMpegPath +
		" -hide_banner -i pipe: -vf scale=" + strconv.Itoa(configuration.VideoSettings.ResolutionWidth) + ":-2 -g 48 -keyint_min 48 -preset ultrafast -f hls -hls_list_size 30 -hls_time " +
		strconv.Itoa(configuration.VideoSettings.ChunkLengthInSeconds) + " -strftime 1 -use_localtime 1 -hls_segment_filename '" +
		outputDir + "/stream-%Y%m%d-%s.ts' -hls_flags delete_segments -segment_wrap 100 " + hlsPlaylistName

	_, err := exec.Command("bash", "-c", ffmpegCmd).Output()
	verifyError(err)
}

func writePlaylist(data string, filePath string) {
	f, err := os.Create(filePath)
	defer f.Close()

	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = f.WriteString(data)
	if err != nil {
		fmt.Println(err)
		return
	}
}
