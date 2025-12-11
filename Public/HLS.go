package Public

import (
	"fmt"

	"OVERTURE/Play/Config"
	"OVERTURE/Play/Functions"
	"OVERTURE/Play/Structs"
)

// HLSOptions configures HLS manifest and playlist fetching

type HLSOptions struct {

	Proxy     *Proxy
	UserAgent string

}

// GetHLSManifest fetches and decodes an HLS master manifest, returning playlists and audio groups

func GetHLSManifest(ManifestURL string, Options *HLSOptions) (*Structs.HLSManifest, error) {

	if Options == nil {

		Options = &HLSOptions{

			UserAgent: Config.Current.GetInnertubeClient().UserAgent,

		}

	}

	if Options.UserAgent == "" {

		Options.UserAgent = Config.Current.GetInnertubeClient().UserAgent

	}

	var ProxyStruct *Structs.Proxy

	if Options.Proxy != nil {

		ProxyStruct = &Structs.Proxy{

			Host:     Options.Proxy.Host,
			Port:     Options.Proxy.Port,
			UserPass: Options.Proxy.UserPass,

		}

	}

	Content, Err := Functions.FetchHLSContent(ManifestURL, ProxyStruct, Options.UserAgent)

	if Err != nil {

		return nil, fmt.Errorf("failed to fetch HLS manifest: %v", Err)

	}

	Manifest := Functions.ParseHLSManifest(Content, ManifestURL)

	return Manifest, nil

}

// GetHLSPlaylist fetches and decodes an HLS media playlist (with segments) from a playlist URI

func GetHLSPlaylist(PlaylistURI string, Options *HLSOptions) (*Structs.HLSMediaPlaylist, error) {

	if Options == nil {

		Options = &HLSOptions{

			UserAgent: Config.Current.GetInnertubeClient().UserAgent,

		}

	}

	if Options.UserAgent == "" {

		Options.UserAgent = Config.Current.GetInnertubeClient().UserAgent

	}

	var ProxyStruct *Structs.Proxy

	if Options.Proxy != nil {

		ProxyStruct = &Structs.Proxy{

			Host:     Options.Proxy.Host,
			Port:     Options.Proxy.Port,
			UserPass: Options.Proxy.UserPass,

		}

	}

	Content, Err := Functions.FetchHLSContent(PlaylistURI, ProxyStruct, Options.UserAgent)

	if Err != nil {

		return nil, fmt.Errorf("failed to fetch HLS playlist: %v", Err)

	}

	Playlist := Functions.ParseMediaPlaylist(Content, PlaylistURI)

	return Playlist, nil

}

// GetHLSSegmentAudio fetches raw audio bytes from an HLS segment URI

func GetHLSSegment(SegmentURI string, Options *HLSOptions) ([]byte, error) {

	if Options == nil {

		Options = &HLSOptions{

			UserAgent: Config.Current.GetInnertubeClient().UserAgent,

		}

	}

	if Options.UserAgent == "" {

		Options.UserAgent = Config.Current.GetInnertubeClient().UserAgent

	}

	var ProxyStruct *Structs.Proxy

	if Options.Proxy != nil {

		ProxyStruct = &Structs.Proxy{

			Host:     Options.Proxy.Host,
			Port:     Options.Proxy.Port,
			UserPass: Options.Proxy.UserPass,

		}

	}

	Bytes, Err := Functions.FetchHLSSegmentBytes(SegmentURI, ProxyStruct, Options.UserAgent)

	if Err != nil {

		return nil, fmt.Errorf("failed to fetch HLS segment: %v", Err)

	}

	return Bytes, nil

}