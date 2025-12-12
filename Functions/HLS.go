package Functions

import (
	"github.com/elucid503/Overture-Play/Structs"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// ParseHLSManifest parses an HLS master manifest and returns structured data
// If the content is actually a media playlist (with segments), it creates a single playlist entry

func ParseHLSManifest(Content string, BaseURL string) *Structs.HLSManifest {

	Manifest := &Structs.HLSManifest{

		BaseURL:     BaseURL,
		AudioGroups: make(map[string][]Structs.HLSAudioVariant),
		Playlists:   []Structs.HLSPlaylist{},

	}

	// Check if this is a media playlist (contains segments) rather than a master manifest
	if strings.Contains(Content, "#EXTINF:") && !strings.Contains(Content, "#EXT-X-STREAM-INF:") {
		// This is a media playlist with segments, not a master manifest
		// Create a single playlist entry pointing to this URL
		Manifest.Playlists = append(Manifest.Playlists, Structs.HLSPlaylist{
			URI: BaseURL,
		})
		return Manifest
	}

	Lines := strings.Split(Content, "\n")

	var CurrentPlaylist *Structs.HLSPlaylist

	for i := 0; i < len(Lines); i++ {

		Line := strings.TrimSpace(Lines[i])

		// Parse audio media

		if strings.HasPrefix(Line, "#EXT-X-MEDIA:") {

			Attrs := ParseLineAttributes(Line)

			if Attrs["TYPE"] == "AUDIO" {

				GroupID := Attrs["GROUP-ID"]
				URI := Attrs["URI"]
				Codecs := Attrs["CODECS"]
				Name := Attrs["NAME"]
				Language := Attrs["LANGUAGE"]
				Default := Attrs["DEFAULT"] == "YES"
				AutoSelect := Attrs["AUTOSELECT"] == "YES"

				if GroupID != "" && URI != "" {

					Manifest.AudioGroups[GroupID] = append(Manifest.AudioGroups[GroupID], Structs.HLSAudioVariant{

						URI:        ResolveURL(BaseURL, URI),
						Codecs:     Codecs,
						Name:       Name,
						Language:   Language,
						Default:    Default,
						AutoSelect: AutoSelect,

					})

				}

			}

		}

		// Parse video stream info

		if strings.HasPrefix(Line, "#EXT-X-STREAM-INF:") {

			Attrs := ParseLineAttributes(Line)
			CurrentPlaylist = &Structs.HLSPlaylist{}

			if Res := Attrs["RESOLUTION"]; Res != "" {

				Parts := strings.Split(Res, "x")

				if len(Parts) == 2 {

					CurrentPlaylist.Resolution.Width, _ = strconv.Atoi(Parts[0])
					CurrentPlaylist.Resolution.Height, _ = strconv.Atoi(Parts[1])

				}

			}

			if Bandwidth := Attrs["BANDWIDTH"]; Bandwidth != "" {

				CurrentPlaylist.Bandwidth, _ = strconv.Atoi(Bandwidth)

			}

			if Frames := Attrs["FRAME-RATE"]; Frames != "" {

				FrameRate, _ := strconv.ParseFloat(Frames, 64)
				CurrentPlaylist.FrameRate = int(FrameRate)

			}

			CurrentPlaylist.Codecs = Attrs["CODECS"]
			CurrentPlaylist.AudioGroupID = Attrs["AUDIO"]

			// Next line should be the URI

			if i+1 < len(Lines) {

				i++

				URILine := strings.TrimSpace(Lines[i])
				CurrentPlaylist.URI = ResolveURL(BaseURL, URILine)
				Manifest.Playlists = append(Manifest.Playlists, *CurrentPlaylist)

			}

		}

	}

	return Manifest

}

// ParseMediaPlaylist parses an HLS media playlist containing segments

func ParseMediaPlaylist(Content string, BaseURL string) *Structs.HLSMediaPlaylist {

	Playlist := &Structs.HLSMediaPlaylist{

		BaseURL:  BaseURL,
		Segments: []Structs.HLSSegment{},
		Version:  3,

	}

	// Normalizes line endings (handle \r\n)

	Content = strings.ReplaceAll(Content, "\r\n", "\n")
	Content = strings.ReplaceAll(Content, "\r", "\n")
	
	Lines := strings.Split(Content, "\n")

	Sequence := 0
	var CurrentDuration float64

	for i := 0; i < len(Lines); i++ {

		Line := strings.TrimSpace(Lines[i])

		if Line == "" {

			continue; // Skip empty lines

		}

		// Parse version

		if VersionInfo, VersionInfoOK := strings.CutPrefix(Line, "#EXT-X-VERSION:"); VersionInfoOK  {

			Version, _ := strconv.Atoi(VersionInfo)
			Playlist.Version = Version

		}

		// Parse target duration

		if TargetDuration, TargetDurOK := strings.CutPrefix(Line, "#EXT-X-TARGETDURATION:"); TargetDurOK  {

			Duration, _ := strconv.Atoi(TargetDuration)
			Playlist.TargetDuration = Duration

		}

		// Check if live stream

		if PlaylistType, PlaylistTypeOK := strings.CutPrefix(Line, "#EXT-X-PLAYLIST-TYPE:"); PlaylistTypeOK  {

			Playlist.IsLive = PlaylistType != "VOD"

		}

		// Parse segment duration

		if strings.HasPrefix(Line, "#EXTINF:") {

			DurationStr := strings.TrimPrefix(Line, "#EXTINF:")
			DurationStr = strings.Split(DurationStr, ",")[0]
			CurrentDuration, _ = strconv.ParseFloat(DurationStr, 64)

		}

		// Parse segment URI - any non-comment line is a segment

		if !strings.HasPrefix(Line, "#") {

			Segment := Structs.HLSSegment{

				URI:      ResolveURL(BaseURL, Line),
				Duration: CurrentDuration,
				Sequence: Sequence,

			}

			Playlist.Segments = append(Playlist.Segments, Segment)
			Sequence++
			CurrentDuration = 0

		}

	}

	return Playlist

}

// FetchHLSContent fetches content from an HLS URL with optional proxy support

func FetchHLSContent(URL string, Proxy *Structs.Proxy, UserAgent string) (string, error) {

	Client := &http.Client{}

	if Proxy != nil {

		ProxyURL := GetProxyURL(Proxy)
		ParsedProxyURL, _ := url.Parse(ProxyURL)
		Client.Transport = &http.Transport{Proxy: http.ProxyURL(ParsedProxyURL)}

	}

	Req, Err := http.NewRequest("GET", URL, nil)

	if Err != nil {

		return "", fmt.Errorf("error creating request: %v", Err)

	}

	Req.Header.Set("User-Agent", UserAgent)

	Resp, Err := Client.Do(Req)

	if Err != nil {

		return "", fmt.Errorf("error executing request: %v", Err)

	}

	defer Resp.Body.Close()

	if Resp.StatusCode != http.StatusOK {

		return "", fmt.Errorf("HTTP error: %d %s", Resp.StatusCode, Resp.Status)

	}

	Body, Err := io.ReadAll(Resp.Body)

	if Err != nil {

		return "", fmt.Errorf("error reading response: %v", Err)

	}

	return string(Body), nil

}

// FetchHLSSegmentBytes fetches raw bytes from an HLS segment

func FetchHLSSegmentBytes(URL string, Proxy *Structs.Proxy, UserAgent string) ([]byte, error) {

	Client := &http.Client{}

	if Proxy != nil {

		ProxyURL := GetProxyURL(Proxy)
		ParsedProxyURL, _ := url.Parse(ProxyURL)
		Client.Transport = &http.Transport{Proxy: http.ProxyURL(ParsedProxyURL)}

	}

	Req, Err := http.NewRequest("GET", URL, nil)

	if Err != nil {

		return nil, fmt.Errorf("error creating request: %v", Err)

	}

	Req.Header.Set("User-Agent", UserAgent)

	Resp, Err := Client.Do(Req)

	if Err != nil {

		return nil, fmt.Errorf("error executing request: %v", Err)

	}

	defer Resp.Body.Close()

	if Resp.StatusCode != http.StatusOK {

		return nil, fmt.Errorf("HTTP error: %d %s", Resp.StatusCode, Resp.Status)

	}

	Bytes, Err := io.ReadAll(Resp.Body)

	if Err != nil {

		return nil, fmt.Errorf("error reading response: %v", Err)

	}

	return Bytes, nil

}

// ResolveURL resolves a relative URL against a base URL

func ResolveURL(BaseURL string, RelativeURL string) string {

	if strings.HasPrefix(RelativeURL, "http://") || strings.HasPrefix(RelativeURL, "https://") {

		return RelativeURL

	}

	Base, Err := url.Parse(BaseURL)

	if Err != nil {

		return RelativeURL

	}

	Relative, Err := url.Parse(RelativeURL)

	if Err != nil {

		return RelativeURL

	}

	Resolved := Base.ResolveReference(Relative)

	return Resolved.String()

}