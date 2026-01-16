// Types - Video
import { HasVideo, IsAudioOnly, QualityRank } from "./format.js";
// FilterFormats returns formats that match the predicate
export function FilterFormats(v, predicate) {
    return v.Formats.filter(predicate);
}
// VideoFormats returns all formats that contain video
export function VideoFormats(v) {
    return FilterFormats(v, (f) => HasVideo(f));
}
// AudioFormats returns all formats that contain audio only
export function AudioFormats(v) {
    return FilterFormats(v, (f) => IsAudioOnly(f));
}
// BestVideoFormat returns the best video format (highest resolution, then bitrate)
export function BestVideoFormat(v) {
    const formats = VideoFormats(v);
    if (formats.length === 0) {
        return null;
    }
    return formats.reduce((best, current) => {
        // First compare by resolution
        const currentPixels = current.Width * current.Height;
        const bestPixels = best.Width * best.Height;
        if (currentPixels !== bestPixels) {
            return currentPixels > bestPixels ? current : best;
        }
        // Then by bitrate
        if (current.Bitrate !== best.Bitrate) {
            return current.Bitrate > best.Bitrate ? current : best;
        }
        return best;
    });
}
// BestAudioFormat returns the best audio format (highest bitrate)
export function BestAudioFormat(v) {
    const formats = AudioFormats(v);
    if (formats.length === 0) {
        return null;
    }
    return formats.reduce((best, current) => {
        // Compare by audio quality rank first
        const currentRank = QualityRank(current.AudioQuality);
        const bestRank = QualityRank(best.AudioQuality);
        if (currentRank !== bestRank) {
            return currentRank > bestRank ? current : best;
        }
        // Then by bitrate
        if (current.Bitrate !== best.Bitrate) {
            return current.Bitrate > best.Bitrate ? current : best;
        }
        return best;
    });
}
// GetFormat returns the format with the given itag
export function GetFormat(v, itag) {
    return v.Formats.find((f) => f.ITag === itag) || null;
}
// HasLiveFormats returns true if the video has live streaming formats (HLS/DASH)
export function HasLiveFormats(v) {
    return v.DASHManifestURL !== "" || v.HLSManifestURL !== "";
}
// GetThumbnail returns the best thumbnail (highest resolution)
export function GetThumbnail(v) {
    if (v.Thumbnails.length === 0) {
        return null;
    }
    return v.Thumbnails.reduce((best, current) => {
        const currentPixels = current.Width * current.Height;
        const bestPixels = best.Width * best.Height;
        return currentPixels > bestPixels ? current : best;
    });
}
// DurationString returns the duration as a human-readable string (HH:MM:SS)
export function DurationString(v) {
    const hours = Math.floor(v.Duration / 3600);
    const minutes = Math.floor((v.Duration % 3600) / 60);
    const seconds = v.Duration % 60;
    if (hours > 0) {
        return `${hours}:${minutes.toString().padStart(2, "0")}:${seconds.toString().padStart(2, "0")}`;
    }
    return `${minutes}:${seconds.toString().padStart(2, "0")}`;
}
// CreateEmptyVideo creates an empty Video with default values
export function CreateEmptyVideo() {
    return {
        ID: "",
        Title: "",
        Description: "",
        Author: "",
        ChannelID: "",
        ChannelURL: "",
        Duration: 0,
        ViewCount: 0,
        PublishDate: "",
        UploadDate: "",
        IsLive: false,
        IsLiveContent: false,
        IsPlayable: false,
        IsPrivate: false,
        IsUpcoming: false,
        LiveBroadcastDetails: null,
        Thumbnails: [],
        Formats: [],
        DASHManifestURL: "",
        HLSManifestURL: "",
        PlayabilityStatus: {
            Status: "",
            Reason: "",
            PlayableInEmbed: false,
            IsAgeRestricted: false,
            MiniplayerStatus: "",
        },
    };
}
//# sourceMappingURL=video.js.map