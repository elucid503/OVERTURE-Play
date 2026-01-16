import { Format } from "./format.js";
import { Thumbnail } from "./thumbnail.js";
export interface Video {
    ID: string;
    Title: string;
    Description: string;
    Author: string;
    ChannelID: string;
    ChannelURL: string;
    Duration: number;
    ViewCount: number;
    PublishDate: string;
    UploadDate: string;
    IsLive: boolean;
    IsLiveContent: boolean;
    IsPlayable: boolean;
    IsPrivate: boolean;
    IsUpcoming: boolean;
    LiveBroadcastDetails: LiveBroadcastDetails | null;
    Thumbnails: Thumbnail[];
    Formats: Format[];
    DASHManifestURL: string;
    HLSManifestURL: string;
    PlayabilityStatus: PlayabilityStatus;
}
export interface PlayabilityStatus {
    Status: string;
    Reason: string;
    PlayableInEmbed: boolean;
    IsAgeRestricted: boolean;
    MiniplayerStatus: string;
}
export interface LiveBroadcastDetails {
    IsLiveNow: boolean;
    StartTimestamp: string;
    EndTimestamp: string;
}
export declare function FilterFormats(v: Video, predicate: (f: Format) => boolean): Format[];
export declare function VideoFormats(v: Video): Format[];
export declare function AudioFormats(v: Video): Format[];
export declare function BestVideoFormat(v: Video): Format | null;
export declare function BestAudioFormat(v: Video): Format | null;
export declare function GetFormat(v: Video, itag: number): Format | null;
export declare function HasLiveFormats(v: Video): boolean;
export declare function GetThumbnail(v: Video): Thumbnail | null;
export declare function DurationString(v: Video): string;
export declare function CreateEmptyVideo(): Video;
//# sourceMappingURL=video.d.ts.map