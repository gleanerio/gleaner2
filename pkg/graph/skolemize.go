package graph

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/rs/xid"
)

// Skolemization replaces blank nodes with URIs. It first attempts to generate
// deterministic, content-based identifiers from node properties (contentUrl,
// name+value, title, name). If no qualifying properties are found, it falls
// back to generating random XIDs.
//
// The URI format includes the schema type:
//
//	<https://gleaner.io/id/{schemaType}/{sha256}>
//
// reference: https://www.w3.org/TR/rdf11-concepts/#dfn-skolem-iri
func Skolemization(nq, key string) (string, error) {
	scanner := bufio.NewScanner(strings.NewReader(nq))

	// need for long lines like in Internet of Water
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	// Collect all lines for property extraction
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	var err = scanner.Err()
	if err != nil {
		log.Errorf("Error decoding source: %s\n", key)
		log.Error(err)
		return nq, err
	}

	// First pass: extract properties for all blank nodes
	nodeProps := collectBlankNodeProperties(lines)

	// Second pass: build the blank node → URI mapping
	// Since a data graph may have several references to any given blank node,
	// we keep a map to ensure consistency across all triples.
	m := make(map[string]string)

	for _, line := range lines {
		parts := strings.SplitN(line, " ", 4)
		if len(parts) < 3 {
			continue
		}
		subj := parts[0]
		obj := parts[2]

		for _, node := range []string{subj, obj} {
			if !strings.HasPrefix(node, "_:") {
				continue
			}
			if _, ok := m[node]; ok {
				continue
			}
			// Try deterministic ID from properties
			if info, found := nodeProps[node]; found {
				if uri := generateDeterministicID(info); uri != "" {
					m[node] = uri
					continue
				}
			}
			// Fall back to XID
			guid := xid.New()
			m[node] = fmt.Sprintf("<https://gleaner.io/xid/genid/%s>", guid.String())
		}
	}

	filebytes := []byte(nq)

	for k, v := range m {
		// The +" " is needed since we have to avoid
		// _:b1 replacing _:b13 with ...3
		filebytes = bytes.Replace(filebytes, []byte(k+" "), []byte(v+" "), -1)
	}

	return string(filebytes), err
}
