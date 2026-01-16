export declare const YouTubeURL = "https://www.youtube.com";
export interface Cookie {
    Name: string;
    Value: string;
    Domain: string;
    Path: string;
    Secure: boolean;
    HttpOnly: boolean;
    Expires?: Date;
}
export declare class Auth {
    Cookies: Cookie[];
    VisitorData: string;
    DataSyncID: string;
    SessionID: string;
    SAPISID: string;
    constructor(cookies: Cookie[]);
    static FromFile(path: string): Auth;
    static FromJSON(path: string): Auth;
    static FromString(cookieHeader: string): Auth;
    private extractAuthData;
    IsLoggedIn(): boolean;
    GetVisitorData(): string;
    GetDataSyncID(): string;
    SetDataSyncID(dataSyncID: string): void;
    GetSessionID(): string;
    GetSAPISIDHash(origin: string): string;
    GetCookieHeader(): string;
    GetCookie(name: string): Cookie | undefined;
}
export declare function ExtractVisitorDataFromHTML(html: string): string;
export declare function ExtractDataSyncIDFromResponse(data: string): string;
//# sourceMappingURL=auth.d.ts.map