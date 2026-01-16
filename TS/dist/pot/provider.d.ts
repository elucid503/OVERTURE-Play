export declare const DefaultServerURL = "http://127.0.0.1:4416";
export interface Request {
    content_binding?: string;
    proxy?: string;
    bypass_cache?: boolean;
    source_address?: string;
    disable_tls_verification?: boolean;
    disable_innertube?: boolean;
}
export interface Response {
    poToken: string;
    contentBinding: string;
    expiresAt: string;
    error?: string;
}
export interface PingResponse {
    server_uptime: number;
    version: string;
}
export declare class Provider {
    private serverURL;
    private cache;
    private cacheTTL;
    constructor(serverURL?: string);
    IsAvailable(): Promise<boolean>;
    Ping(): Promise<PingResponse>;
    GetToken(contentBinding: string): Promise<string>;
    GetTokenWithOptions(contentBinding: string, opts?: Request): Promise<string>;
    GetGVSToken(visitorData: string, dataSyncID?: string): Promise<string>;
    private generateToken;
    ClearCache(): void;
}
//# sourceMappingURL=provider.d.ts.map