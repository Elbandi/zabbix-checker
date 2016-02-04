package lld

// From: https://github.com/cavaliercoder/g2z/blob/master/lld.go

import (
	"bytes"
	"regexp"
	"strings"
)

// DiscoveryData is a slice of DiscoveryItems which is returned for registered discovery rules
// as a JSON encoded string.
type DiscoveryData []DiscoveryItem

// DiscoveryItem is a map of key/value pairs that represents a single instance of a discovered
// asset.
type DiscoveryItem map[string]string

// macroIllegalPattern is a regular expression pattern matching characters which are illegal in a
// Zabbix discovery macro name.
var macroIllegalPattern = regexp.MustCompile(`[^A-Z0-9_]+`)

// Json converts a DiscoveryData struct into a JSON encoded string, compatible with Zabbix
// Low-Level discovery rules from v2.2.0 and above.
func (c DiscoveryData) Json() string {
	b := bytes.Buffer{}

	b.WriteString("{\n\t\"data\":[")

	for i, item := range c {
		if i > 0 {
			b.WriteString(",")
		}

		b.WriteString("\n\t{")

		firstMacro := true
		for macro, val := range item {
			if firstMacro {
				firstMacro = false
			} else {
				b.WriteString(",")
			}

			b.WriteString("\n\t\t\"{#")
			b.WriteString(macroName(macro))
			b.WriteString("}\":\"")
			b.WriteString(jsonEscape(val))
			b.WriteString("\"")
		}

		b.WriteString("}")
	}

	b.WriteString("]}")

	return b.String()
}

// escape JSON values to prevent invalidating discovery response body
func jsonEscape(a string) string {
	return strings.Replace(a, "\"", "\\\"", -1)
}

// format a name string as a discovery macro (E.g `{#MY_MACRO}`)
func macroName(name string) string {
	name = strings.ToUpper(name)
	name = strings.Replace(name, " ", "_", -1)
	name = macroIllegalPattern.ReplaceAllString(name, "")
	return name
}