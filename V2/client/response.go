package client

// PlayerResponse is the top-level response from the player API
type PlayerResponse struct {
	PlayabilityStatus PlayabilityStatus `json:"playabilityStatus"`
	VideoDetails      VideoDetails      `json:"videoDetails"`
	StreamingData     StreamingData     `json:"streamingData"`
}

// PlayabilityStatus indicates if the video can be played
type PlayabilityStatus struct {
	Status          string `json:"status"`
	Reason          string `json:"reason"`
	PlayableInEmbed bool   `json:"playableInEmbed"`
	LiveStreamability *LiveStreamability `json:"liveStreamability"`
}

// LiveStreamability contains live stream specific info
type LiveStreamability struct {
	LiveStreamabilityRenderer LiveStreamabilityRenderer `json:"liveStreamabilityRenderer"`
}

// LiveStreamabilityRenderer contains live stream renderer data
type LiveStreamabilityRenderer struct {
	VideoID      string `json:"videoId"`
	OfflineSlate struct {
		LiveStreamOfflineSlateRenderer struct {
			ScheduledStartTime string `json:"scheduledStartTime"`
		} `json:"liveStreamOfflineSlateRenderer"`
	} `json:"offlineSlate"`
}

// VideoDetails contains video metadata
type VideoDetails struct {
	VideoID          string             `json:"videoId"`
	Title            string             `json:"title"`
	LengthSeconds    string             `json:"lengthSeconds"`
	Keywords         []string           `json:"keywords"`
	ChannelID        string             `json:"channelId"`
	ShortDescription string             `json:"shortDescription"`
	Thumbnail        ThumbnailContainer `json:"thumbnail"`
	ViewCount        string             `json:"viewCount"`
	Author           string             `json:"author"`
	IsLiveContent    bool               `json:"isLiveContent"`
	IsPrivate        bool               `json:"isPrivate"`
	IsOwnerViewing   bool               `json:"isOwnerViewing"`
}

// ThumbnailContainer holds thumbnail data
type ThumbnailContainer struct {
	Thumbnails []ThumbnailData `json:"thumbnails"`
}

// ThumbnailData represents a single thumbnail
type ThumbnailData struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// StreamingData contains streaming format information
type StreamingData struct {
	ExpiresInSeconds string            `json:"expiresInSeconds"`
	Formats          []StreamingFormat `json:"formats"`
	AdaptiveFormats  []StreamingFormat `json:"adaptiveFormats"`
	HLSManifestURL   string            `json:"hlsManifestUrl"`
	DashManifestURL  string            `json:"dashManifestUrl"`
}

// StreamingFormat represents a playback format
type StreamingFormat struct {
	ITag             int        `json:"itag"`
	MimeType         string     `json:"mimeType"`
	URL              string     `json:"url"`
	SignatureCipher  string     `json:"signatureCipher"`

	Bitrate          int        `json:"bitrate"`
	AverageBitrate   int        `json:"averageBitrate"`
	ContentLength    string     `json:"contentLength"`

	Width            int        `json:"width"`
	Height           int        `json:"height"`
	FPS              int        `json:"fps"`
	Quality          string     `json:"quality"`
	QualityLabel     string     `json:"qualityLabel"`

	AudioQuality     string     `json:"audioQuality"`
	AudioChannels    int        `json:"audioChannels"`
	AudioSampleRate  string     `json:"audioSampleRate"`

	ApproxDurationMs string     `json:"approxDurationMs"`
	LastModified     string     `json:"lastModified"`
	ProjectionType   string     `json:"projectionType"`

	InitRange        *RangeData `json:"initRange"`
	IndexRange       *RangeData `json:"indexRange"`
}

// RangeData represents byte ranges for streaming
type RangeData struct {
	Start string `json:"start"`
	End   string `json:"end"`
}
