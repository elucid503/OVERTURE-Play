// Client - Main YouTube client for fetching video information

import { Video, CreateEmptyVideo } from "../types/video.js";
import { Thumbnail } from "../types/thumbnail.js";
import { Format, CreateEmptyFormat } from "../types/format.js";
import { Range } from "../types/range.js";
import { ClientConfig, GetContext, GetContextWithVisitor } from "../innertube/config.js";
import { DefaultClients, DefaultAuthenticatedClients } from "../innertube/clients.js";
import { Provider as POTProvider } from "../pot/provider.js";
import { Decipherer, NewDecipherer, ExtractPlayerID } from "../decipher/decipher.js";
import { Auth, ExtractDataSyncIDFromResponse, ExtractVisitorDataFromHTML } from "../auth/auth.js";
import {
    PlayerResponse,
    ThumbnailContainer,
    StreamingFormat,
    RangeData,
} from "./response.js";

// ClientOptions configures the YouTube client
export interface ClientOptions {
    POTServerURL?: string;
    Clients?: ClientConfig[];
    UserAgent?: string;
    AcceptLang?: string;
    Debug?: boolean;

    // Authentication options
    Auth?: Auth;              // Pre-configured auth
    CookieFile?: string;      // Path to Netscape cookie file
    CookieJSONFile?: string;  // Path to JSON cookie file
    CookieString?: string;    // Cookie header string
}

// Client is the main YouTube client for fetching video information
export class Client {
    private POTProvider: POTProvider | null;
    private Decipherer: Decipherer | null;
    private Auth: Auth | null;

    private Clients: ClientConfig[];
    private PlayerURL: string;
    private PlayerID: string;
    private PlayerCode: string;
    private VisitorData: string;

    private UserAgent: string;
    private AcceptLang: string;
    private Debug: boolean;

    constructor(opts?: ClientOptions) {
        this.POTProvider = new POTProvider(opts?.POTServerURL || "http://127.0.0.1:4416");
        this.Decipherer = null;
        this.Auth = null;
        this.VisitorData = "";

        this.Clients = opts?.Clients || DefaultClients();
        this.PlayerURL = "";
        this.PlayerID = "";
        this.PlayerCode = "";

        this.UserAgent = opts?.UserAgent || "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36";
        this.AcceptLang = opts?.AcceptLang || "en-US,en;q=0.9";
        this.Debug = opts?.Debug || false;

        // Set up authentication
        if (opts?.Auth) {
            this.Auth = opts.Auth;
        } else if (opts?.CookieFile) {
            try {
                this.Auth = Auth.FromFile(opts.CookieFile);
            } catch (err) {
                if (this.Debug) {
                    console.log("Failed to load cookie file:", err);
                }
            }
        } else if (opts?.CookieJSONFile) {
            try {
                this.Auth = Auth.FromJSON(opts.CookieJSONFile);
            } catch (err) {
                if (this.Debug) {
                    console.log("Failed to load JSON cookie file:", err);
                }
            }
        } else if (opts?.CookieString) {
            this.Auth = Auth.FromString(opts.CookieString);
        }

        // Use authenticated clients if logged in
        if (this.Auth?.IsLoggedIn() && !opts?.Clients) {
            this.Clients = DefaultAuthenticatedClients();
        }
    }

    // GetVideo fetches video information and formats
    async GetVideo(videoID: string): Promise<Video> {
        videoID = this.extractVideoID(videoID);
        if (!videoID) {
            throw new Error("invalid video ID or URL");
        }

        // Fetch player info first
        await this.ensurePlayer();

        // Try each client until one works
        let lastErr: Error | null = null;
        for (const clientConfig of this.Clients) {
            try {
                const video = await this.fetchWithClient(videoID, clientConfig);
                return video;
            } catch (err) {
                lastErr = err instanceof Error ? err : new Error(String(err));
                if (this.Debug) {
                    console.log(`Client ${clientConfig.Name} failed:`, err);
                }
            }
        }

        throw new Error(`all clients failed, last error: ${lastErr?.message}`);
    }

    // extractVideoID extracts the video ID from a URL or returns as-is if already an ID
    private extractVideoID(input: string): string {
        input = input.trim();

        // Already an ID (11 characters)
        if (input.length === 11 && /^[a-zA-Z0-9_-]{11}$/.test(input)) {
            return input;
        }

        // Parse as URL
        const patterns = [
            /(?:youtube\.com\/watch\?v=|youtu\.be\/)([a-zA-Z0-9_-]{11})/,
            /youtube\.com\/embed\/([a-zA-Z0-9_-]{11})/,
            /youtube\.com\/v\/([a-zA-Z0-9_-]{11})/,
            /youtube\.com\/shorts\/([a-zA-Z0-9_-]{11})/,
        ];

        for (const pattern of patterns) {
            const match = input.match(pattern);
            if (match && match[1]) {
                return match[1];
            }
        }

        return "";
    }

