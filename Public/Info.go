package Public

import (
	"OVERTURE/Play/Config"
	"OVERTURE/Play/Functions"
	"OVERTURE/Play/Structs"
	"OVERTURE/Play/Utils"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type InfoOptions struct {

	GetHLSFormats bool
	
}

type Proxy struct {

	Host     string
	Port     int

	UserPass *string

}

// Info fetches the video information from YouTube's API, which will include basic metadata and streaming data.
func Info(URLOrID string, Options *InfoOptions, Proxy *Proxy, Cookies *string) (*Structs.YoutubeVideo, error) {

	if Options == nil {

		Options = &InfoOptions{GetHLSFormats: true}

	}

	 SuppliedID := Functions.GetVideoID(URLOrID)

	if SuppliedID == nil {

		return nil, fmt.Errorf("invalid video URL or ID")

	}

	var Hash *string

	if Cookies != nil {

		GeneratedHash, Err := Utils.GenerateHashFromCookies(*Cookies, "https://www.youtube.com")

		if Err == nil {

			Hash = &GeneratedHash

		}

	}
	
	RequestBody := Structs.PlayerRequest{

		Context: Structs.InnertubeContext{

			Client: Structs.InnertubeClient{

				ClientName:    Config.Current.GetInnertubeClient().ClientName,
				ClientVersion: Config.Current.GetInnertubeClient().ClientVersion,
				DeviceMake:    Config.Current.GetInnertubeClient().DeviceMake,
				DeviceModel:   Config.Current.GetInnertubeClient().DeviceModel,
				UserAgent:     Config.Current.GetInnertubeClient().UserAgent,
				OsName:        Config.Current.GetInnertubeClient().OsName,
				OsVersion:     Config.Current.GetInnertubeClient().OsVersion,

			},

		},

		VideoID: *SuppliedID,

		PlaybackContext: Structs.PlaybackContext{

			ContentPlaybackContext: Structs.ContentPlaybackContext{

				AutoCaptionsDefaultOn: false,

				AutonavState:         "STATE_NONE",
				Html5Preference:      "HTML5_PREF_WANTS",

				LactMilliseconds:     "-1",

				SignatureTimestamp:   Config.Current.GetSTS(),

			},

		},
	}

	JSONBody, Err := json.Marshal(RequestBody)

	if Err != nil {

		return nil, fmt.Errorf("error marshaling request body: %v", Err)

	}

	// Creating client

	Client := &http.Client{}

	if Proxy != nil {

		ProxyURL := Functions.GetProxyURL(&Structs.Proxy{ // Converting Proxy struct to Structs.Proxy since internally, while it is the same, the types are 'different'

			Host:     Proxy.Host,
			Port:     Proxy.Port,
			UserPass: Proxy.UserPass,

		})

		ParsedProxyURL, _ := url.Parse(ProxyURL)

		Client.Transport = &http.Transport{Proxy: http.ProxyURL(ParsedProxyURL)}

	}

	// Creating request

	Req, Err := http.NewRequest("POST", Functions.GetAPIURL(Config.Current.GetInnertubeAPIKey(), "player"), bytes.NewBuffer(JSONBody))

	if Err != nil {

		return nil, fmt.Errorf("error creating request: %v", Err)

	}

	// Setting headers

	Req.Header.Set("Origin", "https://www.youtube.com")
	Req.Header.Set("Content-Type", "application/json")
	Req.Header.Set("User-Agent", Config.Current.GetInnertubeClient().UserAgent)

	if Cookies != nil {

		Req.Header.Set("Cookie", *Cookies)

	}

	if Hash != nil {

		Req.Header.Set("Authorization", *Hash)

	}

	// Execute request

	Resp, Err := Client.Do(Req)

	if Err != nil {

		return nil, fmt.Errorf("error executing request: %v", Err)

	}

	defer Resp.Body.Close()

	// Read response

	BodyBytes, Err := io.ReadAll(Resp.Body)

	if Err != nil {

		return nil, fmt.Errorf("error reading response: %v", Err)

	}

	// Parse response

	var ParsedResp map[string]interface{}
	
	if Err := json.Unmarshal(BodyBytes, &ParsedResp); Err != nil {

		return nil, fmt.Errorf("error parsing response JSON: %v", Err)

	}

	// Checking playability status

	if PlayabilityStatus, Ok := ParsedResp["playabilityStatus"].(map[string]interface{}); Ok {

		if Status, Ok := PlayabilityStatus["status"].(string); Ok && Status == "ERROR" {

			Reason := ""

			if R, Ok := PlayabilityStatus["reason"].(string); Ok {

				Reason = R

			}

			return nil, fmt.Errorf("innertube API returned unavailable for %s: %s", URLOrID, Reason)

		}

	}

	Video := Structs.CreateYoutubeVideo(ParsedResp, Config.Current.GetPlayerTokens())

	if Options.GetHLSFormats {

		if StreamingData, Ok := ParsedResp["streamingData"].(map[string]interface{}); Ok {
			
			if ManifestURL, Ok := StreamingData["hlsManifestUrl"].(string); Ok && ManifestURL != "" {

				var ProxyStruct *Structs.Proxy

				if Proxy != nil {

					ProxyStruct = &Structs.Proxy{

						Host:     Proxy.Host,
						Port:     Proxy.Port,
						UserPass: Proxy.UserPass,

					}

				}

				HLSFormats, Err := Functions.FetchHLSFormats(ManifestURL, ProxyStruct, Config.Current.GetInnertubeClient().UserAgent)

				if Err == nil {

					Video.HLSFormats = append(Video.HLSFormats, HLSFormats...)

				}

			}

		}

	}

	return Video, nil

}