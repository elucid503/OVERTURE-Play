// Types - Streaming Protocol and PO Token Policy
// StreamingProtocol enum-like object for use as both type and value
export const StreamingProtocol = {
    HTTPS: "https",
    DASH: "dash",
    HLS: "hls",
};
export function DefaultGVSPoTokenPolicy() {
    return {
        Required: false,
        Recommended: false,
        NotRequiredForPremium: false,
        NotRequiredWithPlayerToken: false,
    };
}
//# sourceMappingURL=streaming.js.map