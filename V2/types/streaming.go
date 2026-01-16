package types

// StreamingProtocol represents the type of streaming protocol
type StreamingProtocol string

const (
	ProtocolHTTPS StreamingProtocol = "https"
	ProtocolDASH  StreamingProtocol = "dash"
	ProtocolHLS   StreamingProtocol = "hls"
)

// PoTokenContext represents the context in which a PO token is used
type PoTokenContext string

const (
	PoTokenContextPlayer PoTokenContext = "player"
	PoTokenContextGVS    PoTokenContext = "gvs"
	PoTokenContextSubs   PoTokenContext = "subs"
)

// PoTokenPolicy defines when PO tokens are required
type PoTokenPolicy struct {
	Required                   bool
	Recommended                bool
	NotRequiredForPremium      bool
	NotRequiredWithPlayerToken bool
}

// DefaultGVSPoTokenPolicy returns the default GVS PO token policy for web clients
func DefaultGVSPoTokenPolicy() PoTokenPolicy {
	return PoTokenPolicy{
		Required:                   true,
		Recommended:                true,
		NotRequiredForPremium:      true,
		NotRequiredWithPlayerToken: false,
	}
}
