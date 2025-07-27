package api

// APIInfo contains metadata about an API provider
type APIInfo struct {
	Name         string
	DefaultModel string
	Models       []string
}

// AvailableAPIs maps API names to their configuration
var AvailableAPIs = map[string]APIInfo{
	"gemini": {
		Name:         "Google Gemini",
		DefaultModel: "gemini-2.5-flash-lite",
		Models:       []string{"gemini-2.5-flash-lite", "gemini-2.5-flash"},
	},
}