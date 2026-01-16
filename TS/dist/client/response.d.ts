export interface PlayerResponse {
    playabilityStatus: PlayabilityStatus;
    videoDetails: VideoDetails;
    streamingData: StreamingData;
}
export interface PlayabilityStatus {
    status: string;
    reason?: string;
    playableInEmbed?: boolean;
    liveStreamability?: LiveStreamability;
}
export interface LiveStreamability {
    liveStreamabilityRenderer: LiveStreamabilityRenderer;
}
export interface LiveStreamabilityRenderer {
    videoId: string;
    offlineSlate?: {
        liveStreamOfflineSlateRenderer: {
            scheduledStartTime?: string;
        };
    };
}
export interface VideoDetails {
    videoId: string;
    title: string;
    lengthSeconds: string;
    keywords?: string[];
    channelId: string;
    shortDescription: string;
    thumbnail: ThumbnailContainer;
    viewCount: string;
    author: string;
    isLiveContent: boolean;
    isPrivate: boolean;
    isOwnerViewing?: boolean;
}
export interface ThumbnailContainer {
    thumbnails: ThumbnailData[];
}
export interface ThumbnailData {
    url: string;
    width: number;
    height: number;
}
export interface StreamingData {
    expiresInSeconds: string;
    formats?: StreamingFormat[];
    adaptiveFormats?: StreamingFormat[];
    hlsManifestUrl?: string;
    dashManifestUrl?: string;
}
export interface StreamingFormat {
    itag: number;
    mimeType: string;
    url?: string;
    signatureCipher?: string;
    bitrate: number;
    averageBitrate?: number;
    contentLength?: string;
    width?: number;
    height?: number;
    fps?: number;
    quality: string;
    qualityLabel?: string;
    audioQuality?: string;
    audioChannels?: number;
    audioSampleRate?: string;
    approxDurationMs?: string;
    lastModified?: string;
    projectionType?: string;
    initRange?: RangeData;
    indexRange?: RangeData;
}
export interface RangeData {
    start: string;
    end: string;
}
//# sourceMappingURL=response.d.ts.map