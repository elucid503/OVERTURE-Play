// Package pot provides PO (Proof of Origin) token generation for YouTube.
// PO tokens are required by YouTube to prevent 403 errors on video streams.

// DefaultServerURL is the default bgutil HTTP server URL
// Run: docker run -d -p 4416:4416 brainicism/bgutil-ytdlp-pot-provider
export const DefaultServerURL = "http://127.0.0.1:4416";

// Request represents a request to the bgutil server
export interface Request {
    content_binding?: string;
    proxy?: string;
    bypass_cache?: boolean;
    source_address?: string;
    disable_tls_verification?: boolean;
    disable_innertube?: boolean;
}

// Response represents the response from bgutil
export interface Response {
    poToken: string;
    contentBinding: string;
    expiresAt: string;
    error?: string;
}

// PingResponse represents the response from bgutil /ping endpoint
export interface PingResponse {
    server_uptime: number;
    version: string;
}

// CachedToken stores a token with its expiration
interface CachedToken {
    Token: string;
    ExpiresAt: Date;
}

// Provider generates PO tokens using a bgutil HTTP server
export class Provider {
    private serverURL: string;
    private cache: Map<string, CachedToken>;
    private cacheTTL: number; // milliseconds

    constructor(serverURL?: string) {
        this.serverURL = serverURL || DefaultServerURL;
        this.cache = new Map();
        this.cacheTTL = 5 * 60 * 60 * 1000; // 5 hours
    }

    // IsAvailable checks if the bgutil server is reachable
    async IsAvailable(): Promise<boolean> {
        try {
            await this.Ping();
            return true;
        } catch {
            return false;
        }
    }

    // Ping checks if the bgutil server is running
    async Ping(): Promise<PingResponse> {
        const resp = await fetch(`${this.serverURL}/ping`);

        if (!resp.ok) {
            throw new Error(`bgutil server returned status ${resp.status}`);
        }

        const pingResp = await resp.json() as PingResponse;
        return pingResp;
    }

    // GetToken fetches a PO token for the given content binding
    // Content binding is typically visitor_data for logged-out users
    // or the session ID (first part of data_sync_id) for logged-in users
    async GetToken(contentBinding: string): Promise<string> {
        return this.GetTokenWithOptions(contentBinding, undefined);
    }

    // GetTokenWithOptions fetches a PO token with custom options
    async GetTokenWithOptions(contentBinding: string, opts?: Request): Promise<string> {
        // Check cache first
        const cached = this.cache.get(contentBinding);
        if (cached && new Date() < cached.ExpiresAt) {
            return cached.Token;
        }

        // Generate new token
        const { token, expiresAt } = await this.generateToken(contentBinding, opts);

        // Cache the token
        this.cache.set(contentBinding, {
            Token: token,
            ExpiresAt: expiresAt,
        });

        return token;
    }

    // GetGVSToken generates a GVS context PO token for video streaming
    // Use visitor_data for logged-out users or data_sync_id for logged-in users
    async GetGVSToken(visitorData: string, dataSyncID?: string): Promise<string> {
        let contentBinding = visitorData;

        // If logged in, use session ID from DataSyncID
        if (dataSyncID) {
            contentBinding = extractSessionID(dataSyncID);
        }

        return this.GetToken(contentBinding);
    }

    // generateToken makes the actual HTTP request to the bgutil server
    private async generateToken(contentBinding: string, opts?: Request): Promise<{ token: string; expiresAt: Date }> {
        const reqBody: Request = {
            ...opts,
            content_binding: contentBinding,
        };

        const resp = await fetch(`${this.serverURL}/get_pot`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(reqBody),
        });

        const body = await resp.text();
        
        let bgResp: Response;
        try {
            bgResp = JSON.parse(body);
        } catch {
            throw new Error(`failed to decode response: ${body}`);
        }

        if (bgResp.error) {
            throw new Error(`bgutil error: ${bgResp.error}`);
        }

        if (!bgResp.poToken) {
            throw new Error("bgutil returned empty token");
        }

        // Use expiry from response if available, otherwise use cache TTL
        let expiresAt: Date;
        if (bgResp.expiresAt) {
            expiresAt = new Date(bgResp.expiresAt);
        } else {
            expiresAt = new Date(Date.now() + this.cacheTTL);
        }

        return { token: bgResp.poToken, expiresAt };
    }

    // ClearCache clears the token cache
    ClearCache(): void {
        this.cache.clear();
    }
}

// extractSessionID extracts the session ID from a DataSyncID
// DataSyncID format is "SESSION_ID||..." - we need only the first part
function extractSessionID(dataSyncID: string): string {
    if (!dataSyncID) {
        return "";
    }

    const parts = dataSyncID.split("||");
    if (parts.length > 0 && parts[0]) {
        return parts[0];
    }

    return dataSyncID;
}
