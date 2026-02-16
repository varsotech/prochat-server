package clientmetadata

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

type Response struct {
	ClientID     string   `json:"client_id"`
	RedirectURIs []string `json:"redirect_uris"`
	LogoURI      *string  `json:"logo_uri,omitempty"`
	// Must be one of `web` or `native`, with `web` as the default if not specified.
	// Only
	//ApplicationType *string `json:"application_type,omitempty"`
	//
	//// `authorization_code` must always be included. `refresh_token` is optional, but must be included if the client will make token refresh requests.
	//GrantTypes []string `json:"grant_types"`
	//
	//// All scope values which might be requested by the client are declared here. The `atproto` scope is required, so must be included here.
	//Scope string `json:"scope"`
	//
	//// `code` must be included
	//ResponseTypes []string `json:"response_types"`
	//
	//
	//// Confidential clients must set this to `private_key_jwt`; public must be `none`.
	//// NOTE: Confidential clients are not currently supported: https://github.com/varsotech/prochat-server/issues/7
	//TokenEndpointAuthMethod string `json:"token_endpoint_auth_method"`
	//
	//// `none` is never allowed here. The current recommended and most-supported algorithm is ES256, but this may evolve over time.
	//TokenEndpointAuthSigningAlg *string `json:"token_endpoint_auth_signing_alg,omitempty"`
	//
	//// DPoP is mandatory for all clients, so this must be present and true
	//DPoPBoundAccessTokens bool `json:"dpop_bound_access_tokens"`
	//
	//// NOTE: JWKS is not supported at the moment: https://github.com/varsotech/prochat-server/issues/7
	////JWKS *JWKS `json:"jwks,omitempty"`
	//
	//// URL pointing to a JWKS JSON object. See `jwks` above for details.
	//JWKSURI *string `json:"jwks_uri,omitempty"`
	//
	// human-readable name of the client
	ClientName string `json:"client_name,omitempty"`

	// not to be confused with client_id, this is a homepage URL for the client. If provided, the client_uri must have the same hostname as client_id. Only https: URIs are allowed.
	ClientURI string `json:"client_uri,omitempty"`

	// URL to human-readable terms of service (ToS) for the client. Only https: URIs are allowed.
	TosURI string `json:"tos_uri,omitempty"`

	// URL to human-readable privacy policy for the client. Only https: URIs are allowed.
	PolicyURI string `json:"policy_uri,omitempty"`
}

func (c Response) Validate() error {
	_, err := NewClientID(c.ClientID, false)
	if err != nil {
		return err
	}

	if len(c.RedirectURIs) == 0 {
		return errors.New("no redirect uris provided")
	}

	for _, uri := range c.RedirectURIs {
		err := ValidateRedirectURI(uri)
		if err != nil {
			return fmt.Errorf("invalid redirect uri provided: %w", err)
		}
	}

	if c.LogoURI != nil {
		logoUri, err := url.Parse(*c.LogoURI)
		if err != nil {
			return fmt.Errorf("invalid logo uri provided: %w", err)
		}

		if logoUri.Scheme != "https" {
			return fmt.Errorf("invalid logo uri provided: only https is supported")
		}
	}

	if c.ClientURI != "" {
		clientURI, err := url.Parse(c.ClientURI)
		if err != nil {
			return fmt.Errorf("invalid client uri provided: %w", err)
		}

		if clientURI.Scheme != "https" {
			return fmt.Errorf("invalid client uri provided: only https is supported")
		}
	}

	if c.TosURI != "" {
		tosUri, err := url.Parse(c.TosURI)
		if err != nil {
			return fmt.Errorf("invalid tos uri provided: %w", err)
		}

		if tosUri.Scheme != "https" {
			return fmt.Errorf("invalid tos uri provided: only https is supported")
		}
	}

	if c.PolicyURI != "" {
		policyUri, err := url.Parse(c.PolicyURI)
		if err != nil {
			return fmt.Errorf("invalid policy uri provided: %w", err)
		}

		if policyUri.Scheme != "https" {
			return errors.New("invalid policy uri provided: only https is supported")
		}
	}

	return nil
}

