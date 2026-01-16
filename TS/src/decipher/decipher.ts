// Package decipher provides signature and n-parameter deciphering for YouTube URLs.
// This is required to bypass throttling and obtain valid stream URLs.

import { NSolver } from "./nsolver.js";

// JavaScript regex patterns for signature function extraction
const jsVarStr = `[a-zA-Z_\\$]\\w*`;
const jsSingleQuote = `'[^'\\\\]*(?:\\\\[\\s\\S][^'\\\\]*)*'`;
const jsDoubleQuote = `"[^"\\\\]*(?:\\\\[\\s\\S][^"\\\\]*)*"`;
const jsQuoteStr = `(?:${jsSingleQuote}|${jsDoubleQuote})`;
const jsKeyStr = `(?:${jsVarStr}|${jsQuoteStr})`;
const jsPropStr = `(?:\\.${jsVarStr}|\\[${jsQuoteStr}\\])`;

const reverseStr = `:function\\(a\\)\\{(?:return )?a\\.reverse\\(\\)\\}`;
const sliceStr = `:function\\(a,b\\)\\{return a\\.slice\\(b\\)\\}`;
const spliceStr = `:function\\(a,b\\)\\{a\\.splice\\(0,b\\)\\}`;
const swapStr = `:function\\(a,b\\)\\{var c=a\\[0\\];a\\[0\\]=a\\[b(?:%a\\.length)?\\];a\\[b(?:%a\\.length)?\\]=c(?:;return a)?\\}`;

const actionsObjRegex = new RegExp(
    `var (${jsVarStr})=\\{((?:(?:${jsKeyStr}${reverseStr}|${jsKeyStr}${sliceStr}|${jsKeyStr}${spliceStr}|${jsKeyStr}${swapStr}),?\\r?\\n?)+)\\};`
);

const actionsFuncRegex = new RegExp(
    `function(?: ${jsVarStr})?\\(a\\)\\{a=a\\.split\\((?:''|"")\\);\\s*((?:(?:a=)?${jsVarStr}${jsPropStr}\\(a,\\d+\\);)+)return a\\.join\\((?:''|"")\\)\\}`
);

const reverseRegex = new RegExp(`(?:^|,)(${jsKeyStr})${reverseStr}`);
const sliceRegex = new RegExp(`(?:^|,)(${jsKeyStr})${sliceStr}`);
const spliceRegex = new RegExp(`(?:^|,)(${jsKeyStr})${spliceStr}`);
const swapRegex = new RegExp(`(?:^|,)(${jsKeyStr})${swapStr}`);

// Decipherer handles signature and n-parameter challenges
export class Decipherer {
    private playerCode: string;
    private playerURL: string;
    private sigTokens: string[];
    private nSolver: NSolver | null;

    constructor(playerCode: string, playerURL: string = "") {
        this.playerCode = playerCode;
        this.playerURL = playerURL;
        this.sigTokens = [];
        this.nSolver = null;
    }

    // Initialize extracts signature tokens and creates n-solver
    async Initialize(): Promise<void> {
        this.extractSignatureTokens();

        // Initialize n-solver
        try {
            this.nSolver = new NSolver(this.playerCode);
        } catch {
            // N-solver is optional, some videos don't need it
            this.nSolver = null;
        }
    }

    // GetSignatureTimestamp returns the signature timestamp from the player code
    GetSignatureTimestamp(): number {
        return GetSignatureTimestamp(this.playerCode);
    }

    // SolveNChallenge solves the n-parameter challenge using the JS runtime
    SolveNChallenge(n: string): string {
        if (!this.nSolver) {
            return n;
        }
        return this.nSolver.Solve(n);
    }

    // DecipherURL deciphers a stream URL by solving signature and n-parameter challenges
    DecipherURL(streamURL: string): string {
        const parsed = new URL(streamURL);
        const query = parsed.searchParams;

        // Check for signature cipher in query
        const sig = query.get("s");
        if (sig) {
            // Decipher signature
            const decipheredSig = this.decipherSignature(sig);

            // Get signature parameter name (usually "sig" or "signature")
            let sp = query.get("sp");
            if (!sp) {
                sp = "signature";
            }

            query.set(sp, decipheredSig);
            query.delete("s");
            query.delete("sp");
        }

        // Handle n-parameter for throttle bypass
        const n = query.get("n");
        if (n) {
            const newN = this.SolveNChallenge(n);
            if (newN) {
                query.set("n", newN);
            }
        }

        return parsed.toString();
    }

    // DecipherSignature deciphers a signature using the extracted tokens
    DecipherSignature(sig: string): string {
        return this.decipherSignature(sig);
    }

