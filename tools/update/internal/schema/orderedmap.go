// Package schema defines the on-disk JSON schema of packages.json, and
// reads/writes it with the field order and formatting the file uses in the
// repo.
package schema

import (
	"bytes"
	"encoding/json"
	"os"
)

// marshalOrderedObject marshals a JSON object with the given keys, in the
// given order, mapping each key to a value via valueOf. Used to control key
// order in generated JSON, since Go's encoding/json always alphabetizes
// map[string]T keys on marshal.
func marshalOrderedObject(keys []string, valueOf func(key string) (any, error)) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, k := range keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		kb, err := json.Marshal(k)
		if err != nil {
			return nil, err
		}
		buf.Write(kb)
		buf.WriteByte(':')

		v, err := valueOf(k)
		if err != nil {
			return nil, err
		}
		vb, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		buf.Write(vb)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// writeIndentedJSON reindents compact (valid, compact JSON) to 2-space
// indentation with a trailing newline and writes it to path.
func writeIndentedJSON(path string, compact []byte) error {
	var buf bytes.Buffer
	if err := json.Indent(&buf, compact, "", "  "); err != nil {
		return err
	}
	buf.WriteByte('\n')
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

// saveIndented marshals v and writes it to path as 2-space indented JSON
// with a trailing newline.
func saveIndented(path string, v any) error {
	compact, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return writeIndentedJSON(path, compact)
}