var errInvalidClientID = errors.New("invalid client ID")

// Internal errors defined for test assertion
var errClientIDEmpty = fmt.Errorf("%w: cannot be empty", errInvalidClientID)
var errClientIDMissingHTTPS = fmt.Errorf("%w: must start with https://", errInvalidClientID)
var errClientIDInvalidURL = fmt.Errorf("%w: invalid url", errInvalidClientID)
var errClientIDInvalidHost = fmt.Errorf("%w: invalid url host", errInvalidClientID)
var errClientIDNoURLPath = fmt.Errorf("%w: missing url path", errInvalidClientID)
var errClientIDNoDotPathSegment = fmt.Errorf("%w: dot segments not allowed in path", errInvalidClientID)
var errClientIDNoFragment = fmt.Errorf("%w: fragment component not alloewd in path", errInvalidClientID)
var errClientIDNoQueryParams = fmt.Errorf("%w: query parameters not allowed", errInvalidClientID)

type ClientID string

func (c ClientID) String() string {
	return string(c)
}

func NewClientID(c string, allowLocalhost bool) (ClientID, error) {
	if c == "" {
		return "", errClientIDEmpty
	}

	if allowLocalhost && (c == "http://localhost" || c == "http://localhost/") {
		return ClientID(c), nil
	}

	parsedUrl, err := url.Parse(c)
	if err != nil {
		return "", fmt.Errorf("%w: %w", errClientIDInvalidURL, err)
	}

	if parsedUrl.Host == "" {
		return "", errClientIDInvalidHost
	}

	// MUST have an "https" scheme
	if parsedUrl.Scheme != "https" {
		return "", errClientIDMissingHTTPS
	}

	// MUST contain a path component
	if parsedUrl.Path == "" || parsedUrl.Path == "/" {
		return "", errClientIDNoURLPath
	}

	// MUST NOT contain single-dot or double-dot path segments
	segments := strings.Split(parsedUrl.Path, "/")
	for _, seg := range segments {
		if seg == "." || seg == ".." {
			return "", errClientIDNoDotPathSegment
		}
	}

	// MUST NOT contain a fragment component
	if parsedUrl.Fragment != "" {
		return "", errClientIDNoFragment
	}

	// SHOULD NOT include a query string component
	if parsedUrl.RawQuery != "" {
		return "", errClientIDNoQueryParams
	}

	return ClientID(c), nil
}

type RedirectURI string

var errInvalidRedirectURI = errors.New("invalid redirect uri")

var errRedirectURIEmpty = fmt.Errorf("%w: cannot be empty", errInvalidRedirectURI)
var errRedirectURIInvalidURI = fmt.Errorf("%w: invalid uri", errInvalidRedirectURI)
var errRedirectURIInvalidHost = fmt.Errorf("%w: invalid host", errInvalidRedirectURI)
var errRedirectURINoDotPathSegment = fmt.Errorf("%w: dot segments not allowed in path", errInvalidRedirectURI)
var errRedirectURINoFragment = fmt.Errorf("%w: fragment component not alloewd in path", errInvalidRedirectURI)

func ValidateRedirectURI(redirectUri string) error {
	if redirectUri == "" {
		return errRedirectURIEmpty
	}

	parsedUrl, err := url.Parse(redirectUri)
	if err != nil {
		return fmt.Errorf("%w: %w", errRedirectURIInvalidURI, err)
	}

	if parsedUrl.Host == "" {
		return errRedirectURIInvalidHost
	}

	// MUST NOT contain single-dot or double-dot path segments
	segments := strings.Split(parsedUrl.Path, "/")
	for _, seg := range segments {
		if seg == "." || seg == ".." {
			return errRedirectURINoDotPathSegment
		}
	}

	// MUST NOT contain a fragment component
	if parsedUrl.Fragment != "" {
		return errRedirectURINoFragment
	}

	return nil
}
