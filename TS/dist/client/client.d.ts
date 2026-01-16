import { Video } from "../types/video.js";
import { ClientConfig } from "../innertube/config.js";
import { Auth } from "../auth/auth.js";
export interface ClientOptions {
    POTServerURL?: string;
    Clients?: ClientConfig[];
    UserAgent?: string;
    AcceptLang?: string;
    Debug?: boolean;
    Auth?: Auth;
    CookieFile?: string;
    CookieJSONFile?: string;
    CookieString?: string;
}
export declare class Client {
    private POTProvider;
    private Decipherer;
    private Auth;
    private Clients;
    private PlayerURL;
    private PlayerID;
    private PlayerCode;
    private VisitorData;
    private UserAgent;
    private AcceptLang;
    private Debug;
    constructor(opts?: ClientOptions);
    GetVideo(videoID: string): Promise<Video>;
    private extractVideoID;
    private ensurePlayer;
    private fetchPlayerURL;
    private fetchPlayerURLFromPage;
    private fetchPlayerURLFromEmbed;
    private fetchPlayerCode;
    private fetchWithClient;
    private getPlayerPOToken;
    private getGVSPOToken;
    private requiresPoToken;
    private parsePlayerResponse;
    private parseFormat;
    private addPOTokenToURL;
    private decipherURL;
    private processNParameter;
    private parseThumbnails;
    private parseRange;
    private parseDuration;
    private parseInt;
    private getRequestHeaders;
    private getAPIRequestHeaders;
    private getVisitorData;
    SetVisitorData(visitorData: string): void;
    IsAuthenticated(): boolean;
    GetAuth(): Auth | null;
}
export declare function NewClient(opts?: ClientOptions): Client;
//# sourceMappingURL=client.d.ts.map