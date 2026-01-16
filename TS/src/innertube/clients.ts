// Innertube - Clients

import { ClientConfig, CreateClientConfig } from "./config.js";

// Predefined client configurations based on yt-dlp reference
// These are the clients that work best for avoiding 403 errors

// Web is the standard web browser client
export const Web: ClientConfig = CreateClientConfig({
    Name: "WEB",
    Version: "2.20250925.01.00",
    Host: "www.youtube.com",
    ContextName: 1,

    RequireJSPlayer: true,
    SupportsCookies: true,
    SupportsAdPlaybackContext: true,

    GVSPoTokenPolicies: {
        https: { Required: true, Recommended: true, NotRequiredForPremium: true },
        dash: { Required: true, Recommended: true, NotRequiredForPremium: true },
        hls: { Required: false, Recommended: true },
    },
    PlayerPoTokenPolicy: { Required: false },
    SubsPoTokenPolicy: { Required: false },
});

// WebSafari returns HLS formats with pre-merged video+audio
export const WebSafari: ClientConfig = CreateClientConfig({
    Name: "WEB",
    Version: "2.20250925.01.00",
    Host: "www.youtube.com",
    ContextName: 1,
    UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.5 Safari/605.1.15,gzip(gfe)",

    RequireJSPlayer: true,
    SupportsCookies: true,
    SupportsAdPlaybackContext: true,

    GVSPoTokenPolicies: {
        https: { Required: true, Recommended: true, NotRequiredForPremium: true },
        dash: { Required: true, Recommended: true, NotRequiredForPremium: true },
        hls: { Required: false, Recommended: true },
    },
    PlayerPoTokenPolicy: { Required: false },
    SubsPoTokenPolicy: { Required: false },
});

// WebEmbedded is for embedded player context
export const WebEmbedded: ClientConfig = CreateClientConfig({
    Name: "WEB_EMBEDDED_PLAYER",
    Version: "1.20250923.21.00",
    Host: "www.youtube.com",
    ContextName: 56,

    RequireJSPlayer: true,
    SupportsCookies: true,

    GVSPoTokenPolicies: {
        https: { Required: false },
        dash: { Required: false },
        hls: { Required: false },
    },
});

// WebMusic is YouTube Music client
export const WebMusic: ClientConfig = CreateClientConfig({
    Name: "WEB_REMIX",
    Version: "1.20250922.03.00",
    Host: "music.youtube.com",
    ContextName: 67,

    RequireJSPlayer: true,
    SupportsCookies: true,
    SupportsAdPlaybackContext: true,

    GVSPoTokenPolicies: {
        https: { Required: true, Recommended: true, NotRequiredForPremium: true },
        dash: { Required: true, Recommended: true, NotRequiredForPremium: true },
        hls: { Required: false, Recommended: true },
    },
});

// WebCreator requires authentication
export const WebCreator: ClientConfig = CreateClientConfig({
    Name: "WEB_CREATOR",
    Version: "1.20250922.03.00",
    Host: "www.youtube.com",
    ContextName: 62,

    RequireJSPlayer: true,
    RequireAuth: true,
    SupportsCookies: true,

    GVSPoTokenPolicies: {
        https: { Required: true, Recommended: true, NotRequiredForPremium: true },
        dash: { Required: true, Recommended: true, NotRequiredForPremium: true },
        hls: { Required: false, Recommended: true },
    },
});

// Android is the Android app client
export const Android: ClientConfig = CreateClientConfig({
    Name: "ANDROID",
    Version: "20.10.38",
    Host: "www.youtube.com",
    ContextName: 3,
    UserAgent: "com.google.android.youtube/20.10.38 (Linux; U; Android 11) gzip",
    OSName: "Android",
    OSVersion: "11",

    RequireJSPlayer: false,

    GVSPoTokenPolicies: {
        https: { Required: true, Recommended: true, NotRequiredWithPlayerToken: true },
        dash: { Required: true, Recommended: true, NotRequiredWithPlayerToken: true },
        hls: { Required: false, Recommended: true, NotRequiredWithPlayerToken: true },
    },
    PlayerPoTokenPolicy: { Required: false, Recommended: true },
});

// AndroidSDKLess doesn't require PO Token (useful fallback)
export const AndroidSDKLess: ClientConfig = CreateClientConfig({
    Name: "ANDROID",
    Version: "20.10.38",
    Host: "www.youtube.com",
    ContextName: 3,
    UserAgent: "com.google.android.youtube/20.10.38 (Linux; U; Android 11) gzip",
    OSName: "Android",
    OSVersion: "11",

    RequireJSPlayer: false,

    // No PO token policies - this client doesn't require them
    GVSPoTokenPolicies: {
        https: { Required: false },
        dash: { Required: false },
        hls: { Required: false },
    },
});

// AndroidVR is Oculus Quest client (doesn't return Kids videos)
export const AndroidVR: ClientConfig = CreateClientConfig({
    Name: "ANDROID_VR",
    Version: "1.65.10",
    Host: "www.youtube.com",
    ContextName: 28,
    UserAgent: "com.google.android.apps.youtube.vr.oculus/1.65.10 (Linux; U; Android 12L; eureka-user Build/SQ3A.220605.009.A1) gzip",
    DeviceMake: "Oculus",
    DeviceModel: "Quest 3",
    OSName: "Android",
    OSVersion: "12L",

    RequireJSPlayer: false,

    GVSPoTokenPolicies: {
        https: { Required: false },
        dash: { Required: false },
        hls: { Required: false },
    },
});

