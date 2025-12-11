package Structs

// HLSManifest represents a decoded HLS master manifest

type HLSManifest struct {

	BaseURL      string
	AudioGroups  map[string][]HLSAudioVariant
	Playlists    []HLSPlaylist

}

// HLSAudioVariant represents an audio track in the manifest

type HLSAudioVariant struct {

	URI          string
	Codecs       string
	Name         string
	Language     string
	Default      bool
	AutoSelect   bool

}

// HLSPlaylist represents a video/audio playlist

type HLSPlaylist struct {

	URI          string
	Codecs       string
	Resolution   HLSResolution
	Bandwidth    int
	FrameRate    int
	AudioGroupID string

}

// HLSResolution represents video resolution

type HLSResolution struct {

	Width  int
	Height int

}

// HLSMediaPlaylist represents a decoded media playlist with segments

type HLSMediaPlaylist struct {

	BaseURL        string
	TargetDuration int
	Segments       []HLSSegment
	IsLive         bool
	Version        int

}

// HLSSegment represents a single media segment

type HLSSegment struct {

	URI      string
	Duration float64
	Sequence int

}