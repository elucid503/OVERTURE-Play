import { Format } from "../types/format.js";
export declare class Handler {
    private UserAgent;
    private ChunkSize;
    private MaxRetries;
    constructor();
    SetUserAgent(userAgent: string): void;
    SetChunkSize(size: number): void;
    SetMaxRetries(retries: number): void;
    Download(format: Format, signal?: AbortSignal): Promise<Buffer>;
    DownloadRange(format: Format, start: number, end: number, signal?: AbortSignal): Promise<Buffer>;
    GetStream(format: Format, signal?: AbortSignal): Promise<{
        body: ReadableStream<Uint8Array>;
        contentLength: number;
    }>;
    GetStreamRange(format: Format, start: number, end: number, signal?: AbortSignal): Promise<{
        body: ReadableStream<Uint8Array>;
        contentLength: number;
    }>;
    private downloadWithRanges;
    private downloadChunk;
    private doChunkRequest;
    private downloadSimple;
    private getHeaders;
    GetStreamInfo(format: Format): Promise<StreamInfo>;
    DownloadWithProgress(format: Format, callback: ProgressCallback, signal?: AbortSignal): Promise<Buffer>;
}
export interface StreamInfo {
    ContentLength: number;
    ContentType: string;
    AcceptRanges: boolean;
}
export interface Progress {
    Total: number;
    Downloaded: number;
    Speed: number;
}
export type ProgressCallback = (progress: Progress) => void;
export declare function NewHandler(): Handler;
//# sourceMappingURL=handler.d.ts.map