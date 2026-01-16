// Response types for player API

// PlayerResponse is the top-level response from the player API
export interface PlayerResponse {
    playabilityStatus: PlayabilityStatus;
    videoDetails: VideoDetails;
    streamingData: StreamingData;
}

// PlayabilityStatus indicates if the video can be played
export interface PlayabilityStatus {
    status: string;
    reason?: string;
    playableInEmbed?: boolean;
    liveStreamability?: LiveStreamability;
}

// LiveStreamability contains live stream specific info
export interface LiveStreamability {
    liveStreamabilityRenderer: LiveStreamabilityRenderer;
}

// LiveStreamabilityRenderer contains live stream renderer data
export interface LiveStreamabilityRenderer {
    videoId: string;
    offlineSlate?: {
        liveStreamOfflineSlateRenderer: {
            scheduledStartTime?: string;
        };
    };
}

// VideoDetails contains video metadata
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

// ThumbnailContainer holds thumbnail data
export interface ThumbnailContainer {
    thumbnails: ThumbnailData[];
}

// ThumbnailData represents a single thumbnail
export interface ThumbnailData {
    url: string;
    width: number;
    height: number;
}

// StreamingData contains streaming format information
export interface StreamingData {
    expiresInSeconds: string;
    formats?: StreamingFormat[];
    adaptiveFormats?: StreamingFormat[];
    hlsManifestUrl?: string;
    dashManifestUrl?: string;
}

// StreamingFormat represents a playback format
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

// RangeData represents byte ranges for streaming
export interface RangeData {
    start: string;
    end: string;
}
