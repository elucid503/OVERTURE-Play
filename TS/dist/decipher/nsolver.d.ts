export declare class NSolver {
    private playerCode;
    private nFuncCode;
    private extractionFailed;
    constructor(playerCode: string);
    Solve(n: string): string;
    private extractNFunction;
    private extractFunctionBody;
    private extractFunctionWithBraceMatching;
    BulkSolve(challenges: string[]): Map<string, string>;
}
//# sourceMappingURL=nsolver.d.ts.map