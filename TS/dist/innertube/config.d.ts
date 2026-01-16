import { StreamingProtocol, PoTokenPolicy } from "../types/streaming.js";
export type GVSPolicies = {
    [StreamingProtocol.HTTPS]: PoTokenPolicy;
    [StreamingProtocol.DASH]: PoTokenPolicy;
    [StreamingProtocol.HLS]: PoTokenPolicy;
};
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
    GVSPoTokenPolicies: GVSPolicies;
    PlayerPoTokenPolicy: PoTokenPolicy;
    SubsPoTokenPolicy: PoTokenPolicy;
    POTokenPolicy: PoTokenPolicy;
}
export interface InnertubeContext {
    client: ClientInfo;
}
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
export interface ThirdPartyContext {
    embedUrl: string;
}
export declare function GetContext(c: ClientConfig): InnertubeContext;
export declare function GetContextWithVisitor(c: ClientConfig, visitorData: string): InnertubeContext;
export declare function RequiresPoToken(c: ClientConfig): boolean;
export declare function CreateClientConfig(partial: Partial<ClientConfig>): ClientConfig;
//# sourceMappingURL=config.d.ts.map