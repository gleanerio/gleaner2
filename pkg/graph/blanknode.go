package graph

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

// schemaOrgPredicates maps short property names to their possible full schema.org predicate URIs
// as they appear in N-Quads (with angle brackets).
var schemaOrgPredicates = map[string][]string{
	"contentUrl": {
		"<https://schema.org/contentUrl>",
		"<http://schema.org/contentUrl>",
	},
	"name": {
		"<https://schema.org/name>",
		"<http://schema.org/name>",
	},
	"value": {
		"<https://schema.org/value>",
		"<http://schema.org/value>",
	},
	"title": {
		"<https://schema.org/title>",
		"<http://schema.org/title>",
	},
}

var rdfTypePreds = []string{
	"<http://www.w3.org/1999/02/22-rdf-syntax-ns#type>",
}

// blankNodeInfo holds the extracted properties for a single blank node.
type blankNodeInfo struct {
	rdfType    string // e.g. "Person", "Dataset"
	contentUrl string
	name       string
	value      string
	title      string
}

// parseNQuadLine extracts subject, predicate, and object from an N-Quads line,
// correctly handling quoted literals that contain spaces.
func parseNQuadLine(line string) (subj, pred, obj string, ok bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return "", "", "", false
	}

	// subject is always first space-delimited token (IRI or blank node)
	idx := strings.IndexByte(line, ' ')
	if idx < 0 {
		return "", "", "", false
	}
	subj = line[:idx]
	line = line[idx+1:]

	// predicate is next space-delimited token (IRI)
	idx = strings.IndexByte(line, ' ')
	if idx < 0 {
		return "", "", "", false
	}
	pred = line[:idx]
	line = line[idx+1:]

	// object: may be a quoted literal (starts with ") or IRI/blank node
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "\"") {
		// Find end of quoted string (handle escaped quotes)
		i := 1
		for i < len(line) {
			if line[i] == '\\' {
				i += 2
				continue
			}
			if line[i] == '"' {
				// Include everything up to the trailing " ." or end
				// The object includes the quotes plus any datatype/lang tag
				rest := line[i+1:]
				dotIdx := strings.Index(rest, " .")
				if dotIdx >= 0 {
					obj = line[:i+1+dotIdx]
				} else {
					obj = line[:i+1]
				}
				return subj, pred, strings.TrimSpace(obj), true
			}
			i++
		}
		// Malformed, return what we have
		obj = line
	} else {
		// IRI or blank node — take until next space
		idx = strings.IndexByte(line, ' ')
		if idx >= 0 {
			obj = line[:idx]
		} else {
			obj = line
		}
	}
	return subj, pred, obj, true
}

// collectBlankNodeProperties parses N-Quads lines and builds a map from blank node ID
// to its extracted properties (type, contentUrl, name, value, title).
func collectBlankNodeProperties(lines []string) map[string]*blankNodeInfo {
	nodes := make(map[string]*blankNodeInfo)

	for _, line := range lines {
		subj, pred, obj, ok := parseNQuadLine(line)
		if !ok {
			continue
		}

		if !strings.HasPrefix(subj, "_:") {
			continue
		}

		info, ok := nodes[subj]
		if !ok {
			info = &blankNodeInfo{}
			nodes[subj] = info
		}

		// Check for rdf:type
		for _, typePred := range rdfTypePreds {
			if pred == typePred {
				info.rdfType = extractSchemaType(obj)
			}
		}

		// Check for schema.org properties
		objVal := extractLiteralValue(obj)

		for _, p := range schemaOrgPredicates["contentUrl"] {
			if pred == p {
				info.contentUrl = objVal
			}
		}
		for _, p := range schemaOrgPredicates["name"] {
			if pred == p {
				info.name = objVal
			}
		}
		for _, p := range schemaOrgPredicates["value"] {
			if pred == p {
				info.value = objVal
			}
		}
		for _, p := range schemaOrgPredicates["title"] {
			if pred == p {
				info.title = objVal
			}
		}
	}

	return nodes
}

// generateDeterministicID produces a content-based URI for a blank node.
// Returns empty string if no qualifying properties are found (caller should fall back to XID).
//
// Priority:
//  1. contentUrl
//  2. name + value (both must be present)
//  3. title
//  4. name
//
// URI format: <https://gleaner.io/id/{schemaType}/{sha256}>
func generateDeterministicID(info *blankNodeInfo) string {
	if info == nil {
		return ""
	}

	schemaType := info.rdfType
	if schemaType == "" {
		schemaType = "Thing"
	}

	var seed string
	switch {
	case info.contentUrl != "":
		seed = "contentUrl:" + info.contentUrl
	case info.name != "" && info.value != "":
		seed = "name:" + info.name + "|value:" + info.value
	case info.title != "":
		seed = "title:" + info.title
	case info.name != "":
		seed = "name:" + info.name
	default:
		return ""
	}

	hash := sha256.Sum256([]byte(seed))
	return fmt.Sprintf("<https://gleaner.io/id/%s/%x>", schemaType, hash)
}

// extractSchemaType takes an RDF type URI like <https://schema.org/Person> and returns "Person".
func extractSchemaType(obj string) string {
	obj = strings.Trim(obj, "<>")
	// Handle both http and https schema.org
	for _, prefix := range []string{"https://schema.org/", "http://schema.org/"} {
		if strings.HasPrefix(obj, prefix) {
			return strings.TrimPrefix(obj, prefix)
		}
	}
	// For non-schema.org types, prefer fragment, then last path segment
	if i := strings.LastIndex(obj, "#"); i >= 0 {
		return obj[i+1:]
	}
	if i := strings.LastIndex(obj, "/"); i >= 0 {
		return obj[i+1:]
	}
	return obj
}

// extractLiteralValue extracts the string value from an N-Quads literal or IRI.
// e.g. `"some value"` -> `some value`, `"val"^^<xsd:string>` -> `val`,
// `<https://example.org/file.csv>` -> `https://example.org/file.csv`
func extractLiteralValue(obj string) string {
	obj = strings.TrimSpace(obj)

	// IRI: <...>
	if strings.HasPrefix(obj, "<") && strings.Contains(obj, ">") {
		end := strings.Index(obj, ">")
		return obj[1:end]
	}

	// Literal: "..."
	if strings.HasPrefix(obj, "\"") {
		// Find the closing quote (handle escaped quotes)
		s := obj[1:]
		var result strings.Builder
		for i := 0; i < len(s); i++ {
			if s[i] == '\\' && i+1 < len(s) {
				result.WriteByte(s[i+1])
				i++
				continue
			}
			if s[i] == '"' {
				return result.String()
			}
			result.WriteByte(s[i])
		}
	}

	return obj
}
