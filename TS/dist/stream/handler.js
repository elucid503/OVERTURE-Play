// Package stream provides streaming downloads with range support
// Handler manages streaming downloads with range support
export class Handler {
    UserAgent;
    ChunkSize;
    MaxRetries;
    constructor() {
        this.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36";
        this.ChunkSize = 10 * 1024 * 1024; // 10MB chunks
        this.MaxRetries = 3;
    }
    // SetUserAgent sets the user agent for requests
    SetUserAgent(userAgent) {
        this.UserAgent = userAgent;
    }
    // SetChunkSize sets the chunk size for range downloads
    SetChunkSize(size) {
        this.ChunkSize = size;
    }
    // SetMaxRetries sets the maximum number of retries
    SetMaxRetries(retries) {
        this.MaxRetries = retries;
    }
    // Download downloads a format and returns the data as a Buffer
    async Download(format, signal) {
        if (!format.URL) {
            throw new Error("format has no URL");
        }
        // For formats with known content length, use range requests
        if (format.ContentLength > 0) {
            return this.downloadWithRanges(format.URL, 0, format.ContentLength, signal);
        }
        // Otherwise, simple download
        return this.downloadSimple(format.URL, signal);
    }
    // DownloadRange downloads a specific byte range
    async DownloadRange(format, start, end, signal) {
        if (!format.URL) {
            throw new Error("format has no URL");
        }
        return this.downloadWithRanges(format.URL, start, end, signal);
    }
    // GetStream returns a readable stream for the format
    async GetStream(format, signal) {
        return this.GetStreamRange(format, 0, -1, signal);
    }
    // GetStreamRange returns a readable stream for a specific byte range
    async GetStreamRange(format, start, end, signal) {
        if (!format.URL) {
            throw new Error("format has no URL");
        }
        const headers = this.getHeaders();
        // Set range header
        if (end > 0) {
            headers["Range"] = `bytes=${start}-${end}`;
        }
        else if (start > 0) {
            headers["Range"] = `bytes=${start}-`;
        }
        const resp = await fetch(format.URL, {
            method: "GET",
            headers,
            signal,
        });
        if (resp.status !== 200 && resp.status !== 206) {
            throw new Error(`unexpected status: ${resp.status}`);
        }
        const contentLength = parseInt(resp.headers.get("Content-Length") || "0", 10);
        if (!resp.body) {
            throw new Error("response has no body");
        }
        return { body: resp.body, contentLength };
    }
    // downloadWithRanges downloads using chunked range requests
    async downloadWithRanges(url, start, end, signal) {
        const chunks = [];
        let downloaded = start;
        while (downloaded < end) {
            if (signal?.aborted) {
                throw new Error("Download aborted");
            }
            let chunkEnd = downloaded + this.ChunkSize - 1;
            if (chunkEnd >= end) {
                chunkEnd = end - 1;
            }
            const chunk = await this.downloadChunk(url, downloaded, chunkEnd, signal);
            chunks.push(chunk);
            downloaded = chunkEnd + 1;
        }
        return Buffer.concat(chunks);
    }
    // downloadChunk downloads a single chunk with retries
    async downloadChunk(url, start, end, signal) {
        let lastErr = null;
        for (let attempt = 0; attempt < this.MaxRetries; attempt++) {
            if (attempt > 0) {
                if (signal?.aborted) {
                    throw new Error("Download aborted");
                }
                await sleep(attempt * 1000);
            }
            try {
                return await this.doChunkRequest(url, start, end, signal);
            }
            catch (err) {
                lastErr = err instanceof Error ? err : new Error(String(err));
            }
        }
        throw new Error(`failed after ${this.MaxRetries} retries: ${lastErr?.message}`);
    }
    // doChunkRequest performs a single chunk request
    async doChunkRequest(url, start, end, signal) {
        const headers = this.getHeaders();
        headers["Range"] = `bytes=${start}-${end}`;
        const resp = await fetch(url, {
            method: "GET",
            headers,
            signal,
        });
        if (resp.status !== 206 && resp.status !== 200) {
            throw new Error(`unexpected status: ${resp.status}`);
        }
        const arrayBuffer = await resp.arrayBuffer();
        return Buffer.from(arrayBuffer);
    }
    // downloadSimple performs a simple download without range requests
    async downloadSimple(url, signal) {
        const headers = this.getHeaders();
        const resp = await fetch(url, {
            method: "GET",
            headers,
            signal,
        });
        if (resp.status !== 200) {
            throw new Error(`unexpected status: ${resp.status}`);
        }
        const arrayBuffer = await resp.arrayBuffer();
        return Buffer.from(arrayBuffer);
    }
    // getHeaders returns required headers for requests
    getHeaders() {
        return {
            "User-Agent": this.UserAgent,
            "Accept": "*/*",
            "Accept-Language": "en-US,en;q=0.9",
            "Origin": "https://www.youtube.com",
            "Referer": "https://www.youtube.com/",
        };
    }
    // GetStreamInfo fetches stream metadata without downloading
    async GetStreamInfo(format) {
        if (!format.URL) {
            throw new Error("format has no URL");
        }
        const headers = this.getHeaders();
        const resp = await fetch(format.URL, {
            method: "HEAD",
            headers,
        });
        if (resp.status !== 200) {
            throw new Error(`unexpected status: ${resp.status}`);
        }
        const contentLength = parseInt(resp.headers.get("Content-Length") || "0", 10);
        const contentType = resp.headers.get("Content-Type") || "";
        const acceptRanges = resp.headers.get("Accept-Ranges") === "bytes";
        return {
            ContentLength: contentLength,
            ContentType: contentType,
            AcceptRanges: acceptRanges,
        };
    }
    // DownloadWithProgress downloads with progress reporting
    async DownloadWithProgress(format, callback, signal) {
        if (!format.URL) {
            throw new Error("format has no URL");
        }
        let total = format.ContentLength;
        if (total <= 0) {
            try {
                const info = await this.GetStreamInfo(format);
                if (info.ContentLength > 0) {
                    total = info.ContentLength;
                }
            }
            catch {
                // Ignore errors getting stream info
            }
        }
        const chunks = [];
        let downloaded = 0;
        const startTime = Date.now();
        if (total > 0) {
            while (downloaded < total) {
                if (signal?.aborted) {
                    throw new Error("Download aborted");
                }
                let chunkEnd = downloaded + this.ChunkSize - 1;
                if (chunkEnd >= total) {
                    chunkEnd = total - 1;
                }
                const chunk = await this.downloadChunk(format.URL, downloaded, chunkEnd, signal);
                chunks.push(chunk);
                downloaded += chunk.length;
                const elapsed = (Date.now() - startTime) / 1000;
                const speed = downloaded / elapsed;
                callback({
                    Total: total,
                    Downloaded: downloaded,
                    Speed: speed,
                });
            }
            return Buffer.concat(chunks);
        }
        // Simple download for unknown length
        const data = await this.downloadSimple(format.URL, signal);
        callback({
            Total: data.length,
            Downloaded: data.length,
            Speed: data.length / ((Date.now() - startTime) / 1000),
        });
        return data;
    }
}
// NewHandler creates a new stream handler with default settings
export function NewHandler() {
    return new Handler();
}
// Helper function
function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}
//# sourceMappingURL=handler.js.map