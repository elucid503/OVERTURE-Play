package main

import (
	"OVERTURE/Play/Public"
	"fmt"
)

func main() {

	// Simple Tester

	Res, ParseErr := Public.Info("https://youtube.com/watch?v=dQw4w9WgXcQ", &Public.InfoOptions{ GetHLSFormats: true }, nil, nil);

	if ParseErr != nil {

		fmt.Println("Info Fetch Error:", ParseErr);
		return;

	}

	fmt.Println("HLS Formats Length:", len(Res.HLSFormats));

	// Get HLS manifest and audio segments

	HLSManifest, ParseErr := Public.GetHLSManifest(Res.HLSFormats[0].URL, nil);

	if ParseErr != nil {

		fmt.Println("HLS Parse Error:", ParseErr);
		return;

	}

	fmt.Println("HLS Playlists:", len(HLSManifest.Playlists));
	
	if len(HLSManifest.Playlists) > 0 {

		First := HLSManifest.Playlists[0];

		ParsedPlaylist, ParseErr := Public.GetHLSPlaylist(First.URI, &Public.HLSOptions{});

		if ParseErr != nil {

			fmt.Println("HLS Playlist Parse Error:", ParseErr);
			return;

		}

		fmt.Println("HLS Segments:", len(ParsedPlaylist.Segments));

		if len(ParsedPlaylist.Segments) > 0 {

			FirstSegment := ParsedPlaylist.Segments[0];

			AudioSegment, FetchErr := Public.GetHLSSegment(FirstSegment.URI, &Public.HLSOptions{});

			if FetchErr != nil {

				fmt.Println("HLS Audio Segment Fetch Error:", FetchErr);
				return;

			}

			fmt.Println("HLS Audio Segment Byte Length:", len(AudioSegment));
		
		}
		
	}

}