    // ensurePlayer fetches and caches the player script
    private async ensurePlayer(): Promise<void> {
        if (this.Decipherer) {
            return;
        }

        // Fetch player URL from YouTube page
        const playerURL = await this.fetchPlayerURL();
        this.PlayerURL = playerURL;
        this.PlayerID = ExtractPlayerID(playerURL);

        // Fetch player code
        const playerCode = await this.fetchPlayerCode(playerURL);
        this.PlayerCode = playerCode;

        // Create decipherer
        this.Decipherer = await NewDecipherer(playerCode, playerURL);
    }

    // fetchPlayerURL gets the current player URL from YouTube
    private async fetchPlayerURL(): Promise<string> {
        const resp = await fetch("https://www.youtube.com/iframe_api", {
            headers: this.getRequestHeaders(),
        });

        const body = await resp.text();

        // Extract player ID from iframe_api - format: \/s\/player\/XXXXXXXX\/ (escaped) or /s/player/XXXXXXXX/
        const re = /\\?\/s\\?\/player\\?\/([0-9a-fA-F]{8})\\?\//;
        const match = body.match(re);
        if (match && match[1]) {
            const playerID = match[1];
            return `https://www.youtube.com/s/player/${playerID}/player_ias.vflset/en_US/base.js`;
        }

        // Fallback: fetch from YouTube page
        return this.fetchPlayerURLFromPage();
    }

