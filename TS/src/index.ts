// OVERTURE-Play TypeScript Library
// A YouTube video downloading library with 403 error prevention

// Main Client
export { Client, NewClient, ClientOptions } from "./client/client.js";

// Types
export {
    Video,
    PlayabilityStatus,
    LiveBroadcastDetails,
    FilterFormats,
    VideoFormats,
    AudioFormats,
    BestVideoFormat,
    BestAudioFormat,
    GetFormat,
    HasLiveFormats,
    GetThumbnail,
    DurationString,
    CreateEmptyVideo,
} from "./types/video.js";

export {
    Format,
    HasVideo,
    HasAudio,
    IsAudioOnly,
    IsVideoOnly,
    IsAdaptive,
    SupportsRange,
    Extension,
    FormatID,
    FormatString,
    QualityRank,
    CreateEmptyFormat,
} from "./types/format.js";

export { Thumbnail } from "./types/thumbnail.js";
export { Range } from "./types/range.js";
export { StreamingProtocol, PoTokenPolicy, DefaultGVSPoTokenPolicy } from "./types/streaming.js";

// Innertube Clients
export {
    ClientConfig,
    InnertubeContext,
    ClientInfo,
    ThirdPartyContext,
    GetContext,
    GetContextWithVisitor,
    RequiresPoToken,
    CreateClientConfig,
} from "./innertube/config.js";

export {
    Web,
    WebSafari,
    WebEmbedded,
    WebMusic,
    WebCreator,
    Android,
    AndroidSDKLess,
    AndroidVR,
    IOS,
    MWeb,
    TV,
    TVDowngraded,
    TVSimply,
    TVEmbedded,
    DefaultClients,
    DefaultWebClients,
    DefaultAndroidClients,
    DefaultAuthenticatedClients,
    DefaultPremiumClients,
    GetClientByName,
} from "./innertube/clients.js";

// PO Token Provider
export {
    Provider as POTProvider,
    DefaultServerURL as POTDefaultServerURL,
    Request as POTRequest,
    Response as POTResponse,
    PingResponse as POTPingResponse,
} from "./pot/provider.js";

// Decipher
export {
    Decipherer,
    NewDecipherer,
    GetSignatureTimestamp,
    ExtractPlayerID,
} from "./decipher/decipher.js";

export { NSolver } from "./decipher/nsolver.js";

// Stream Handler
export {
    Handler as StreamHandler,
    NewHandler as NewStreamHandler,
    StreamInfo,
    Progress,
    ProgressCallback,
} from "./stream/handler.js";

// Authentication
export {
    Auth,
    Cookie,
    YouTubeURL,
    ExtractVisitorDataFromHTML,
    ExtractDataSyncIDFromResponse,
} from "./auth/auth.js";
