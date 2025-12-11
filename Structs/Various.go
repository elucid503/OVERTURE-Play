package Structs

import "regexp"

// Local definitions to avoid importing Config and creating an import cycle.
// These mirror the small part of Config needed for PlayerRequest JSON shape.
type InnertubeClient struct {

	ClientName    string `json:"clientName"`
	ClientVersion string `json:"clientVersion"`
	DeviceMake    string `json:"deviceMake"`
	DeviceModel   string `json:"deviceModel"`
	UserAgent     string `json:"userAgent"`
	OsName        string `json:"osName"`
	OsVersion     string `json:"osVersion"`

}

type InnertubeContext struct {

	Client InnertubeClient `json:"client"`

}

type Format struct {

	Itag             int      `json:"itag"`
	MimeType         string   `json:"mimeType"`
	QualityLabel     *string  `json:"qualityLabel"`
	Bitrate          *int     `json:"bitrate"`
	AudioBitrate     *int     `json:"audioBitrate"`
	Codec            string   `json:"codec"`
	Type             string   `json:"type"`
	Width            *int     `json:"width,omitempty"`
	Height           *int     `json:"height,omitempty"`
	InitRange        *Range   `json:"initRange,omitempty"`
	IndexRange       *Range   `json:"indexRange,omitempty"`
	LastModified     *int64   `json:"lastModifiedTimestamp,omitempty"`
	ContentLength    *int64   `json:"contentLength,omitempty"`
	Quality          *string  `json:"quality,omitempty"`
	AudioChannels    *int     `json:"audioChannels,omitempty"`
	AudioSampleRate  *int     `json:"audioSampleRate,omitempty"`
	LoudnessDb       *float64 `json:"loudnessDb,omitempty"`
	S                *string  `json:"s,omitempty"`
	Sp               *string  `json:"sp,omitempty"`
	Fps              *int     `json:"fps,omitempty"`
	AverageBitrate   *int     `json:"averageBitrate,omitempty"`
	ProjectionType   *string  `json:"projectionType,omitempty"`
	ApproxDurationMs *int64   `json:"approxDurationMs,omitempty"`
	SignatureCipher  *string  `json:"signatureCipher,omitempty"`
	URL              string   `json:"url"`
	HasAudio         bool     `json:"hasAudio"`
	HasVideo         bool     `json:"hasVideo"`
	IsLive           bool     `json:"isLive"`
	IsHLS            bool     `json:"isHLS"`
	IsDashMPD        bool     `json:"isDashMPD"`

}

type Range struct {
	
	Start int `json:"start"`
	End   int `json:"end"`

}

type Thumbnail struct {

	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	
}

type VideoDetails struct {

	ID                string      `json:"id"`
	URL               string      `json:"url"`
	Title             string      `json:"title"`
	Thumbnails        []Thumbnail `json:"thumbnails"`
	Description       string      `json:"description"`
	Duration          int64       `json:"duration"`
	ViewCount         int64       `json:"viewCount"`
	Author            string      `json:"author"`
	ChannelID         string      `json:"channelId"`
	Keywords          []string    `json:"keywords"`
	AllowRatings      bool        `json:"allowRatings"`
	AverageRating     float64     `json:"averageRating"`
	IsOwnerViewing    bool        `json:"isOwnerViewing"`
	IsCrawlable       bool        `json:"isCrawlable"`
	IsUnpluggedCorpus bool        `json:"isUnpluggedCorpus"`
	IsPrivate         bool        `json:"isPrivate"`
	IsLiveContent     bool        `json:"isLiveContent"`
	Formats           []Format    `json:"formats"`

}

// This is duplicated in Public, but should be referenced from here internally.

type Proxy struct {

	Host     string
	Port     int

	UserPass *string

}

type ContentPlaybackContext struct {

	AutoCaptionsDefaultOn bool   `json:"autoCaptionsDefaultOn"`
	AutonavState          string `json:"autonavState"`
	Html5Preference       string `json:"html5Preference"`
	LactMilliseconds      string `json:"lactMilliseconds"`
	SignatureTimestamp    int    `json:"signatureTimestamp"`

}

type PlaybackContext struct {

	ContentPlaybackContext ContentPlaybackContext `json:"contentPlaybackContext"`

}

type PlayerRequest struct {
	Context         InnertubeContext `json:"context"`
	VideoID         string           `json:"videoId"`
	PlaybackContext PlaybackContext  `json:"playbackContext"`

}

func GetMetadataFromFormat(f *Format) *Format {

	f.HasVideo = f.QualityLabel != nil && *f.QualityLabel != ""
	f.HasAudio = f.AudioBitrate != nil && *f.AudioBitrate > 0

	f.IsLive = regexp.MustCompile(`\bsource[/=]yt_(live|premiere)_broadcast\b`).MatchString(f.URL)
	f.IsHLS = regexp.MustCompile(`/manifest/hls_(variant|playlist)/`).MatchString(f.URL)
	f.IsDashMPD = regexp.MustCompile(`/manifest/dash/`).MatchString(f.URL)

	return f

}