    // fetchPlayerURLFromPage extracts player URL from main YouTube page
    private async fetchPlayerURLFromPage(): Promise<string> {
        // Try a video watch page first - more reliable for extracting player
        const resp = await fetch("https://www.youtube.com/watch?v=dQw4w9WgXcQ", {
            headers: this.getRequestHeaders(),
        });

        const body = await resp.text();

        // Extract visitor_data from the page if we don't have it yet
        if (!this.VisitorData) {
            this.VisitorData = ExtractVisitorDataFromHTML(body);
        }

        // Look for player script URL with multiple patterns
        const patterns = [
            // JSON format in ytInitialPlayerResponse
            /"jsUrl"\s*:\s*"(\/s\/player\/[^"]+\/player_ias\.vflset\/[^"]+\/base\.js)"/,
            // PLAYER_JS_URL format
            /"PLAYER_JS_URL"\s*:\s*"(\/s\/player\/[^"]+base\.js)"/,
            // Script tag src
            /<script[^>]+src="(\/s\/player\/[^"]+\/base\.js)"/,
            // Raw URL pattern - capture full path
            /(\/s\/player\/[a-zA-Z0-9_-]+\/player_ias\.vflset\/[a-zA-Z_]+\/base\.js)/,
            // Alternative format with hash
            /\/s\/player\/([a-fA-F0-9]{8})\//,
        ];

        for (let i = 0; i < patterns.length; i++) {
            const match = body.match(patterns[i]);
            if (match && match[1]) {
                let playerPath = match[1];
                // Last pattern returns just the hash, need to construct URL
                if (i === patterns.length - 1 && !playerPath.includes("/")) {
                    playerPath = `/s/player/${playerPath}/player_ias.vflset/en_US/base.js`;
                }
                if (!playerPath.startsWith("http")) {
                    playerPath = "https://www.youtube.com" + playerPath;
                }
                return playerPath;
            }
        }

        // Final fallback: try embed page
        return this.fetchPlayerURLFromEmbed();
    }

    // fetchPlayerURLFromEmbed extracts player URL from embed page
    private async fetchPlayerURLFromEmbed(): Promise<string> {
        const resp = await fetch("https://www.youtube.com/embed/dQw4w9WgXcQ", {
            headers: this.getRequestHeaders(),
        });

        const body = await resp.text();

        // Embed page uses simpler patterns
        const patterns = [
            /"jsUrl"\s*:\s*"([^"]+base\.js)"/,
            /"PLAYER_JS_URL"\s*:\s*"([^"]+)"/,
            /\/s\/player\/([a-fA-F0-9]{8})\//,
        ];

        for (let i = 0; i < patterns.length; i++) {
            const match = body.match(patterns[i]);
            if (match && match[1]) {
                let playerPath = match[1];
                if (i === patterns.length - 1 && !playerPath.includes("/")) {
                    playerPath = `/s/player/${playerPath}/player_ias.vflset/en_US/base.js`;
                }
                if (!playerPath.startsWith("http")) {
                    playerPath = "https://www.youtube.com" + playerPath;
                }
                return playerPath;
            }
        }

        throw new Error("player URL not found");
    }

    // fetchPlayerCode downloads the player JavaScript code
    private async fetchPlayerCode(playerURL: string): Promise<string> {
        const resp = await fetch(playerURL, {
            headers: this.getRequestHeaders(),
        });

        return resp.text();
    }

    // fetchWithClient fetches video info using a specific innertube client
    private async fetchWithClient(videoID: string, clientConfig: ClientConfig): Promise<Video> {
        // Get context with visitor data if available
        const visitorData = this.getVisitorData();
        const ctx = visitorData
            ? GetContextWithVisitor(clientConfig, visitorData)
            : GetContext(clientConfig);

        let sts = 0;
        if (this.Decipherer) {
            sts = this.Decipherer.GetSignatureTimestamp();
        }

        const payload: Record<string, unknown> = {
            context: ctx,
            videoId: videoID,
            playbackContext: {
                contentPlaybackContext: {
                    signatureTimestamp: sts,
                    html5Preference: "HTML5_PREF_WANTS",
                },
            },
            racyCheckOk: true,
            contentCheckOk: true,
        };

        // Add Player PO token if required for this client (bound to video ID)
        const playerPOToken = await this.getPlayerPOToken(videoID, clientConfig);
        if (playerPOToken) {
            payload.serviceIntegrityDimensions = {
                poToken: playerPOToken,
            };
        }

        // Make player API request - no API key needed for modern clients
        const apiURL = "https://www.youtube.com/youtubei/v1/player?prettyPrint=false";

        const resp = await fetch(apiURL, {
            method: "POST",
            headers: this.getAPIRequestHeaders(clientConfig),
            body: JSON.stringify(payload),
        });

        const body = await resp.text();

        // Parse response
        return this.parsePlayerResponse(body, clientConfig, videoID);
    }

    // getPlayerPOToken gets a PO token for the player API request (bound to video ID)
    private async getPlayerPOToken(videoID: string, clientConfig: ClientConfig): Promise<string> {
        // Check if client requires PO token for player
        if (!this.requiresPoToken(clientConfig)) {
            return "";
        }

        // Try to get PO token from provider
        if (!this.POTProvider) {
            return "";
        }

        try {
            const available = await this.POTProvider.IsAvailable();
            if (!available) {
                return "";
            }

            // Player PO token is bound to video ID
            return await this.POTProvider.GetToken(videoID);
        } catch {
            return "";
        }
    }

    // getGVSPOToken gets a GVS PO token for stream URLs (bound to visitor_data or data_sync_id)
    private async getGVSPOToken(videoID: string, clientConfig: ClientConfig): Promise<string> {
        // Check if client has GVS PO token policies
        if (!clientConfig.GVSPoTokenPolicies || Object.keys(clientConfig.GVSPoTokenPolicies).length === 0) {
            return "";
        }

        // Try to get PO token from provider
        if (!this.POTProvider) {
            return "";
        }

        try {
            const available = await this.POTProvider.IsAvailable();
            if (!available) {
                return "";
            }

            // GVS PO token is bound to visitor_data (unauthenticated) or data_sync_id (authenticated)
            const visitorData = this.getVisitorData();
            let dataSyncID = "";
            if (this.Auth?.IsLoggedIn()) {
                dataSyncID = this.Auth.GetDataSyncID();
            }

            return await this.POTProvider.GetGVSToken(visitorData, dataSyncID);
        } catch {
            return "";
        }
    }

    // requiresPoToken checks if a client requires PO token
    private requiresPoToken(clientConfig: ClientConfig): boolean {
        for (const protocol of Object.keys(clientConfig.GVSPoTokenPolicies)) {
            const policy = clientConfig.GVSPoTokenPolicies[protocol as keyof typeof clientConfig.GVSPoTokenPolicies];
            if (policy && policy.Required) {
                return true;
            }
        }
        return clientConfig.PlayerPoTokenPolicy?.Required || false;
    }

    // parsePlayerResponse parses the player API response
    private async parsePlayerResponse(data: string, clientConfig: ClientConfig, videoID: string): Promise<Video> {
        let resp: PlayerResponse;
        try {
            resp = JSON.parse(data);
        } catch (err) {
            throw new Error(`failed to parse response: ${err}`);
        }

        // Check for playability errors
        if (resp.playabilityStatus.status !== "OK") {
            throw new Error(
                `video not playable: ${resp.playabilityStatus.status} - ${resp.playabilityStatus.reason || "unknown reason"}`
            );
        }

        // Get GVS PO token for stream URLs (bound to visitor_data or data_sync_id)
        const gvsPOToken = await this.getGVSPOToken(videoID, clientConfig);

        const video: Video = {
            ...CreateEmptyVideo(),
            ID: resp.videoDetails?.videoId || "",
            Title: resp.videoDetails?.title || "",
            Description: resp.videoDetails?.shortDescription || "",
            Author: resp.videoDetails?.author || "",
            ChannelID: resp.videoDetails?.channelId || "",
            Duration: this.parseDuration(resp.videoDetails?.lengthSeconds || "0"),
            ViewCount: this.parseInt(resp.videoDetails?.viewCount || "0"),
            IsLive: resp.videoDetails?.isLiveContent || false,
            IsPrivate: resp.videoDetails?.isPrivate || false,
            Formats: [],
            Thumbnails: this.parseThumbnails(resp.videoDetails?.thumbnail),
            IsPlayable: true,
            DASHManifestURL: resp.streamingData?.dashManifestUrl || "",
            HLSManifestURL: resp.streamingData?.hlsManifestUrl || "",
            PlayabilityStatus: {
                Status: resp.playabilityStatus.status,
                Reason: resp.playabilityStatus.reason || "",
                PlayableInEmbed: resp.playabilityStatus.playableInEmbed || false,
                IsAgeRestricted: false,
                MiniplayerStatus: "",
            },
        };

        // Parse formats
        const allFormats = [
            ...(resp.streamingData?.formats || []),
            ...(resp.streamingData?.adaptiveFormats || []),
        ];

        for (const sf of allFormats) {
            try {
                const format = this.parseFormat(sf, gvsPOToken);
                video.Formats.push(format);
            } catch (err) {
                if (this.Debug) {
                    console.log(`Failed to parse format ${sf.itag}:`, err);
                }
            }
        }

        return video;
    }

    // parseFormat parses a streaming format and deciphers URLs
    private parseFormat(sf: StreamingFormat, gvsPOToken: string): Format {
        const format: Format = {
            ...CreateEmptyFormat(),
            ITag: sf.itag,
            MimeType: sf.mimeType,
            Bitrate: sf.bitrate,
            AverageBitrate: sf.averageBitrate || 0,
            ContentLength: this.parseInt(sf.contentLength || "0"),
            Width: sf.width || 0,
            Height: sf.height || 0,
            FPS: sf.fps || 0,
            Quality: sf.quality,
            QualityLabel: sf.qualityLabel || "",
            AudioQuality: sf.audioQuality || "",
            AudioChannels: sf.audioChannels || 0,
            AudioSampleRate: this.parseInt(sf.audioSampleRate || "0"),
            IndexRange: this.parseRange(sf.indexRange),
            InitRange: this.parseRange(sf.initRange),
        };

        // Parse codec information from mimeType
        const codecMatch = sf.mimeType.match(/codecs="([^"]+)"/);
        if (codecMatch && codecMatch[1]) {
            format.Codec = codecMatch[1];
            // Parse video and audio codecs
            const codecs = codecMatch[1].split(",").map((c) => c.trim());
            if (codecs.length > 0) {
                if (format.Width > 0 || format.Height > 0) {
                    format.VideoCodec = codecs[0];
                    format.AudioCodec = codecs[1] || "";
                } else {
                    format.AudioCodec = codecs[0];
                }
            }
        }

        // Get URL
        let streamURL: string;
        if (sf.url) {
            streamURL = sf.url;
        } else if (sf.signatureCipher) {
            // Decipher the URL
            streamURL = this.decipherURL(sf.signatureCipher);
        } else {
            throw new Error(`no URL available for format ${sf.itag}`);
        }

        // Process n-parameter
        streamURL = this.processNParameter(streamURL);

        // Add GVS PO token if available
        if (gvsPOToken) {
            streamURL = this.addPOTokenToURL(streamURL, gvsPOToken);
        }

        format.URL = streamURL;
        return format;
    }

    // addPOTokenToURL adds a PO token to the stream URL
    private addPOTokenToURL(streamURL: string, poToken: string): string {
        try {
            const parsedURL = new URL(streamURL);
            parsedURL.searchParams.set("pot", poToken);
            return parsedURL.toString();
        } catch {
            return streamURL;
        }
    }

    // decipherURL deciphers a signature cipher
    private decipherURL(signatureCipher: string): string {
        const params = new URLSearchParams(signatureCipher);

        let streamURL = params.get("url") || "";
        const signature = params.get("s") || "";
        let signatureParam = params.get("sp") || "";
        if (!signatureParam) {
            signatureParam = "sig";
        }

        if (signature && this.Decipherer) {
            const deciphered = this.Decipherer.DecipherSignature(signature);

            const parsedURL = new URL(streamURL);
            parsedURL.searchParams.set(signatureParam, deciphered);
            streamURL = parsedURL.toString();
        }

        return streamURL;
    }

    // processNParameter processes and solves the n-parameter challenge
    private processNParameter(streamURL: string): string {
        if (!this.Decipherer) {
            return streamURL;
        }

        try {
            const parsedURL = new URL(streamURL);
            const n = parsedURL.searchParams.get("n");
            if (!n) {
                return streamURL;
            }

            const solved = this.Decipherer.SolveNChallenge(n);
            if (solved === n) {
                return streamURL;
            }

            parsedURL.searchParams.set("n", solved);
            return parsedURL.toString();
        } catch {
            return streamURL;
        }
    }

    // parseThumbnails parses thumbnail data
    private parseThumbnails(data?: ThumbnailContainer): Thumbnail[] {
        if (!data || !data.thumbnails) {
            return [];
        }
        return data.thumbnails.map((t) => ({
            URL: t.url,
            Width: t.width,
            Height: t.height,
        }));
    }

    // parseRange parses a range object
    private parseRange(r?: RangeData): Range | null {
        if (!r) {
            return null;
        }
        return {
            Start: this.parseInt(r.start),
            End: this.parseInt(r.end),
        };
    }

    // parseDuration parses duration string to seconds
    private parseDuration(s: string): number {
        return this.parseInt(s);
    }

    // parseInt parses an integer string
    private parseInt(s: string): number {
        const num = parseInt(s, 10);
        return isNaN(num) ? 0 : num;
    }

    // getRequestHeaders returns standard request headers
    private getRequestHeaders(): Record<string, string> {
        const headers: Record<string, string> = {
            "User-Agent": this.UserAgent,
            "Accept-Language": this.AcceptLang,
            Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
        };

        // Add cookie header if auth is configured
        if (this.Auth) {
            headers["Cookie"] = this.Auth.GetCookieHeader();
        }

        return headers;
    }

    // getAPIRequestHeaders returns headers for API requests
    private getAPIRequestHeaders(clientConfig: ClientConfig): Record<string, string> {
        const headers: Record<string, string> = {
            "Content-Type": "application/json",
            "X-YouTube-Client-Name": String(clientConfig.ContextName),
            "X-YouTube-Client-Version": clientConfig.Version,
            Origin: "https://www.youtube.com",
            Referer: "https://www.youtube.com/",
            "User-Agent": clientConfig.UserAgent || this.UserAgent,
        };

        // Add visitor ID header
        const visitorData = this.getVisitorData();
        if (visitorData) {
            headers["X-Goog-Visitor-Id"] = visitorData;
        }

        // Add authentication headers
        if (this.Auth) {
            headers["Cookie"] = this.Auth.GetCookieHeader();

            // Add SAPISIDHASH for authenticated requests
            if (this.Auth.IsLoggedIn()) {
                const sapisidhash = this.Auth.GetSAPISIDHash("https://www.youtube.com");
                if (sapisidhash) {
                    headers["Authorization"] = sapisidhash;
                    headers["X-Origin"] = "https://www.youtube.com";
                }
            }
        }

        return headers;
    }

    // getVisitorData returns the visitor data for API requests
    private getVisitorData(): string {
        if (this.VisitorData) {
            return this.VisitorData;
        }
        if (this.Auth) {
            return this.Auth.GetVisitorData();
        }
        return "";
    }

    // SetVisitorData sets the visitor data for API requests
    SetVisitorData(visitorData: string): void {
        this.VisitorData = visitorData;
    }

    // IsAuthenticated returns true if the client has valid authentication
    IsAuthenticated(): boolean {
        return this.Auth !== null && this.Auth.IsLoggedIn();
    }

    // GetAuth returns the current Auth object
    GetAuth(): Auth | null {
        return this.Auth;
    }
}

// NewClient creates a new YouTube client with default configuration
export function NewClient(opts?: ClientOptions): Client {
    return new Client(opts);
}
