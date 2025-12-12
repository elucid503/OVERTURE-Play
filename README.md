# github.com/elucid503/Overture-Play

## Overview

A supporting module which fetches YouTube video metadata and HLS streaming resources via an internal Innertube client.

## Public Interface

### `Public.Info(URLOrID, Options, Proxy, Cookies) → *Structs.YoutubeVideo, error`

Calls YouTube Innertube `player` API to resolve video metadata and streaming data.

**Parameters:**
- `URLOrID` (string): A full YouTube URL or a video ID
- `Options` (*Public.InfoOptions):
  - `GetHLSFormats` (bool): When true, resolves HLS formats from the manifest URL if available
- `Proxy` (*Public.Proxy, optional): Proxy configuration
  - `Host` (string), `Port` (int), `UserPass` (*string) for basic proxy auth
- `Cookies` (*string, optional): Raw cookie header; used to derive `Authorization` for certain videos

**Returns:** Populated `Structs.YoutubeVideo` with core metadata; may include `HLSFormats` when enabled

### `Public.GetHLSManifest(ManifestURL, Options) → *Structs.HLSManifest, error`

Fetches and parses an HLS master manifest, extracting playlists and audio groups.

**Parameters:**
- `ManifestURL` (string): Absolute URL to the master `.m3u8`
- `Options` (*Public.HLSOptions):
  - `Proxy` (*Public.Proxy): Optional proxy to use for requests
  - `UserAgent` (string): Overrides the default Innertube client UA

### `Public.GetHLSPlaylist(PlaylistURI, Options) → *Structs.HLSMediaPlaylist, error`

Fetches and parses a media playlist `.m3u8`, including segment entries.

**Parameters:**
- `PlaylistURI` (string): Absolute or manifest-relative playlist URI
- `Options` (*Public.HLSOptions): Same as above

### `Public.GetHLSSegment(SegmentURI, Options) → []byte, error`

Downloads raw bytes for a single HLS segment.

**Parameters:**
- `SegmentURI` (string): Absolute or playlist-relative segment URI
- `Options` (*Public.HLSOptions): Same as above

## Configuration

The module uses:
- `Config.Current.GetInnertubeClient()` for defaults (client name, version, user agent, etc.)
- `Config.Current.GetInnertubeAPIKey()` for API calls
- `Config.Current.GetSTS()` for signature timestamp
- `Config.Current.GetPlayerTokens()` for player tokens

## Usage Notes

- If `Options` or `UserAgent` are omitted, calls default to the configured Innertube client UA
- When providing `Cookies`, an `Authorization` header may be derived for restricted content
- Proxy settings can be applied either via `Public.Info()` (HTTP client proxy) or HLS helpers (internal fetch proxy)

## Examples

### Resolve video info
```go
video, err := Public.Info("https://www.youtube.com/watch?v=<your_video_id_goes_here>", &Public.InfoOptions{GetHLSFormats: true}, nil, nil)
```

### Fetch HLS master manifest
```go
manifest, err := Public.GetHLSManifest(manifestURL, &Public.HLSOptions{UserAgent: "..."})
```

### Fetch media playlist
```go
playlist, err := Public.GetHLSPlaylist(playlistURI, &Public.HLSOptions{})
```

### Download a segment
```go
bytes, err := Public.GetHLSSegment(segmentURI, &Public.HLSOptions{})
```