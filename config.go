package launchdarkly

import "os"

// FromEnv builds a Provider from the plugin's environment. Secrets and endpoint
// come from the plugin process, never from the Rollops target spec (Rollops
// passes only the flag name, environment, and percentage).
//
//	LAUNCHDARKLY_API_URL   base URL (default https://app.launchdarkly.com)
//	LAUNCHDARKLY_TOKEN     API access token (required)
//	LAUNCHDARKLY_PROJECT   project key (default "default")
func FromEnv() Provider {
	base := os.Getenv("LAUNCHDARKLY_API_URL")
	if base == "" {
		base = "https://app.launchdarkly.com"
	}
	return Provider{
		BaseURL: base,
		Token:   os.Getenv("LAUNCHDARKLY_TOKEN"),
		Project: os.Getenv("LAUNCHDARKLY_PROJECT"),
	}
}
