package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/elucid503/Overture-Play/POToken"
	"github.com/elucid503/Overture-Play/Public"
)

// Tester for HLS Streaming with Automatic PO Token Generation

func main() {

	Generator := POToken.NewGenerator(nil) // Uses default settings (localhost:4416)

	// Checks if bgutil server is available

	PingResp, PingErr := Generator.Ping()

	if PingErr != nil {

		fmt.Println("bgutil server not available:", PingErr)
		fmt.Println("Run: docker run --name bgutil-provider -d -p 4416:4416 brainicism/bgutil-ytdlp-pot-provider")
		fmt.Println("Continuing without automatic PO token generation...")

		Generator = nil

	} else {

		fmt.Printf("bgutil server v%s available (uptime: %.0fs)\n", PingResp.Version, PingResp.ServerUptime)

	}

	// Fetch video info

	Response, Error := Public.Info("KqBc7R86Nbw", &Public.InfoOptions{GetHLSFormats: true}, nil, nil)

	if Error != nil {

		fmt.Println("Error Fetching:", Error.Error())
		os.Exit(1)

	}

	// Display visitor_data (used as content binding for PO token)

	fmt.Println("\nVideo Info:")
	fmt.Println("Visitor Data:", Response.VisitorData[:min(40, len(Response.VisitorData))], "...")

	if Response.DataSyncID != "" {

		fmt.Println("Data Sync ID:", Response.DataSyncID)
		
		// Extract session ID (first part before ||) for content binding
		
		SessionID := Response.DataSyncID
		
		if idx := strings.Index(Response.DataSyncID, "||"); idx > 0 {
		
			SessionID = Response.DataSyncID[:idx]
		
		}
		
		fmt.Println("Session ID (content binding):", SessionID)

	}

	// Streaming Status

	StreamingStatus := Response.JSON["playabilityStatus"].(map[string]interface{})["status"]

	fmt.Println("Status:", StreamingStatus.(string))

	if StreamingStatus.(string) != "OK" {

		fmt.Println("Cannot proceed, streaming not reported available")
		os.Exit(0)

	}

	// HLS Formats

	if len(Response.HLSFormats) < 1 {

		fmt.Println("No HLS formats available")
		os.Exit(0)

	}

	HLSManifest := Response.HLSFormats[0].URL

	fmt.Println("HLS URL:", HLSManifest[0:min(50, len(HLSManifest))], "...")

	// Configure HLS options with automatic PO token generation
	// Without cookie - the bgutil server handles its own session

	HLSOptions := &Public.HLSOptions{

		Generator:       Generator,            // Auto-generate tokens
		VisitorData:     Response.VisitorData, // Content binding for token
		DataSyncID:      "",                   // Empty - not logged in
		IsAuthenticated: false,                // Not logged in

	}

	// Pre-generate token to show it works (optional - happens automatically on first request)

	if Generator != nil {

		fmt.Println("\nGenerating PO Token for GVS (using visitor_data)...")

		// Not logged in - use visitor_data as content binding

		Token, TokenErr := Generator.GetPoTokenForGVS(Response.VisitorData, "", false)

		if TokenErr != nil {

			fmt.Println("Error generating token:", TokenErr)

		} else {

			fmt.Println("Token:", Token[:min(50, len(Token))], "...")

		}

	}

	// Fetch manifest (PO token applied automatically)

	fmt.Println("\nFetching HLS Manifest...")

	Manifest, Error := Public.GetHLSManifest(HLSManifest, HLSOptions)

	if Error != nil {

		fmt.Println("Error:", Error.Error())
		os.Exit(1)

	}

	if len(Manifest.Playlists) < 1 {

		fmt.Println("No playlists found in manifest")
		os.Exit(1)

	}

	fmt.Printf("Found %d playlists\n", len(Manifest.Playlists))

	// Fetch playlist (PO token applied automatically)

	fmt.Println("\nFetching HLS Playlist...")

	Segments, Error := Public.GetHLSPlaylist(Manifest.Playlists[0].URI, HLSOptions)

	if Error != nil {

		fmt.Println("Error:", Error.Error())
		os.Exit(1)

	}

	fmt.Printf("Found %d segments\n", len(Segments.Segments))

	// Fetch segments (PO token applied automatically)

	fmt.Println("\nFetching Segments (first 10)...")

	MaxSegments := min(10, len(Segments.Segments))

	for i := 0; i < MaxSegments; i++ {

		Segment := Segments.Segments[i]

		SegmentData, Error := Public.GetHLSSegment(Segment.URI, HLSOptions)

		if Error != nil {

			fmt.Printf("Segment %d: %v\n", i+1, Error)

		} else {

			fmt.Printf("Segment %d: %d bytes\n", i+1, len(SegmentData))

		}

		time.Sleep(1000 * time.Millisecond)

	}

	fmt.Println("\nDone!")

}