// Package types provides core type definitions for the YouTube library.
package types

import "fmt"

// Video represents a YouTube video with all its metadata and available formats
type Video struct {
	ID          string
	Title       string
	Description string
	Duration    int
	ViewCount   int

	Author      string
	ChannelID   string
	ChannelURL  string
	UploadDate  string

	Thumbnails []Thumbnail
	Keywords   []string
	Category   string

	IsLive        bool
	IsLiveContent bool
	IsPrivate     bool
	AgeRestricted bool

	Formats []Format

	// Internal data for format URL resolution
	VisitorData string
	DataSyncID  string
	PlayerURL   string
}

// VideoDetails provides a simplified view of video metadata
type VideoDetails struct {
	ID          string
	URL         string
	Title       string
	Description string
	Duration    int
	ViewCount   int

	Author    string
	ChannelID string

	Thumbnails []Thumbnail
	Keywords   []string

	IsLive        bool
	IsPrivate     bool
	AgeRestricted bool
}

// URL returns the video watch URL
func (v *Video) URL() string {
	return fmt.Sprintf("https://www.youtube.com/watch?v=%s", v.ID)
}

// Details returns a simplified VideoDetails struct
func (v *Video) Details() VideoDetails {
	return VideoDetails{
		ID:            v.ID,
		URL:           v.URL(),
		Title:         v.Title,
		Description:   v.Description,
		Duration:      v.Duration,
		ViewCount:     v.ViewCount,
		Author:        v.Author,
		ChannelID:     v.ChannelID,
		Thumbnails:    v.Thumbnails,
		Keywords:      v.Keywords,
		IsLive:        v.IsLive,
		IsPrivate:     v.IsPrivate,
		AgeRestricted: v.AgeRestricted,
	}
}

// BestThumbnail returns the highest resolution thumbnail
func (v *Video) BestThumbnail() *Thumbnail {
	if len(v.Thumbnails) == 0 {
		return nil
	}

	best := &v.Thumbnails[0]
	for i := range v.Thumbnails {
		if v.Thumbnails[i].Width > best.Width {
			best = &v.Thumbnails[i]
		}
	}

	return best
}

// FilterFormats returns formats matching the given filter function
func (v *Video) FilterFormats(filter func(f Format) bool) []Format {
	var result []Format
	for _, f := range v.Formats {
		if filter(f) {
			result = append(result, f)
		}
	}
	return result
}

// VideoFormats returns only formats with video
func (v *Video) VideoFormats() []Format {
	return v.FilterFormats(func(f Format) bool {
		return f.HasVideo()
	})
}

// AudioFormats returns only formats with audio (including audio-only)
func (v *Video) AudioFormats() []Format {
	return v.FilterFormats(func(f Format) bool {
		return f.HasAudio()
	})
}

// AudioOnlyFormats returns only audio-only formats
func (v *Video) AudioOnlyFormats() []Format {
	return v.FilterFormats(func(f Format) bool {
		return f.IsAudioOnly()
	})
}

// VideoOnlyFormats returns only video-only formats (no audio)
func (v *Video) VideoOnlyFormats() []Format {
	return v.FilterFormats(func(f Format) bool {
		return f.IsVideoOnly()
	})
}

// StreamableFormats returns formats that support range requests
func (v *Video) StreamableFormats() []Format {
	return v.FilterFormats(func(f Format) bool {
		return f.ContentLength > 0
	})
}
