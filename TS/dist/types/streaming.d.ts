export declare const StreamingProtocol: {
    readonly HTTPS: "https";
    readonly DASH: "dash";
    readonly HLS: "hls";
};
export type StreamingProtocol = (typeof StreamingProtocol)[keyof typeof StreamingProtocol];
export type PoTokenContext = "player" | "gvs" | "subs";
export interface PoTokenPolicy {
    Required?: boolean;
    Recommended?: boolean;
    NotRequiredForPremium?: boolean;
    NotRequiredWithPlayerToken?: boolean;
}
export declare function DefaultGVSPoTokenPolicy(): PoTokenPolicy;
//# sourceMappingURL=streaming.d.ts.map