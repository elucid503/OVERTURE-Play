// Types - Streaming Protocol and PO Token Policy

// StreamingProtocol enum-like object for use as both type and value
export const StreamingProtocol = {
    HTTPS: "https",
    DASH: "dash",
    HLS: "hls",
} as const;

export type StreamingProtocol = (typeof StreamingProtocol)[keyof typeof StreamingProtocol];

export type PoTokenContext = "player" | "gvs" | "subs";

export interface PoTokenPolicy {
    Required?: boolean;
    Recommended?: boolean;
    NotRequiredForPremium?: boolean;
    NotRequiredWithPlayerToken?: boolean;
}

export function DefaultGVSPoTokenPolicy(): PoTokenPolicy {
    return {
        Required: false,
        Recommended: false,
        NotRequiredForPremium: false,
        NotRequiredWithPlayerToken: false,
    };
}
