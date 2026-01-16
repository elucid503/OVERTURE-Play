// Package innertube provides YouTube innertube client configurations.
// These configurations are based on the yt-dlp reference implementation.
package innertube

import (
	"github.com/elucid503/overture-play/v2/types"
)

// ClientConfig represents an innertube client configuration
type ClientConfig struct {
	Name        string
	Version     string
	APIKey      string

	UserAgent   string
	DeviceMake  string
	DeviceModel string
	OSName      string
	OSVersion   string

	Host        string
	ContextName int

	RequireJSPlayer           bool
	SupportsCookies           bool
	SupportsAdPlaybackContext bool
	RequireAuth               bool

	// PO Token policies per streaming protocol
	GVSPoTokenPolicies  map[types.StreamingProtocol]types.PoTokenPolicy
	PlayerPoTokenPolicy types.PoTokenPolicy
	SubsPoTokenPolicy   types.PoTokenPolicy

	// Convenience field for direct access
	POTokenPolicy types.PoTokenPolicy
}

// InnertubeContext represents the client context sent with API requests
type InnertubeContext struct {
	Client ClientInfo `json:"client"`
}

// ClientInfo contains client identification information
type ClientInfo struct {
	ClientName     string `json:"clientName"`
	ClientVersion  string `json:"clientVersion"`
	UserAgent      string `json:"userAgent,omitempty"`
	DeviceMake     string `json:"deviceMake,omitempty"`
	DeviceModel    string `json:"deviceModel,omitempty"`
	OSName         string `json:"osName,omitempty"`
	OSVersion      string `json:"osVersion,omitempty"`
	AndroidSDKVer  int    `json:"androidSdkVersion,omitempty"`
	HL             string `json:"hl"`
	TimeZone       string `json:"timeZone"`
	UTCOffsetMins  int    `json:"utcOffsetMinutes"`
	VisitorData    string `json:"visitorData,omitempty"`
}

// ThirdPartyContext for embedded player context
type ThirdPartyContext struct {
	EmbedURL string `json:"embedUrl"`
}

// GetContext returns an InnertubeContext for this client config
func (c *ClientConfig) GetContext() InnertubeContext {
	return InnertubeContext{
		Client: ClientInfo{
			ClientName:    c.Name,
			ClientVersion: c.Version,
			UserAgent:     c.UserAgent,
			DeviceMake:    c.DeviceMake,
			DeviceModel:   c.DeviceModel,
			OSName:        c.OSName,
			OSVersion:     c.OSVersion,
			HL:            "en",
			TimeZone:      "UTC",
			UTCOffsetMins: 0,
		},
	}
}

// GetContextWithVisitor returns an InnertubeContext with visitor data
func (c *ClientConfig) GetContextWithVisitor(visitorData string) InnertubeContext {
	ctx := c.GetContext()
	ctx.Client.VisitorData = visitorData
	return ctx
}

// RequiresPoToken returns true if this client requires a PO token
func (c *ClientConfig) RequiresPoToken() bool {
	// Check if any GVS policy requires a token
	for _, policy := range c.GVSPoTokenPolicies {
		if policy.Required {
			return true
		}
	}
	return c.PlayerPoTokenPolicy.Required
}
