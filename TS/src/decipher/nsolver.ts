// Package decipher provides n-parameter solving for YouTube URLs.
// This is required to bypass throttling and obtain valid stream URLs.
// Note: The n-parameter solver may not work for all videos due to YouTube's obfuscation.
// When it fails, downloads may be throttled but will still work.

// NSolver handles n-parameter solving using JavaScript evaluation
export class NSolver {
    private playerCode: string;
    private nFuncCode: string;
    private extractionFailed: boolean;

    constructor(playerCode: string) {
        this.playerCode = playerCode;
        this.nFuncCode = "";
        this.extractionFailed = false;
        this.extractNFunction();
    }

    // Solve solves the n-parameter challenge
    Solve(n: string): string {
        if (!this.nFuncCode || this.extractionFailed) {
            return n;
        }

        try {
            // Create a function from the extracted code and execute it
            const func = new Function(`
                "use strict";
                ${this.nFuncCode}
                return nFunction("${n.replace(/"/g, '\\"')}");
            `);
            const result = func();
            return result !== undefined && result !== null ? String(result) : n;
        } catch (err) {
            // Mark as failed so we don't spam the console with errors
            this.extractionFailed = true;
            // This is expected - YouTube's n-function often has external dependencies
            // Downloads will still work but may be throttled
            return n;
        }
    }

    // extractNFunction extracts the n-parameter transformation function from player code
    private extractNFunction(): void {
        // Pattern to find the n function name
        const patterns = [
            // Modern pattern
            /\.get\("n"\)\)&&\(b=([a-zA-Z0-9$]+)(?:\[(\d+)\])?\([a-zA-Z0-9]\)/,
            // Alternative pattern
            /\b([a-zA-Z0-9]+)\s*=\s*function\([a-zA-Z]\)\s*\{\s*var\s+[a-zA-Z]=\[[^\]]+\]/,
            // Another variant
            /(?:^|[^a-zA-Z0-9$])([a-zA-Z0-9$]+)\s*=\s*function\([a-z]\)\s*\{(?:[^}]+\}){2,}[^}]+return\s+[a-z]\.join\(""\)/,
        ];

        let funcName = "";
        for (const pattern of patterns) {
            const match = this.playerCode.match(pattern);
            if (match && match[1]) {
                funcName = match[1];
                break;
            }
        }

        if (!funcName) {
            // N function not found, which is okay - some videos don't need it
            return;
        }

        // Extract the function body
        const funcBody = this.extractFunctionBody(funcName);
        if (!funcBody) {
            return;
        }

        // Create wrapper for execution
        this.nFuncCode = `var nFunction = ${funcBody};`;
    }

    // extractFunctionBody extracts a complete function body from the player code
    private extractFunctionBody(funcName: string): string | null {
        // Escape special regex characters in function name
        const escapedName = funcName.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");

        // Try different patterns to find the function
        const patterns = [
            // Function expression: var name = function(...
            new RegExp(`(?:var\\s+)?${escapedName}\\s*=\\s*(function\\([^)]*\\)\\s*\\{[^}]+(?:\\{[^}]*\\}[^}]*)*\\})`),
            // Function declaration: function name(...
            new RegExp(`(function\\s+${escapedName}\\s*\\([^)]*\\)\\s*\\{[^}]+(?:\\{[^}]*\\}[^}]*)*\\})`),
        ];

        for (const pattern of patterns) {
            const match = this.playerCode.match(pattern);
            if (match && match[1]) {
                return match[1];
            }
        }

        // Try to extract using brace matching
        return this.extractFunctionWithBraceMatching(funcName);
    }

    // extractFunctionWithBraceMatching extracts function using brace matching
    private extractFunctionWithBraceMatching(funcName: string): string | null {
        // Find function start
        const escapedName = funcName.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
        const patterns = [
            new RegExp(`${escapedName}\\s*=\\s*function`),
            new RegExp(`function\\s+${escapedName}\\s*\\(`),
        ];

        let startIdx = -1;
        let funcStartOffset = 0;

        for (const pattern of patterns) {
            const match = this.playerCode.match(pattern);
            if (match && match.index !== undefined) {
                startIdx = match.index;
                // Find the actual function keyword
                const funcIdx = this.playerCode.slice(startIdx).indexOf("function");
                if (funcIdx >= 0) {
                    funcStartOffset = funcIdx;
                }
                break;
            }
        }

        if (startIdx < 0) {
            return null;
        }

        // Find opening brace
        const funcStart = startIdx + funcStartOffset;
        let braceStart = this.playerCode.slice(funcStart).indexOf("{");
        if (braceStart < 0) {
            return null;
        }

        braceStart += funcStart + 1;
        let braceCount = 1;
        let endIdx = braceStart;

        while (braceCount > 0 && endIdx < this.playerCode.length) {
            const char = this.playerCode[endIdx];
            if (char === "{") {
                braceCount++;
            } else if (char === "}") {
                braceCount--;
            }
            endIdx++;
        }

        if (braceCount !== 0) {
            return null;
        }

        return this.playerCode.slice(funcStart, endIdx);
    }

    // BulkSolve solves multiple n-parameter challenges
    BulkSolve(challenges: string[]): Map<string, string> {
        const results = new Map<string, string>();

        for (const n of challenges) {
            const solved = this.Solve(n);
            results.set(n, solved);
        }

        return results;
    }
}
