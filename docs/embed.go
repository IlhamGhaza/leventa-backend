// Package docs embeds the OpenAPI specification served at /openapi.yaml.
package docs

import _ "embed"

// OpenAPI is the OpenAPI 3 specification of the REST API.
//
//go:embed openapi.yaml
var OpenAPI []byte
