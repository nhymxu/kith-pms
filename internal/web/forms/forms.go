// Package forms provides helpers for parsing HTML form submissions.
package forms

import (
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

// ParseIndexed parses form fields like "contact[0][type]" into a dense slice of maps.
// Gaps in indices are dropped (dense-packed). The prefix argument selects the field
// family, e.g. "contact" matches "contact[0][type]", "contact[1][value]", etc.
//
// Example input:
//
//	url.Values{
//	  "contact[0][type]":  {"phone"},
//	  "contact[0][value]": {"555-1234"},
//	  "contact[2][type]":  {"email"},   // gap at index 1 — becomes index 1 in output
//	}
//
// Returns:
//
//	[]map[string]string{
//	  {"type": "phone", "value": "555-1234"},
//	  {"type": "email"},
//	}
func ParseIndexed(values url.Values, prefix string) []map[string]string {
	// Collect distinct numeric indices and their field maps.
	indexMap := map[int]map[string]string{}

	for key, vals := range values {
		if !strings.HasPrefix(key, prefix+"[") {
			continue
		}

		rest := key[len(prefix):]
		// rest is like "[0][type]"
		idxEnd := strings.Index(rest, "]")
		if idxEnd < 2 {
			continue // malformed
		}

		idxStr := rest[1:idxEnd]

		idx, err := strconv.Atoi(idxStr)
		if err != nil || idx < 0 {
			continue
		}

		// field is like "[type]"
		fieldPart := rest[idxEnd+1:]
		if !strings.HasPrefix(fieldPart, "[") || !strings.HasSuffix(fieldPart, "]") {
			continue
		}

		field := fieldPart[1 : len(fieldPart)-1]
		if field == "" {
			continue
		}

		if indexMap[idx] == nil {
			indexMap[idx] = map[string]string{}
		}

		v := ""
		if len(vals) > 0 {
			v = vals[0]
		}

		indexMap[idx][field] = v
	}

	// Sort keys to produce stable, dense-packed output.
	indices := make([]int, 0, len(indexMap))
	for idx := range indexMap {
		indices = append(indices, idx)
	}

	sort.Ints(indices)

	result := make([]map[string]string, 0, len(indices))
	for _, idx := range indices {
		result = append(result, indexMap[idx])
	}

	return result
}

// GetField is a convenience helper that returns a field value from a ParseIndexed row,
// returning an empty string when the key is absent.
func GetField(row map[string]string, key string) string {
	return row[key]
}

// IndexedName formats an indexed form field name, e.g. IndexedName("contact", 0, "type") → "contact[0][type]".
func IndexedName(prefix string, index int, field string) string {
	return fmt.Sprintf("%s[%d][%s]", prefix, index, field)
}
