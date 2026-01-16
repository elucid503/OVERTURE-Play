// Types - Format
// HasVideo returns true if this format contains video
export function HasVideo(f) {
    return f.Width > 0 && f.Height > 0;
}
// HasAudio returns true if this format contains audio
export function HasAudio(f) {
    return f.AudioQuality !== "" || f.AudioChannels > 0 || f.AudioSampleRate > 0;
}
// IsAudioOnly returns true if this format has audio but no video
export function IsAudioOnly(f) {
    return HasAudio(f) && !HasVideo(f);
}
// IsVideoOnly returns true if this format has video but no audio
export function IsVideoOnly(f) {
    return HasVideo(f) && !HasAudio(f);
}
// IsAdaptive returns true if this is an adaptive (separate audio/video) format
export function IsAdaptive(f) {
    return IsAudioOnly(f) || IsVideoOnly(f);
}
// SupportsRange returns true if this format supports HTTP range requests
export function SupportsRange(f) {
    return f.ContentLength > 0;
}
// Extension returns the file extension for this format
export function Extension(f) {
    if (!f.MimeType) {
        return "mp4";
    }
    const parts = f.MimeType.split(";");
    if (parts.length === 0) {
        return "mp4";
    }
    const mimeType = parts[0].trim();
    switch (mimeType) {
        case "video/mp4":
            return "mp4";
        case "video/webm":
            return "webm";
        case "video/3gpp":
            return "3gp";
        case "audio/mp4":
            return "m4a";
        case "audio/webm":
            return "webm";
        case "audio/mpeg":
            return "mp3";
        default:
            if (mimeType.startsWith("video/")) {
                return "mp4";
            }
            if (mimeType.startsWith("audio/")) {
                return "m4a";
            }
            return "mp4";
    }
}
// FormatID returns a unique identifier for this format
export function FormatID(f) {
    return String(f.ITag);
}
// FormatString returns a human-readable description of the format
export function FormatString(f) {
    const parts = [];
    parts.push(`itag=${f.ITag}`);
    if (HasVideo(f)) {
        parts.push(`${f.Width}x${f.Height}`);
        if (f.FPS > 0) {
            parts.push(`${f.FPS}fps`);
        }
    }
    if (HasAudio(f)) {
        if (f.AudioSampleRate > 0) {
            parts.push(`${f.AudioSampleRate}Hz`);
        }
        if (f.AudioChannels > 0) {
            parts.push(`${f.AudioChannels}ch`);
        }
    }
    if (f.Bitrate > 0) {
        parts.push(`${Math.floor(f.Bitrate / 1000)}kbps`);
    }
    if (f.ContentLength > 0) {
        parts.push(`${(f.ContentLength / 1024 / 1024).toFixed(1)}MB`);
    }
    return parts.join(" ");
}
// QualityRank returns a numeric rank for the quality (higher is better)
export function QualityRank(quality) {
    const ranks = {
        "tiny": 0,
        "small": 1,
        "medium": 2,
        "large": 3,
        "hd720": 4,
        "hd1080": 5,
        "hd1440": 6,
        "hd2160": 7,
        "hd2880": 8,
        "highres": 9,
        "audio_quality_low": 1,
        "audio_quality_medium": 2,
        "audio_quality_high": 3,
    };
    const rank = ranks[quality.toLowerCase()];
    return rank !== undefined ? rank : -1;
}
// CreateEmptyFormat creates an empty Format with default values
export function CreateEmptyFormat() {
    return {
        ITag: 0,
        URL: "",
        MimeType: "",
        Quality: "",
        QualityLabel: "",
        Width: 0,
        Height: 0,
        FPS: 0,
        Bitrate: 0,
        AverageBitrate: 0,
        ContentLength: 0,
        AudioQuality: "",
        AudioChannels: 0,
        AudioSampleRate: 0,
        Codec: "",
        VideoCodec: "",
        AudioCodec: "",
        IndexRange: null,
        InitRange: null,
        HasDRM: false,
        SignatureCipher: "",
        Signature: "",
        SignatureParam: "",
        NParam: "",
        ClientName: "",
    };
}
//# sourceMappingURL=format.js.map