package Config

import (
	"Overture-Play/Functions"
	"Overture-Play/Utils"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sync"
	"time"
)

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

var Current = &YoutubeConfig{

	InnertubeAPIKey:     "AIzaSyAO_FJ2SlqU8Q4STEHLGCilw_Y9_11qcW8",
	InnertubeAPIVersion: "v1",

	InnertubeContext: InnertubeContext{

		Client: InnertubeClient{

			ClientName:    "WEB",
			ClientVersion: "2.20250312.04.00",
			UserAgent:     "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.5 Safari/605.1.15,gzip(gfe)",
		
		},

	},

	STS:          0,

	PlayerJSURL:  "",
	PlayerTokens: nil,

	Mutex:           sync.RWMutex{},

}

type YoutubeConfig struct {

	InnertubeAPIKey     string
	InnertubeAPIVersion string

	InnertubeContext    InnertubeContext

	STS                 int

	PlayerJSURL         string
	PlayerTokens        []string

	Mutex                  sync.RWMutex

}

func (c *YoutubeConfig) Update(Lang string) error {

	c.Mutex.Lock()
	defer c.Mutex.Unlock() // Write lock to prevent concurrent updates

	YouTubePageResp, ErrReadingBody := http.Get(fmt.Sprintf("https://www.youtube.com/?hl=%s", Lang))

	if ErrReadingBody != nil {

		return ErrReadingBody

	}

	defer YouTubePageResp.Body.Close()

	BodyBytes, ErrReadingBody := io.ReadAll(YouTubePageResp.Body)

	if ErrReadingBody != nil {

		return ErrReadingBody

	}

	body := string(BodyBytes)

	// Extract ytcfg.set configuration

	ConfigRegex := regexp.MustCompile(`ytcfg\.set\(({.+?})\);`)
	ConfigMatches := ConfigRegex.FindStringSubmatch(body)

	if len(ConfigMatches) < 2 {

		return fmt.Errorf("could not extract YouTube config")

	}

	var YTConfig map[string]interface{}

	if ParseErr := Functions.ParseJSONResponse([]byte(ConfigMatches[1]), &YTConfig); ParseErr != nil {

		return ParseErr

	}

	// Update config values

	if YouTubeAPIKey, APIKeyFetchOk := YTConfig["INNERTUBE_API_KEY"].(string); APIKeyFetchOk {

		c.InnertubeAPIKey = YouTubeAPIKey

	}

	if YouTubeAPIVersion, APIVersionFetchOk := YTConfig["INNERTUBE_API_VERSION"].(string); APIVersionFetchOk {

		c.InnertubeAPIVersion = YouTubeAPIVersion
		
	}

	if YouTubeSTS, STSFetchOk := YTConfig["STS"].(float64); STSFetchOk {

		c.STS = int(YouTubeSTS)
		
	}

	// Update player tokens if player JS URL changed

	if PlayerJSURL, ok := YTConfig["PLAYER_JS_URL"].(string); ok && PlayerJSURL != c.PlayerJSURL {

		c.PlayerJSURL = PlayerJSURL

		PlayerResp, err := http.Get(fmt.Sprintf("https://www.youtube.com%s", PlayerJSURL))

		if err == nil {

			defer PlayerResp.Body.Close()

			PlayerBytes, err := io.ReadAll(PlayerResp.Body)

			if err == nil {

				player := string(PlayerBytes)
				c.PlayerTokens = Utils.ExtractTokens(player)

			}

		}
	}

	// Schedules next update in 15 minutes

	go func() {

		time.Sleep(15 * time.Minute)
		c.Update(Lang)

	}()

	return nil

}

func (c *YoutubeConfig) GetPlayerTokens() []string {

	c.Mutex.RLock()
	defer c.Mutex.RUnlock() // Read lock to ensure no write operations occur during token retrieval

	// Return a copy to prevent race conditions if the slice is modified

	if c.PlayerTokens == nil {

		return nil

	}

	TokensCopy := make([]string, len(c.PlayerTokens))
	copy(TokensCopy, c.PlayerTokens)
	
	return TokensCopy
	
}

func (c *YoutubeConfig) GetSTS() int {

	c.Mutex.RLock()
	defer c.Mutex.RUnlock()

	return c.STS

}

func (c *YoutubeConfig) GetInnertubeAPIKey() string {

	c.Mutex.RLock()
	defer c.Mutex.RUnlock()

	return c.InnertubeAPIKey

}

func (c *YoutubeConfig) GetInnertubeClient() InnertubeClient {

	c.Mutex.RLock()
	defer c.Mutex.RUnlock()

	// Return a copy of the client struct to prevent race conditions
	return c.InnertubeContext.Client

}

func Init() {

	go Current.Update("en") // Initial call

}