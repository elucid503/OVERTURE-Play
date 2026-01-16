// Innertube - Config

import { StreamingProtocol, PoTokenPolicy, DefaultGVSPoTokenPolicy } from "../types/streaming.js";

// GVSPolicies type for streaming protocol to policy mapping
export type GVSPolicies = {
    [StreamingProtocol.HTTPS]: PoTokenPolicy;
    [StreamingProtocol.DASH]: PoTokenPolicy;
    [StreamingProtocol.HLS]: PoTokenPolicy;
};

// ClientConfig represents an innertube client configuration
export interface ClientConfig {
    Name: string;
    Version: string;
    APIKey: string;

    UserAgent: string;
    DeviceMake: string;
    DeviceModel: string;
    OSName: string;
    OSVersion: string;

    Host: string;
    ContextName: number;

    RequireJSPlayer: boolean;
    SupportsCookies: boolean;
    SupportsAdPlaybackContext: boolean;
    RequireAuth: boolean;

    // PO Token policies per streaming protocol
    GVSPoTokenPolicies: GVSPolicies;
    PlayerPoTokenPolicy: PoTokenPolicy;
    SubsPoTokenPolicy: PoTokenPolicy;

    // Convenience field for direct access
    POTokenPolicy: PoTokenPolicy;
}

// InnertubeContext represents the client context sent with API requests
export interface InnertubeContext {
    client: ClientInfo;
}

// ClientInfo contains client identification information
export interface ClientInfo {
    clientName: string;
    clientVersion: string;
    userAgent?: string;
    deviceMake?: string;
    deviceModel?: string;
    osName?: string;
    osVersion?: string;
    androidSdkVersion?: number;
    hl: string;
    timeZone: string;
    utcOffsetMinutes: number;
    visitorData?: string;
}

// ThirdPartyContext for embedded player context
export interface ThirdPartyContext {
    embedUrl: string;
}

// GetContext returns an InnertubeContext for this client config
export function GetContext(c: ClientConfig): InnertubeContext {
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
export function GetContextWithVisitor(c: ClientConfig, visitorData: string): InnertubeContext {
    const ctx = GetContext(c);
    ctx.client.visitorData = visitorData;
    return ctx;
}

// RequiresPoToken returns true if this client requires a PO token
export function RequiresPoToken(c: ClientConfig): boolean {
    // Check if any GVS policy requires a token
    for (const protocol of Object.keys(c.GVSPoTokenPolicies)) {
        const policy = c.GVSPoTokenPolicies[protocol as StreamingProtocol];
        if (policy && policy.Required) {
            return true;
        }
    }
    return c.PlayerPoTokenPolicy?.Required || false;
}

// CreateClientConfig creates a ClientConfig with default values
export function CreateClientConfig(partial: Partial<ClientConfig>): ClientConfig {
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