    // decipherSignature applies the signature transformation
    private decipherSignature(sig: string): string {
        let arr = sig.split("");

        for (const token of this.sigTokens) {
            if (token.length < 1) {
                continue;
            }

            switch (token[0]) {
                case "r":
                    // Reverse
                    arr = arr.reverse();
                    break;

                case "s": {
                    // Slice
                    const pos = parseInt(token.slice(1), 10);
                    if (!isNaN(pos) && pos < arr.length) {
                        arr = arr.slice(pos);
                    }
                    break;
                }

                case "p": {
                    // Splice
                    const pos = parseInt(token.slice(1), 10);
                    if (!isNaN(pos) && pos < arr.length) {
                        arr = arr.slice(pos);
                    }
                    break;
                }

                case "w": {
                    // Swap
                    const pos = parseInt(token.slice(1), 10);
                    if (!isNaN(pos)) {
                        const swapPos = pos % arr.length;
                        [arr[0], arr[swapPos]] = [arr[swapPos], arr[0]];
                    }
                    break;
                }
            }
        }

        return arr.join("");
    }

    // extractSignatureTokens extracts signature transformation tokens from player code
    private extractSignatureTokens(): void {
        const objects = this.playerCode.match(actionsObjRegex);
        const functions = this.playerCode.match(actionsFuncRegex);

        if (!objects || objects.length < 3 || !functions || functions.length < 2) {
            // Try alternative extraction methods
            this.extractSignatureTokensAlt();
            return;
        }

        const obj = objects[1].replace(/\$/g, "\\$");
        const objBody = objects[2].replace(/\$/g, "\\$");
        const funcBody = functions[1].replace(/\$/g, "\\$");

        const reverseKey = extractKey(reverseRegex, objBody);
        const sliceKey = extractKey(sliceRegex, objBody);
        const spliceKey = extractKey(spliceRegex, objBody);
        const swapKey = extractKey(swapRegex, objBody);

        const keys = `(${reverseKey}|${sliceKey}|${spliceKey}|${swapKey})`;
        const tokenizeRegex = new RegExp(
            `(?:a=)?${obj}(?:\\.${keys}|\\[(?:'${keys}'|"${keys}")\\])\\(a,(\\d+)\\)`,
            "g"
        );

        let match;
        while ((match = tokenizeRegex.exec(funcBody)) !== null) {
            if (match.length < 5) {
                continue;
            }

            let key = match[1];
            if (!key) {
                key = match[2];
            }
            if (!key) {
                key = match[3];
            }

            switch (key) {
                case reverseKey:
                    this.sigTokens.push("r");
                    break;
                case sliceKey:
                    this.sigTokens.push("s" + match[4]);
                    break;
                case spliceKey:
                    this.sigTokens.push("p" + match[4]);
                    break;
                case swapKey:
                    this.sigTokens.push("w" + match[4]);
                    break;
            }
        }
    }

    // extractSignatureTokensAlt tries alternative patterns for signature extraction
    private extractSignatureTokensAlt(): void {
        // Alternative signature function pattern
        const altPattern = /\b[cs]\s*&&\s*[adf]\.set\([^,]+\s*,\s*encodeURIComponent\(([a-zA-Z0-9$]+)\(/;
        const match = this.playerCode.match(altPattern);

        if (!match || match.length < 2) {
            // If no signature function found, the URL may not need deciphering
            return;
        }

        // Further extraction would go here
    }
}

// extractKey extracts a key from the object body using the given regex
function extractKey(re: RegExp, body: string): string {
    const match = body.match(re);
    if (!match || match.length < 2) {
        return "";
    }

    let key = match[1];
    // Remove quotes if present
    key = key.replace(/^["']|["']$/g, "");
    return key;
}

// GetSignatureTimestamp extracts the signature timestamp from player code
export function GetSignatureTimestamp(playerCode: string): number {
    const patterns = [/(?:signatureTimestamp|sts)\s*:\s*(\d{5})/, /"STS"\s*:\s*(\d{5})/];

    for (const pattern of patterns) {
        const match = playerCode.match(pattern);
        if (match && match[1]) {
            const sts = parseInt(match[1], 10);
            if (!isNaN(sts)) {
                return sts;
            }
        }
    }

    return 0;
}

// ExtractPlayerID extracts the player ID from a player URL
export function ExtractPlayerID(playerURL: string): string {
    const patterns = [/\/s\/player\/([a-zA-Z0-9_-]{8,})\//, /\/([a-zA-Z0-9_-]{8,})\/player/, /\b(vfl[a-zA-Z0-9_-]+)\b/];

    for (const pattern of patterns) {
        const match = playerURL.match(pattern);
        if (match && match[1]) {
            return match[1];
        }
    }

    return "";
}

// NewDecipherer creates a new Decipherer from player code
export async function NewDecipherer(playerCode: string, playerURL: string = ""): Promise<Decipherer> {
    const d = new Decipherer(playerCode, playerURL);
    await d.Initialize();
    return d;
}
