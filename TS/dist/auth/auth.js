// Auth - Authentication support for YouTube via cookies
// Provides cookie-based authentication for accessing premium features, private videos, and age-restricted content
import { createHash } from "crypto";
import { readFileSync } from "fs";
// YouTubeURL is the base YouTube URL for cookies
export const YouTubeURL = "https://www.youtube.com";
// Auth represents authenticated YouTube session data
export class Auth {
    Cookies;
    VisitorData;
    DataSyncID;
    SessionID;
    SAPISID;
    constructor(cookies) {
        this.Cookies = cookies;
        this.VisitorData = "";
        this.DataSyncID = "";
        this.SessionID = "";
        this.SAPISID = "";
        this.extractAuthData();
    }
    // NewAuthFromFile loads cookies from a Netscape cookie file
    static FromFile(path) {
        const content = readFileSync(path, "utf-8");
        const cookies = [];
        const lines = content.split("\n");
        for (const line of lines) {
            const trimmed = line.trim();
            // Skip comments and empty lines
            if (trimmed === "" || trimmed.startsWith("#")) {
                continue;
            }
            const parts = trimmed.split("\t");
            if (parts.length < 7) {
                continue;
            }
            const domain = parts[0];
            // const hostOnly = parts[1] === "FALSE";
            const cookiePath = parts[2];
            const secure = parts[3] === "TRUE";
            // Parse expiry
            let expiryTime;
            const expiryInt = parseInt(parts[4], 10);
            if (expiryInt > 0) {
                expiryTime = new Date(expiryInt * 1000);
            }
            const name = parts[5];
            const value = parts[6];
            cookies.push({
                Name: name,
                Value: value,
                Domain: domain,
                Path: cookiePath,
                Secure: secure,
                HttpOnly: true,
                Expires: expiryTime,
            });
        }
        return new Auth(cookies);
    }
    // NewAuthFromJSON loads cookies from a JSON file (exported from browser extension)
    static FromJSON(path) {
        const content = readFileSync(path, "utf-8");
        const jsonCookies = JSON.parse(content);
        const cookies = [];
        for (const jc of jsonCookies) {
            let expiryTime;
            if (jc.expirationDate && jc.expirationDate > 0) {
                expiryTime = new Date(jc.expirationDate * 1000);
            }
            cookies.push({
                Name: jc.name,
                Value: jc.value,
                Domain: jc.domain,
                Path: jc.path,
                Secure: jc.secure,
                HttpOnly: jc.httpOnly,
                Expires: expiryTime,
            });
        }
        return new Auth(cookies);
    }
    // NewAuthFromString parses cookies from a Cookie header string
    static FromString(cookieHeader) {
        const cookies = [];
        // Parse "name=value; name2=value2" format
        const pairs = cookieHeader.split(";");
        for (const pair of pairs) {
            const trimmed = pair.trim();
            if (trimmed === "") {
                continue;
            }
            const idx = trimmed.indexOf("=");
            if (idx < 0) {
                continue;
            }
            const name = trimmed.substring(0, idx).trim();
            const value = trimmed.substring(idx + 1).trim();
            cookies.push({
                Name: name,
                Value: value,
                Domain: ".youtube.com",
                Path: "/",
                Secure: true,
                HttpOnly: true,
            });
        }
        return new Auth(cookies);
    }
    // extractAuthData extracts visitor data and other auth info from cookies
    extractAuthData() {
        for (const cookie of this.Cookies) {
            switch (cookie.Name) {
                case "VISITOR_INFO1_LIVE":
                    this.VisitorData = cookie.Value;
                    break;
                case "__Secure-3PAPISID":
                case "SAPISID":
                    this.SAPISID = cookie.Value;
                    break;
            }
        }
    }
    // IsLoggedIn returns true if the auth has valid login cookies
    IsLoggedIn() {
        for (const cookie of this.Cookies) {
            if (cookie.Name === "__Secure-3PSID" || cookie.Name === "SID") {
                return true;
            }
        }
        return false;
    }
    // GetVisitorData returns the visitor data from cookies
    GetVisitorData() {
        return this.VisitorData;
    }
    // GetDataSyncID returns the data sync ID for PO token generation
    GetDataSyncID() {
        return this.DataSyncID;
    }
    // SetDataSyncID sets the data sync ID extracted from API responses
    SetDataSyncID(dataSyncID) {
        this.DataSyncID = dataSyncID;
        // Extract session ID from data sync ID
        if (dataSyncID !== "") {
            const parts = dataSyncID.split("||");
            if (parts.length > 0 && parts[0] !== "") {
                this.SessionID = parts[0];
            }
        }
    }
    // GetSessionID returns the session ID for PO token generation
    GetSessionID() {
        return this.SessionID;
    }
    // GetSAPISIDHash generates the SAPISIDHASH authorization header
    GetSAPISIDHash(origin) {
        if (this.SAPISID === "") {
            return "";
        }
        const timestamp = Math.floor(Date.now() / 1000);
        const input = `${timestamp} ${this.SAPISID} ${origin}`;
        // SHA1 hash
        const hash = createHash("sha1").update(input).digest("hex");
        return `SAPISIDHASH ${timestamp}_${hash}`;
    }
    // GetCookieHeader returns the cookies as a Cookie header string
    GetCookieHeader() {
        const pairs = [];
        for (const cookie of this.Cookies) {
            pairs.push(`${cookie.Name}=${cookie.Value}`);
        }
        return pairs.join("; ");
    }
    // GetCookie returns a specific cookie by name
    GetCookie(name) {
        return this.Cookies.find((c) => c.Name === name);
    }
}
// ExtractVisitorDataFromHTML extracts visitor_data from YouTube HTML page
export function ExtractVisitorDataFromHTML(html) {
    const patterns = [
        /"VISITOR_DATA"\s*:\s*"([^"]+)"/,
        /ytcfg\.set\s*\(\s*\{[^}]*"VISITOR_DATA"\s*:\s*"([^"]+)"/,
    ];
    for (const pattern of patterns) {
        const match = html.match(pattern);
        if (match && match[1]) {
            return match[1];
        }
    }
    return "";
}
// ExtractDataSyncIDFromResponse extracts data_sync_id from API response
export function ExtractDataSyncIDFromResponse(data) {
    const pattern = /"dataSyncId"\s*:\s*"([^"]+)"/;
    const match = data.match(pattern);
    if (match && match[1]) {
        return match[1];
    }
    return "";
}
//# sourceMappingURL=auth.js.map