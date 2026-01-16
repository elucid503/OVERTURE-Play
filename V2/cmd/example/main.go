package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	youtube "github.com/elucid503/overture-play/v2"
	"github.com/elucid503/overture-play/v2/client"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: example <video-id-or-url>")
		os.Exit(1)
	}

	videoInput := os.Args[1]

	// Create a new YouTube client
	client := youtube.NewWithOptions(client.ClientOptions{})

	// Fetch video information
	fmt.Printf("Fetching video: %s\n\n", videoInput)

	video, err := client.GetVideo(videoInput)
	if err != nil {
		fmt.Printf("Error fetching video: %v\n", err)
		os.Exit(1)
	}

	// Print video details
	fmt.Printf("Title:       %s\n", video.Title)
	fmt.Printf("Author:      %s\n", video.Author)
	fmt.Printf("Duration:    %d seconds\n", video.Duration)
	fmt.Printf("Views:       %d\n\n", video.ViewCount)

	// Print available formats
	fmt.Println("Video Formats:")
	for _, format := range video.VideoFormats() {
		fmt.Printf("  ITag: %-4d | %-10s | %dx%d @ %dfps | %s\n",
			format.ITag,
			format.QualityLabel,
			format.Width,
			format.Height,
			format.FPS,
			format.Extension(),
		)
	}

	fmt.Println("\nAudio Formats:")
	for _, format := range video.AudioFormats() {
		fmt.Printf("  ITag: %-4d | %-15s | %d kbps | %s\n",
			format.ITag,
			format.AudioQuality,
			format.Bitrate/1000,
			format.Extension(),
		)
	}

	// Find best audio format
	audioFormats := video.AudioFormats()
	if len(audioFormats) == 0 {
		fmt.Println("\nNo audio formats available")
		return
	}

	bestAudio := audioFormats[0]
	for _, f := range audioFormats {
		if f.Bitrate > bestAudio.Bitrate {
			bestAudio = f
		}
	}

	// Create safe filename
	safeTitle := sanitizeFilename(video.Title)
	filename := fmt.Sprintf("%s.%s", safeTitle, bestAudio.Extension())

	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("Downloading: %s\n", filename)
	fmt.Printf("Format: ITag %d, %d kbps, %s\n", bestAudio.ITag, bestAudio.Bitrate/1000, bestAudio.Extension())

	if bestAudio.ContentLength > 0 {
		fmt.Printf("Size: %.2f MB\n", float64(bestAudio.ContentLength)/1024/1024)
	}

	// Create output file
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Download with progress
	ctx := context.Background()
	startTime := time.Now()

	fmt.Println()
	err = youtube.DownloadWithProgress(ctx, bestAudio, file, func(p youtube.StreamProgress) {
		if p.Total > 0 {
			percent := float64(p.Downloaded) / float64(p.Total) * 100
			mbDownloaded := float64(p.Downloaded) / 1024 / 1024
			mbTotal := float64(p.Total) / 1024 / 1024
			speed := p.Speed / 1024 / 1024

			fmt.Printf("\r  Progress: %.1f%% (%.2f / %.2f MB) @ %.2f MB/s",
				percent, mbDownloaded, mbTotal, speed)
		} else {
			mbDownloaded := float64(p.Downloaded) / 1024 / 1024
			fmt.Printf("\r  Downloaded: %.2f MB", mbDownloaded)
		}
	})

	fmt.Println()

	if err != nil {
		fmt.Printf("\nError downloading: %v\n", err)
		os.Exit(1)
	}

	elapsed := time.Since(startTime)
	fmt.Printf("\n✓ Download complete in %.1f seconds\n", elapsed.Seconds())
	fmt.Printf("  Saved to: %s\n", filename)
}

func sanitizeFilename(name string) string {
	// Replace characters not allowed in filenames
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)

	name = replacer.Replace(name)

	// Truncate if too long
	if len(name) > 100 {
		name = name[:100]
	}

	return strings.TrimSpace(name)
}