// IOS is the iOS app client (provides HLS live streams)
export const IOS: ClientConfig = CreateClientConfig({
    Name: "IOS",
    Version: "20.10.4",
    Host: "www.youtube.com",
    ContextName: 5,
    UserAgent: "com.google.ios.youtube/20.10.4 (iPhone16,2; U; CPU iOS 18_3_2 like Mac OS X;)",
    DeviceMake: "Apple",
    DeviceModel: "iPhone16,2",
    OSName: "iPhone",
    OSVersion: "18.3.2.22D82",

    RequireJSPlayer: false,

    GVSPoTokenPolicies: {
        https: { Required: true, Recommended: true, NotRequiredWithPlayerToken: true },
        dash: { Required: false },
        hls: { Required: true, Recommended: true, NotRequiredWithPlayerToken: true },
    },
    PlayerPoTokenPolicy: { Required: false, Recommended: true },
});

// MWeb is the mobile web client (has ultralow formats)
export const MWeb: ClientConfig = CreateClientConfig({
    Name: "MWEB",
    Version: "2.20250925.01.00",
    Host: "www.youtube.com",
    ContextName: 2,
    UserAgent: "Mozilla/5.0 (iPad; CPU OS 16_7_10 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.6 Mobile/15E148 Safari/604.1,gzip(gfe)",

    RequireJSPlayer: true,
    SupportsCookies: true,
    SupportsAdPlaybackContext: true,

    GVSPoTokenPolicies: {
        https: { Required: true, Recommended: true, NotRequiredForPremium: true },
        dash: { Required: true, Recommended: true, NotRequiredForPremium: true },
        hls: { Required: false, Recommended: true },
    },
});

// TV is the smart TV client
export const TV: ClientConfig = CreateClientConfig({
    Name: "TVHTML5",
    Version: "7.20250923.13.00",
    Host: "www.youtube.com",
    ContextName: 7,
    UserAgent: "Mozilla/5.0 (ChromiumStylePlatform) Cobalt/Version",

    RequireJSPlayer: true,
    SupportsCookies: true,

    GVSPoTokenPolicies: {
        https: { Required: false },
        dash: { Required: false },
        hls: { Required: false },
    },
});

// TVDowngraded is the older TV client (works better for some videos)
export const TVDowngraded: ClientConfig = CreateClientConfig({
    Name: "TVHTML5",
    Version: "5.20251105",
    Host: "www.youtube.com",
    ContextName: 7,
    UserAgent: "Mozilla/5.0 (ChromiumStylePlatform) Cobalt/Version",

    RequireJSPlayer: true,
    SupportsCookies: true,

    GVSPoTokenPolicies: {
        https: { Required: false },
        dash: { Required: false },
        hls: { Required: false },
    },
});

// TVSimply is a simplified TV client
export const TVSimply: ClientConfig = CreateClientConfig({
    Name: "TVHTML5_SIMPLY",
    Version: "1.0",
    Host: "www.youtube.com",
    ContextName: 75,

    RequireJSPlayer: true,

    GVSPoTokenPolicies: {
        https: { Required: true, Recommended: true },
        dash: { Required: true, Recommended: true },
        hls: { Required: false, Recommended: true },
    },
});

// TVEmbedded is for TV embedded player (requires auth)
export const TVEmbedded: ClientConfig = CreateClientConfig({
    Name: "TVHTML5_SIMPLY_EMBEDDED_PLAYER",
    Version: "2.0",
    Host: "www.youtube.com",
    ContextName: 85,

    RequireJSPlayer: true,
    RequireAuth: true,
    SupportsCookies: true,

    GVSPoTokenPolicies: {
        https: { Required: false },
        dash: { Required: false },
        hls: { Required: false },
    },
});

// DefaultClients returns the recommended client order for unauthenticated users
export function DefaultClients(): ClientConfig[] {
    return [AndroidSDKLess, Web, TV];
}

// DefaultWebClients returns web-based clients
export function DefaultWebClients(): ClientConfig[] {
    return [Web, WebSafari, MWeb];
}

// DefaultAndroidClients returns Android clients
export function DefaultAndroidClients(): ClientConfig[] {
    return [AndroidSDKLess, Android, AndroidVR];
}

// DefaultAuthenticatedClients returns the recommended client order for authenticated users
export function DefaultAuthenticatedClients(): ClientConfig[] {
    return [TVDowngraded, WebSafari, Web];
}

// DefaultPremiumClients returns the recommended client order for premium subscribers
export function DefaultPremiumClients(): ClientConfig[] {
    return [TVDowngraded, WebCreator, Web];
}

// GetClientByName returns a client config by name
export function GetClientByName(name: string): ClientConfig | null {
    const clients: Record<string, ClientConfig> = {
        "web": Web,
        "web_safari": WebSafari,
        "web_embedded": WebEmbedded,
        "web_music": WebMusic,
        "web_creator": WebCreator,
        "android": Android,
        "android_sdkless": AndroidSDKLess,
        "android_vr": AndroidVR,
        "ios": IOS,
        "mweb": MWeb,
        "tv": TV,
        "tv_downgraded": TVDowngraded,
        "tv_simply": TVSimply,
        "tv_embedded": TVEmbedded,
    };

    return clients[name.toLowerCase()] || null;
}
