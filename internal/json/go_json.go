//go:build go_json
// +build go_json

package json

import (
	"github.com/goccy/go-json"
)

var (
	Name          = "go_json"
	Marshal       = json.Marshal
	Unmarshal     = json.Unmarshal
	MarshalIndent = json.MarshalIndent
	NewDecoder    = json.NewDecoder
	NewEncoder    = json.NewEncoder
)
