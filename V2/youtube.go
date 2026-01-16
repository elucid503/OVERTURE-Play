// Package youtube provides a Go library for fetching YouTube video information
// and downloading video/audio streams while properly handling authentication
// requirements to avoid 403 errors.
//
// Key features:
//   - PO Token generation for avoiding 403 errors
//   - Signature and n-parameter deciphering
//   - Range-based streaming for efficient downloads
//   - Support for multiple innertube clients
//   - Proper handling of authentication requirements
package youtube

import (
	"context"
	"io"

	"github.com/elucid503/overture-play/v2/client"
	"github.com/elucid503/overture-play/v2/innertube"
	"github.com/elucid503/overture-play/v2/pot"
	"github.com/elucid503/overture-play/v2/stream"
	"github.com/elucid503/overture-play/v2/types"
)

// Re-export core types for convenient access
type (
	Video     = types.Video
	Format    = types.Format
	Thumbnail = types.Thumbnail
	Range     = types.Range

	Client        = client.Client
	ClientOptions = client.ClientOptions

	StreamHandler  = stream.Handler
	StreamInfo     = stream.StreamInfo
	StreamProgress = stream.Progress

	POTProvider = pot.Provider

	ClientConfig = innertube.ClientConfig
)

// Re-export progress callback type
type ProgressCallback = stream.ProgressCallback

// New creates a new YouTube client with default options
func New() *Client {
	return client.NewClient()
}

// NewWithOptions creates a new YouTube client with the specified options
func NewWithOptions(opts ClientOptions) *Client {
	return client.NewClientWithOptions(opts)
}

// GetVideo fetches video information for the given video ID or URL
// This is a convenience function that creates a new client internally
func GetVideo(videoIDOrURL string) (*Video, error) {
	return New().GetVideo(videoIDOrURL)
}

// NewStreamHandler creates a new stream handler for downloading videos
func NewStreamHandler() *StreamHandler {
	return stream.NewHandler()
}

// Download downloads a format to a writer
// This is a convenience function for simple downloads
func Download(ctx context.Context, format Format, w io.Writer) error {
	return stream.NewHandler().Download(ctx, format, w)
}

// DownloadWithProgress downloads with progress reporting
func DownloadWithProgress(ctx context.Context, format Format, w io.Writer, callback ProgressCallback) error {
	return stream.NewHandler().DownloadWithProgress(ctx, format, w, callback)
}

// GetStream returns a reader for streaming the format
func GetStream(ctx context.Context, format Format) (io.ReadCloser, int64, error) {
	return stream.NewHandler().GetStream(ctx, format)
}

// GetStreamRange returns a reader for a specific byte range
func GetStreamRange(ctx context.Context, format Format, start, end int64) (io.ReadCloser, int64, error) {
	return stream.NewHandler().GetStreamRange(ctx, format, start, end)
}

// DefaultClients returns the default list of innertube clients used for fetching
func DefaultClients() []ClientConfig {
	return innertube.DefaultClients()
}

// DefaultWebClients returns web-based innertube clients
func DefaultWebClients() []ClientConfig {
	return innertube.DefaultWebClients()
}

// DefaultAndroidClients returns Android innertube clients
func DefaultAndroidClients() []ClientConfig {
	return innertube.DefaultAndroidClients()
}
