package Functions

import (
	"github.com/elucid503/Overture-Play/Structs"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var (

	VideoRegex = regexp.MustCompile(`^[\w-]{11}$`)
	
	ValidPathDomains = regexp.MustCompile(`^https?://(youtu\.be/|(www\.)?youtube\.com/(embed|v|shorts)/)`)
	ValidQueryDomains = []string{"youtube.com", "www.youtube.com", "m.youtube.com", "music.youtube.com"}

)

func GetVideoURL(ID string) string {

	return fmt.Sprintf("https://www.youtube.com/watch?v=%s", ID)

}

func GetAPIURL(apiKey string, Param string) string {

	return fmt.Sprintf("https://www.youtube.com/youtubei/v1/%s?key=%s", Param, apiKey)

}

func GetVideoID(URLOrID string) *string {

	if VideoRegex.MatchString(URLOrID) {

		return &URLOrID

	}

	parsedURL, err := url.Parse(URLOrID)

	if err != nil {

		return nil

	}

	YoutubeID := parsedURL.Query().Get("v")

	if ValidPathDomains.MatchString(URLOrID) && YoutubeID == "" {

		Paths := strings.Split(parsedURL.Path, "/")

		var Pos int

		if parsedURL.Hostname() == "youtu.be" {

			Pos = 1 // For youtu.be URLs

		} else {

			Pos = 2 // For youtube.com URLs
		}

		if Pos < len(Paths) && len(Paths[Pos]) >= 11 {

			YoutubeID = Paths[Pos][:11] // Fallback: Extract the first 11 characters of the path segment

		}

	} else {

		ValidDomain := false

		for _, domain := range ValidQueryDomains {

			if parsedURL.Hostname() == domain {
				
				ValidDomain = true
				break

			}

		}

		if !ValidDomain {

			return nil

		}

	}

	if VideoRegex.MatchString(YoutubeID) {

		return &YoutubeID

	}

	return nil

}



func FetchHLSFormats(URL string, Proxy *Structs.Proxy, userAgent string) ([]Structs.Format, error) {

	var Found []Structs.Format

	HTTPClient := &http.Client{}

	if Proxy != nil {

		ProxyURL := fmt.Sprintf("http://%s:%d", Proxy.Host, Proxy.Port)

		if Proxy.UserPass != nil {

			ProxyURL = fmt.Sprintf("http://%s@%s:%d", *Proxy.UserPass, Proxy.Host, Proxy.Port)

		}

		ProxyURLParsed, _ := url.Parse(ProxyURL)
		HTTPClient.Transport = &http.Transport{Proxy: http.ProxyURL(ProxyURLParsed)}

	}

	HTTPInstance, InstanceErr := http.NewRequest("GET", URL, nil)

	if InstanceErr != nil {

		return Found, InstanceErr

	}

	HTTPInstance.Header.Set("User-Agent", userAgent)

	resp, ReqErr := HTTPClient.Do(HTTPInstance)

	if ReqErr != nil {

		return Found, ReqErr

	}

	defer resp.Body.Close() // Close the response body after reading it; no resource leaks here...

	if resp.StatusCode != http.StatusOK {

		return Found, fmt.Errorf("error on HLS format fetch request: %d %s", resp.StatusCode, resp.Status)

	}

	Body, ReqErr := io.ReadAll(resp.Body)

	if ReqErr != nil {

		return Found, ReqErr

	}

	TextResp := string(Body)

	if TextResp == "" {

		return Found, fmt.Errorf("empty HLS format response")

	}

	// Parse M3U8 manifest

	Manifest := ParseM3U8(TextResp)
	AudioFormats := make(map[string]Structs.Format)

	// Process audio groups

	for GroupID, Variants := range Manifest.AudioGroups {

		for _, AudioVariant := range Variants {

			TagMatch := regexp.MustCompile(`itag/(\d+)`).FindStringSubmatch(AudioVariant.URI)

			if len(TagMatch) < 2 {
				
				continue

			}

			Tag, _ := strconv.Atoi(TagMatch[1])

			ParsedFormat := Structs.Format{

				Itag:         Tag,
				MimeType:     "audio/mp4",
				QualityLabel: nil,
				Bitrate:      nil,
				AudioBitrate: nil,
				Codec:        AudioVariant.Codecs,
				Type:         "audio",
				URL:          AudioVariant.URI,
				HasAudio:     true,
				HasVideo:     false,
				IsLive:       false,
				IsHLS:        true,
				IsDashMPD:    false,

			}

			AudioFormats[GroupID] = ParsedFormat

		}

	}

	// Process video playlists

	for _, Playlist := range Manifest.Playlists {

		TagMatch := regexp.MustCompile(`itag/(\d+)`).FindStringSubmatch(Playlist.URI)

		if len(TagMatch) < 2 {

			continue

		}

		Tag, _ := strconv.Atoi(TagMatch[1])

		Quality := fmt.Sprintf("%dp", Playlist.Resolution.Height)

		ParsedFormat := Structs.Format{

			Itag:         Tag,
			MimeType:     fmt.Sprintf("video/mp4; codecs=\"%s\"", Playlist.Codecs),
			QualityLabel: &Quality,
			Bitrate:      &Playlist.Bandwidth,
			AudioBitrate: nil,
			Codec:        Playlist.Codecs,
			Type:         "video",
			Width:        &Playlist.Resolution.Width,
			Height:       &Playlist.Resolution.Height,
			Fps:          &Playlist.FrameRate,
			URL:          Playlist.URI,
			HasAudio:     Playlist.AudioGroupID != "",
			HasVideo:     true,
			IsLive:       false,
			IsHLS:        true,
			IsDashMPD:    false,

		}

		if Playlist.AudioGroupID != "" {

			if AudioFormat, ok := AudioFormats[Playlist.AudioGroupID]; ok {

				ParsedFormat.AudioBitrate = AudioFormat.AudioBitrate

			}
		}

		Found = append(Found, ParsedFormat)
	}

	// Add audio-only formats

	for _, AudioFormat := range AudioFormats {
		
		AlreadyAdded := false

		for _, f := range Found {

			if f.URL == AudioFormat.URL {

				AlreadyAdded = true
				break

			}

		}

		if !AlreadyAdded {

			Found = append(Found, AudioFormat)

		}

	}

	return Found, nil

}

type M3U8Manifest struct {
	
	AudioGroups map[string][]AudioVariant
	Playlists   []Playlist

}

type AudioVariant struct {

	URI    string
	Codecs string

}

type Playlist struct {

	URI          string
	Codecs       string
	Resolution   Resolution
	Bandwidth    int
	FrameRate    int
	AudioGroupID string

}

type Resolution struct {

	Width  int
	Height int

}

func ParseM3U8(Content string) M3U8Manifest {

	Manifest := M3U8Manifest{

		AudioGroups: make(map[string][]AudioVariant),
		Playlists:   []Playlist{},

	}

	ManifestLines := strings.Split(Content, "\n")

	var CurrentPlaylist *Playlist

	for i := 0; i < len(ManifestLines); i++ {

		ManifestLine := strings.TrimSpace(ManifestLines[i])

		// Parses audio media

		if strings.HasPrefix(ManifestLine, "#EXT-X-MEDIA:") {
			
			ManifestLineAttrs := ParseLineAttributes(ManifestLine)

			if ManifestLineAttrs["TYPE"] == "AUDIO" {

				GroupID := ManifestLineAttrs["GROUP-ID"]
				URI := ManifestLineAttrs["URI"]
				Codecs := ManifestLineAttrs["CODECS"]

				if GroupID != "" && URI != "" {

					Manifest.AudioGroups[GroupID] = append(Manifest.AudioGroups[GroupID], AudioVariant{

						URI:    URI,
						Codecs: Codecs,

					})

				}

			}

		}

		// Parse video stream info

		if strings.HasPrefix(ManifestLine, "#EXT-X-STREAM-INF:") {

			LineAttrs := ParseLineAttributes(ManifestLine)
			CurrentPlaylist = &Playlist{}

			if Res := LineAttrs["RESOLUTION"]; Res != "" {

				Parts := strings.Split(Res, "x")

				if len(Parts) == 2 {

					CurrentPlaylist.Resolution.Width, _ = strconv.Atoi(Parts[0])
					CurrentPlaylist.Resolution.Height, _ = strconv.Atoi(Parts[1])

				}

			}

			if Bandwidth := LineAttrs["BANDWIDTH"]; Bandwidth != "" {

				CurrentPlaylist.Bandwidth, _ = strconv.Atoi(Bandwidth)

			}

			if Frames := LineAttrs["FRAME-RATE"]; Frames != "" {

				fr, _ := strconv.ParseFloat(Frames, 64)
				CurrentPlaylist.FrameRate = int(fr)

			}

			CurrentPlaylist.Codecs = LineAttrs["CODECS"]
			CurrentPlaylist.AudioGroupID = LineAttrs["AUDIO"]

			// Next line should be the URI

			if i+1 < len(ManifestLines) {

				i++ // Moves to the next line

				CurrentPlaylist.URI = strings.TrimSpace(ManifestLines[i])
				Manifest.Playlists = append(Manifest.Playlists, *CurrentPlaylist)

			}

		}

	}

	return Manifest
}

func ParseLineAttributes(Line string) map[string]string {

	Attrs := make(map[string]string)

	// Remove the tag part

	ColonIndex := strings.Index(Line, ":")

	if ColonIndex == -1 {

		return Attrs

	}
	Line = Line[ColonIndex+1:]

	// Parse key=value pairs

	Regex := regexp.MustCompile(`([A-Z-]+)=("([^"]*)"|([^,]*))`)
	Matches := Regex.FindAllStringSubmatch(Line, -1)

	for _, match := range Matches {

		if len(match) >= 4 {

			Key := match[1]
			Val := match[3]

			if Val == "" {

				Val = match[4]

			}

			Attrs[Key] = Val

		}

	}

	return Attrs

}

func GetProxyURL(Proxy *Structs.Proxy) string {

	if Proxy == nil {

		return ""

	}

	if Proxy.UserPass != nil {

		return fmt.Sprintf("http://%s@%s:%d", *Proxy.UserPass, Proxy.Host, Proxy.Port) // Must include user:pass in the proxy URL
		
	}

	return fmt.Sprintf("http://%s:%d", Proxy.Host, Proxy.Port) // No need for user:pass

}

func ParseJSONResponse(Body []byte, TargetStruct interface{}) error {

	return json.Unmarshal(Body, TargetStruct)

}