package stream

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/elucid503/overture-play/v2/types"
)

// Handler manages streaming downloads with range support
type Handler struct {
	HTTPClient *http.Client
	UserAgent  string

	ChunkSize  int64
	MaxRetries int
}

// NewHandler creates a new stream handler with default settings
func NewHandler() *Handler {
	return &Handler{
		HTTPClient: &http.Client{
			Timeout: 0, // No timeout for streaming
		},
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",

		ChunkSize:  10 * 1024 * 1024, // 10MB chunks
		MaxRetries: 3,
	}
}

// NewHandlerWithClient creates a handler with a custom HTTP client
func NewHandlerWithClient(client *http.Client) *Handler {
	h := NewHandler()
	h.HTTPClient = client
	return h
}

// Download downloads a format to a writer
func (h *Handler) Download(ctx context.Context, format types.Format, w io.Writer) error {
	if format.URL == "" {
		return fmt.Errorf("format has no URL")
	}

	// For formats with known content length, use range requests
	if format.ContentLength > 0 {
		return h.downloadWithRanges(ctx, format, w, 0, int64(format.ContentLength))
	}

	// Otherwise, simple download
	return h.downloadSimple(ctx, format.URL, w)
}

// DownloadRange downloads a specific byte range
func (h *Handler) DownloadRange(ctx context.Context, format types.Format, w io.Writer, start, end int64) error {
	if format.URL == "" {
		return fmt.Errorf("format has no URL")
	}

	return h.downloadWithRanges(ctx, format, w, start, end)
}

// GetStream returns a reader for streaming the format
func (h *Handler) GetStream(ctx context.Context, format types.Format) (io.ReadCloser, int64, error) {
	return h.GetStreamRange(ctx, format, 0, -1)
}

// GetStreamRange returns a reader for a specific byte range
func (h *Handler) GetStreamRange(ctx context.Context, format types.Format, start, end int64) (io.ReadCloser, int64, error) {
	if format.URL == "" {
		return nil, 0, fmt.Errorf("format has no URL")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", format.URL, nil)
	if err != nil {
		return nil, 0, err
	}

	h.setHeaders(req)

	// Set range header
	if end > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
	} else if start > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", start))
	}

	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		resp.Body.Close()
		return nil, 0, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return resp.Body, resp.ContentLength, nil
}

// downloadWithRanges downloads using chunked range requests
func (h *Handler) downloadWithRanges(ctx context.Context, format types.Format, w io.Writer, start, end int64) error {
	var downloaded int64 = start

	for downloaded < end {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		chunkEnd := downloaded + h.ChunkSize - 1
		if chunkEnd >= end {
			chunkEnd = end - 1
		}

		err := h.downloadChunk(ctx, format.URL, w, downloaded, chunkEnd)
		if err != nil {
			return err
		}

		downloaded = chunkEnd + 1
	}

	return nil
}

// downloadChunk downloads a single chunk with retries
func (h *Handler) downloadChunk(ctx context.Context, url string, w io.Writer, start, end int64) error {
	var lastErr error

	for attempt := 0; attempt < h.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(attempt) * time.Second):
			}
		}

		err := h.doChunkRequest(ctx, url, w, start, end)
		if err == nil {
			return nil
		}

		lastErr = err
	}

	return fmt.Errorf("failed after %d retries: %w", h.MaxRetries, lastErr)
}

// doChunkRequest performs a single chunk request
func (h *Handler) doChunkRequest(ctx context.Context, url string, w io.Writer, start, end int64) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	h.setHeaders(req)
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))

	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	_, err = io.Copy(w, resp.Body)
	return err
}

// downloadSimple performs a simple download without range requests
func (h *Handler) downloadSimple(ctx context.Context, url string, w io.Writer) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	h.setHeaders(req)

	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	_, err = io.Copy(w, resp.Body)
	return err
}

// setHeaders sets required headers for requests
func (h *Handler) setHeaders(req *http.Request) {
	req.Header.Set("User-Agent", h.UserAgent)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Origin", "https://www.youtube.com")
	req.Header.Set("Referer", "https://www.youtube.com/")
}

// StreamInfo contains metadata about a stream
type StreamInfo struct {
	ContentLength int64
	ContentType   string
	AcceptRanges  bool
}

// GetStreamInfo fetches stream metadata without downloading
func (h *Handler) GetStreamInfo(ctx context.Context, format types.Format) (*StreamInfo, error) {
	if format.URL == "" {
		return nil, fmt.Errorf("format has no URL")
	}

	req, err := http.NewRequestWithContext(ctx, "HEAD", format.URL, nil)
	if err != nil {
		return nil, err
	}

	h.setHeaders(req)

	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	info := &StreamInfo{
		ContentLength: resp.ContentLength,
		ContentType:   resp.Header.Get("Content-Type"),
		AcceptRanges:  resp.Header.Get("Accept-Ranges") == "bytes",
	}

	// Try to get content length from header if not set
	if info.ContentLength <= 0 {
		if cl := resp.Header.Get("Content-Length"); cl != "" {
			info.ContentLength, _ = strconv.ParseInt(cl, 10, 64)
		}
	}

	return info, nil
}

// Progress tracks download progress
type Progress struct {
	Total      int64
	Downloaded int64
	Speed      float64 // bytes per second
}

// ProgressCallback is called with download progress updates
type ProgressCallback func(Progress)

// DownloadWithProgress downloads with progress reporting
func (h *Handler) DownloadWithProgress(ctx context.Context, format types.Format, w io.Writer, callback ProgressCallback) error {
	if format.URL == "" {
		return fmt.Errorf("format has no URL")
	}

	total := int64(format.ContentLength)
	if total <= 0 {
		info, err := h.GetStreamInfo(ctx, format)
		if err == nil && info.ContentLength > 0 {
			total = info.ContentLength
		}
	}

	pw := &progressWriter{
		writer:    w,
		total:     total,
		callback:  callback,
		startTime: time.Now(),
	}

	if total > 0 {
		return h.downloadWithRanges(ctx, format, pw, 0, total)
	}

	return h.downloadSimple(ctx, format.URL, pw)
}

// progressWriter wraps a writer to track progress
type progressWriter struct {
	writer     io.Writer
	total      int64
	downloaded int64
	callback   ProgressCallback
	startTime  time.Time
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n, err := pw.writer.Write(p)
	if err != nil {
		return n, err
	}

	pw.downloaded += int64(n)

	if pw.callback != nil {
		elapsed := time.Since(pw.startTime).Seconds()
		speed := float64(pw.downloaded) / elapsed

		pw.callback(Progress{
			Total:      pw.total,
			Downloaded: pw.downloaded,
			Speed:      speed,
		})
	}

	return n, nil
}
