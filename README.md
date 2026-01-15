# github.com/elucid503/Overture-Play

## Overview

A supporting module which fetches YouTube video metadata and HLS streaming resources via an internal Innertube client. Includes **automatic PO Token generation** to prevent 403 errors and delayed 200s on segment requests.

## PO Token Support

PO (Proof of Origin) tokens are required by YouTube to authenticate streaming requests. Without a valid PO token, segment requests may return 403 errors or have delayed responses.

### Automatic PO Token Generation

This module supports **fully automatic** PO token generation using the [bgutil-ytdlp-pot-provider](https://github.com/Brainicism/bgutil-ytdlp-pot-provider) server. No manual token extraction required!

#### Quick Start

1. **Start the bgutil server** (one-time setup):
   ```bash
   docker run --name bgutil-provider -d -p 4416:4416 brainicism/bgutil-ytdlp-pot-provider
   ```

2. **Verify it's running**:
   ```bash
   curl http://127.0.0.1:4416/ping
   # Should return: {"server_uptime":...,"version":"..."}
   ```

3. **Use automatic token generation**:
   ```go
   // Create generator (connects to bgutil server at localhost:4416)
   generator := POToken.NewGenerator(nil)
   
   // Get video info (extracts visitor_data automatically)
   video, _ := Public.Info(videoID, &Public.InfoOptions{GetHLSFormats: true}, nil, nil)
   
   // Configure HLS with auto-generation
   hlsOpts := &Public.HLSOptions{
       Generator:   generator,
       VisitorData: video.VisitorData,
   }
   
   // Tokens are generated automatically on first request
   manifest, _ := Public.GetHLSManifest(video.HLSManifestURL, hlsOpts)
   ```

#### Server Options

**Docker** (recommended):
```bash
docker run --name bgutil-provider -d -p 4416:4416 brainicism/bgutil-ytdlp-pot-provider
```

**Native** (requires Node.js 18+):
```bash
git clone --single-branch --branch 1.2.2 https://github.com/Brainicism/bgutil-ytdlp-pot-provider.git
cd bgutil-ytdlp-pot-provider/server/
npm install
npx tsc
node build/main.js
```

**Custom port**:
```go
generator := POToken.NewGenerator(&POToken.GeneratorOptions{
    ServerURL: "http://127.0.0.1:8080",
})
```

### How PO Tokens Work

1. **Content Binding**: PO tokens are bound to `visitor_data` (for unauthenticated users) or `data_sync_id` (for authenticated users)
2. **URL Application**: 
   - For manifest/playlist URLs: Token is appended as `/pot/<token>` path component
   - For segment URLs: Token is appended as `?pot=<token>` query parameter
3. **Token Validity**: Tokens are valid for ~6 hours and are automatically cached

### Manual Token Usage

If you prefer to provide tokens manually (e.g., from browser extraction):

```go
hlsOpts := &Public.HLSOptions{
    PoToken: "your_base64url_encoded_po_token",
}
```

## Public Interface

### `POToken.NewGenerator(Options) → *POToken.BgUtilGenerator`

Creates a PO token generator that connects to the bgutil HTTP server.

**Parameters:**
- `Options` (*POToken.GeneratorOptions, optional):
  - `ServerURL` (string): bgutil server URL (default: `http://127.0.0.1:4416`)
  - `Timeout` (time.Duration): Request timeout (default: 30s)
  - `CacheTTL` (time.Duration): Token cache duration (default: 5h)

**Generator Methods:**
```go
// Check if server is available
pingResp, err := generator.Ping()

// Generate token for visitor_data
token, err := generator.GetPoTokenWithVisitorData(visitorData)

// Generate GVS token (handles auth vs non-auth)
token, err := generator.GetPoTokenForGVS(visitorData, dataSyncID, isAuthenticated)

// Clear token cache
generator.InvalidateCache()
```

### `Public.Info(URLOrID, Options, Proxy, Cookies) → *Structs.YoutubeVideo, error`

Calls YouTube Innertube `player` API to resolve video metadata and streaming data.

**Parameters:**
- `URLOrID` (string): A full YouTube URL or a video ID
- `Options` (*Public.InfoOptions):
  - `GetHLSFormats` (bool): When true, resolves HLS formats from the manifest URL if available
- `Proxy` (*Public.Proxy, optional): Proxy configuration
  - `Host` (string), `Port` (int), `UserPass` (*string) for basic proxy auth
- `Cookies` (*string, optional): Raw cookie header; used to derive `Authorization` for certain videos

**Returns:** Populated `Structs.YoutubeVideo` with:
- `JSON`: Raw API response
- `HLSFormats`: Parsed HLS format list (when enabled)
- `NormalFormats`: Direct streaming formats
- `VisitorData`: Visitor data for PO token generation
- `DataSyncID`: Data sync ID for authenticated PO tokens
- `HLSManifestURL`: Original HLS manifest URL (before PO token applied)

### `Public.GetHLSManifest(ManifestURL, Options) → *Structs.HLSManifest, error`

Fetches and parses an HLS master manifest, extracting playlists and audio groups.

**Parameters:**
- `ManifestURL` (string): Absolute URL to the master `.m3u8`
- `Options` (*Public.HLSOptions):
  - `Proxy` (*Public.Proxy): Optional proxy to use for requests
  - `UserAgent` (string): Overrides the default Innertube client UA
  - `PoToken` (string): Manual PO token (if not using generator)
  - `Generator` (*POToken.BgUtilGenerator): Auto-generate tokens
  - `VisitorData` (string): Content binding for auto-generation
  - `DataSyncID` (string): For authenticated users
  - `IsAuthenticated` (bool): Use DataSyncID instead of VisitorData

### `Public.GetHLSPlaylist(PlaylistURI, Options) → *Structs.HLSMediaPlaylist, error`

Fetches and parses a media playlist `.m3u8`, including segment entries.

### `Public.GetHLSSegment(SegmentURI, Options) → []byte, error`

Downloads raw bytes for a single HLS segment.

### Utility Functions

```go
// Apply PO token to a manifest or playlist URL (appends /pot/<token>)
url := Public.ApplyPoTokenToManifestURL(manifestURL, poToken)

// Apply PO token to a segment URL (appends ?pot=<token>)
url := Public.ApplyPoTokenToSegmentURL(segmentURL, poToken)

// Clean and normalize a PO token
cleanToken := Public.CleanPoToken(rawToken)
```

## Configuration

The module uses:
- `Config.Current.GetInnertubeClient()` for defaults (client name, version, user agent, etc.)
- `Config.Current.GetInnertubeAPIKey()` for API calls
- `Config.Current.GetSTS()` for signature timestamp
- `Config.Current.GetPlayerTokens()` for player tokens

## Examples

### Automatic PO Token Generation (Recommended)

```go
package main

import (
    "fmt"
    "github.com/elucid503/Overture-Play/POToken"
    "github.com/elucid503/Overture-Play/Public"
)

func main() {
    // Create generator (requires bgutil server running)
    generator := POToken.NewGenerator(nil)
    
    // Verify server is available
    if _, err := generator.Ping(); err != nil {
        fmt.Println("Start bgutil server: docker run -d -p 4416:4416 brainicism/bgutil-ytdlp-pot-provider")
        return
    }
    
    // Get video info
    video, _ := Public.Info("dQw4w9WgXcQ", &Public.InfoOptions{GetHLSFormats: true}, nil, nil)
    
    // Configure auto-generation
    hlsOpts := &Public.HLSOptions{
        Generator:   generator,
        VisitorData: video.VisitorData,
    }
    
    // Fetch HLS content (tokens generated automatically)
    manifest, _ := Public.GetHLSManifest(video.HLSManifestURL, hlsOpts)
    playlist, _ := Public.GetHLSPlaylist(manifest.Playlists[0].URI, hlsOpts)
    
    for _, segment := range playlist.Segments {
        data, _ := Public.GetHLSSegment(segment.URI, hlsOpts)
        fmt.Printf("Segment: %d bytes\n", len(data))
    }
}
```

### Manual PO Token Usage

```go
// If you have a token from an external source
hlsOpts := &Public.HLSOptions{
    PoToken: "your_base64url_encoded_po_token",
}

manifest, _ := Public.GetHLSManifest(manifestURL, hlsOpts)
```

### Direct Token Generation

```go
// Generate token directly without HLS helpers
generator := POToken.NewGenerator(nil)
token, err := generator.GetPoTokenWithVisitorData(visitorData)
if err != nil {
    // Handle error
}
fmt.Println("Generated token:", token)
```

### Authenticated User (with DataSyncID)

```go
hlsOpts := &Public.HLSOptions{
    Generator:       generator,
    VisitorData:     video.VisitorData,
    DataSyncID:      video.DataSyncID,
    IsAuthenticated: true, // Uses DataSyncID instead of VisitorData
}
```

## Troubleshooting

### bgutil server not responding
```bash
# Check if container is running
docker ps | grep bgutil

# Check logs
docker logs bgutil-provider

# Restart if needed
docker restart bgutil-provider
```

### 403 errors on segments
1. Ensure bgutil server is running
2. Verify token is being generated (check `generator.Ping()`)
3. Try invalidating cache: `generator.InvalidateCache()`

### Token generation timeout
```go
generator := POToken.NewGenerator(&POToken.GeneratorOptions{
    Timeout: 60 * time.Second, // Increase timeout
})
```