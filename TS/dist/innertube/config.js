// Innertube - Config
import { StreamingProtocol, DefaultGVSPoTokenPolicy } from "../types/streaming.js";
// GetContext returns an InnertubeContext for this client config
export function GetContext(c) {
    return {
        client: {
            clientName: c.Name,
            clientVersion: c.Version,
            userAgent: c.UserAgent || undefined,
            deviceMake: c.DeviceMake || undefined,
            deviceModel: c.DeviceModel || undefined,
            osName: c.OSName || undefined,
            osVersion: c.OSVersion || undefined,
            hl: "en",
            timeZone: "UTC",
            utcOffsetMinutes: 0,
        },
    };
}
// GetContextWithVisitor returns an InnertubeContext with visitor data
export function GetContextWithVisitor(c, visitorData) {
    const ctx = GetContext(c);
    ctx.client.visitorData = visitorData;
    return ctx;
}
// RequiresPoToken returns true if this client requires a PO token
export function RequiresPoToken(c) {
    // Check if any GVS policy requires a token
    for (const protocol of Object.keys(c.GVSPoTokenPolicies)) {
        const policy = c.GVSPoTokenPolicies[protocol];
        if (policy && policy.Required) {
            return true;
        }
    }
    return c.PlayerPoTokenPolicy?.Required || false;
}
// CreateClientConfig creates a ClientConfig with default values
export function CreateClientConfig(partial) {
    return {
        Name: "",
        Version: "",
        APIKey: "",
        UserAgent: "",
        DeviceMake: "",
        DeviceModel: "",
        OSName: "",
        OSVersion: "",
        Host: "www.youtube.com",
        ContextName: 0,
        RequireJSPlayer: false,
        SupportsCookies: false,
        SupportsAdPlaybackContext: false,
        RequireAuth: false,
        GVSPoTokenPolicies: {
            https: DefaultGVSPoTokenPolicy(),
            dash: DefaultGVSPoTokenPolicy(),
            hls: DefaultGVSPoTokenPolicy(),
        },
        PlayerPoTokenPolicy: DefaultGVSPoTokenPolicy(),
        SubsPoTokenPolicy: DefaultGVSPoTokenPolicy(),
        POTokenPolicy: DefaultGVSPoTokenPolicy(),
        ...partial,
    };
}
//# sourceMappingURL=config.js.map