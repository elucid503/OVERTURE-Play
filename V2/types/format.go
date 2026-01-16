package types

import (
	"fmt"
	"strconv"
	"strings"
)

// Format represents a video/audio format available for streaming
type Format struct {
	ITag         int
	URL          string
	MimeType     string
	Quality      string
	QualityLabel string

	Width  int
	Height int
	FPS    int

	Bitrate        int
	AverageBitrate int
	ContentLength  int

	AudioQuality   string
	AudioChannels  int
	AudioSampleRate int

	Codec      string
	VideoCodec string
	AudioCodec string

	// For DASH/HLS formats
	IndexRange  *Range
	InitRange   *Range

	// Whether this format has DRM protection
	HasDRM bool

	// Internal: signature cipher data if URL needs deciphering
	SignatureCipher string
	Signature       string
	SignatureParam  string

	// Internal: n parameter for throttle bypass
	NParam string

	// Client that provided this format
	ClientName string
}

// Range represents a byte range (used for DASH initialization/index)
type Range struct {
	Start int
	End   int
}

// HasVideo returns true if this format contains video
func (f *Format) HasVideo() bool {
	return f.Width > 0 && f.Height > 0
}

// HasAudio returns true if this format contains audio
func (f *Format) HasAudio() bool {
	return f.AudioQuality != "" || f.AudioChannels > 0 || f.AudioSampleRate > 0
}

// IsAudioOnly returns true if this format has audio but no video
func (f *Format) IsAudioOnly() bool {
	return f.HasAudio() && !f.HasVideo()
}

// IsVideoOnly returns true if this format has video but no audio
func (f *Format) IsVideoOnly() bool {
	return f.HasVideo() && !f.HasAudio()
}

// IsAdaptive returns true if this is an adaptive (separate audio/video) format
func (f *Format) IsAdaptive() bool {
	return f.IsAudioOnly() || f.IsVideoOnly()
}

// SupportsRange returns true if this format supports HTTP range requests
func (f *Format) SupportsRange() bool {
	return f.ContentLength > 0
}

// Extension returns the file extension for this format
func (f *Format) Extension() string {
	if f.MimeType == "" {
		return "mp4"
	}

	parts := strings.Split(f.MimeType, ";")
	if len(parts) == 0 {
		return "mp4"
	}

	mimeType := strings.TrimSpace(parts[0])

	switch mimeType {
	case "video/mp4":
		return "mp4"
	case "video/webm":
		return "webm"
	case "video/3gpp":
		return "3gp"
	case "audio/mp4":
		return "m4a"
	case "audio/webm":
		return "webm"
	case "audio/mpeg":
		return "mp3"
	default:
		if strings.HasPrefix(mimeType, "video/") {
			return "mp4"
		}
		if strings.HasPrefix(mimeType, "audio/") {
			return "m4a"
		}
		return "mp4"
	}
}

// FormatID returns a unique identifier for this format
func (f *Format) FormatID() string {
	return strconv.Itoa(f.ITag)
}

// String returns a human-readable description of the format
func (f *Format) String() string {
	var parts []string

	parts = append(parts, fmt.Sprintf("itag=%d", f.ITag))

	if f.HasVideo() {
		parts = append(parts, fmt.Sprintf("%dx%d", f.Width, f.Height))
		if f.FPS > 0 {
			parts = append(parts, fmt.Sprintf("%dfps", f.FPS))
		}
	}

	if f.HasAudio() {
		if f.AudioSampleRate > 0 {
			parts = append(parts, fmt.Sprintf("%dHz", f.AudioSampleRate))
		}
		if f.AudioChannels > 0 {
			parts = append(parts, fmt.Sprintf("%dch", f.AudioChannels))
		}
	}

	if f.Bitrate > 0 {
		parts = append(parts, fmt.Sprintf("%dkbps", f.Bitrate/1000))
	}

	if f.ContentLength > 0 {
		parts = append(parts, fmt.Sprintf("%.1fMB", float64(f.ContentLength)/1024/1024))
	}

	return strings.Join(parts, " ")
}

// Thumbnail represents a video thumbnail
type Thumbnail struct {
	URL    string
	Width  int
	Height int
}

// QualityRank returns a numeric rank for the quality (higher is better)
func QualityRank(quality string) int {
	ranks := map[string]int{
		"tiny":               0,
		"small":              1,
		"medium":             2,
		"large":              3,
		"hd720":              4,
		"hd1080":             5,
		"hd1440":             6,
		"hd2160":             7,
		"hd2880":             8,
		"highres":            9,
		"audio_quality_low":  1,
		"audio_quality_medium": 2,
		"audio_quality_high":   3,
	}

	if rank, ok := ranks[strings.ToLower(quality)]; ok {
		return rank
	}
	return -1
}
