export declare class Decipherer {
    private playerCode;
    private playerURL;
    private sigTokens;
    private nSolver;
    constructor(playerCode: string, playerURL?: string);
    Initialize(): Promise<void>;
    GetSignatureTimestamp(): number;
    SolveNChallenge(n: string): string;
    DecipherURL(streamURL: string): string;
    DecipherSignature(sig: string): string;
    private decipherSignature;
    private extractSignatureTokens;
    private extractSignatureTokensAlt;
}
export declare function GetSignatureTimestamp(playerCode: string): number;
export declare function ExtractPlayerID(playerURL: string): string;
export declare function NewDecipherer(playerCode: string, playerURL?: string): Promise<Decipherer>;
//# sourceMappingURL=decipher.d.ts.map