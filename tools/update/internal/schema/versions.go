package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

// Sigs lists the SIG names in the fixed order used throughout this repo's
// versions.json/hashes.json (and mk-release.nix's sigs attrset).
var Sigs = []string{"cluster-api", "kube-state-metrics", "metrics-server", "external-dns"}

// SigOwner is the GitHub org that owns each SIG's upstream repository.
var SigOwner = map[string]string{
	"cluster-api":        "kubernetes-sigs",
	"kube-state-metrics": "kubernetes",
	"metrics-server":     "kubernetes-sigs",
	"external-dns":       "kubernetes-sigs",
}

// SigSet holds the tracked SIG versions for a Kubernetes minor, in the
// fixed field order versions.json uses.
type SigSet struct {
	ClusterAPI       string `json:"cluster-api"`
	KubeStateMetrics string `json:"kube-state-metrics"`
	MetricsServer    string `json:"metrics-server"`
	ExternalDNS      string `json:"external-dns"`
}

// Get returns the version tracked for the given SIG name, or "" if unknown.
func (s SigSet) Get(sig string) string {
	switch sig {
	case "cluster-api":
		return s.ClusterAPI
	case "kube-state-metrics":
		return s.KubeStateMetrics
	case "metrics-server":
		return s.MetricsServer
	case "external-dns":
		return s.ExternalDNS
	default:
		return ""
	}
}

// Set updates the version tracked for the given SIG name. Unknown SIG names
// are a no-op.
func (s *SigSet) Set(sig, version string) {
	switch sig {
	case "cluster-api":
		s.ClusterAPI = version
	case "kube-state-metrics":
		s.KubeStateMetrics = version
	case "metrics-server":
		s.MetricsServer = version
	case "external-dns":
		s.ExternalDNS = version
	}
}

// MinorEntry is the versions.json value for one Kubernetes minor.
type MinorEntry struct {
	Version string `json:"version"`
	Sigs    SigSet `json:"sigs"`
}

// VersionsFile is the schema of versions.json.
type VersionsFile struct {
	Supported  []string              `json:"supported"`
	Latest     string                `json:"latest"`
	Kubernetes map[string]MinorEntry `json:"kubernetes"`
}

// MarshalJSON writes the kubernetes object with keys ordered per Supported,
// rather than the alphabetical order Go's map marshaling would otherwise
// produce.
func (v VersionsFile) MarshalJSON() ([]byte, error) {
	supportedJSON, err := json.Marshal(v.Supported)
	if err != nil {
		return nil, err
	}
	latestJSON, err := json.Marshal(v.Latest)
	if err != nil {
		return nil, err
	}
	kubernetesJSON, err := marshalOrderedObject(v.Supported, func(k string) (any, error) {
		entry, ok := v.Kubernetes[k]
		if !ok {
			return nil, fmt.Errorf("schema: versions.json missing kubernetes entry for supported minor %q", k)
		}
		return entry, nil
	})
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	buf.WriteString(`{"supported":`)
	buf.Write(supportedJSON)
	buf.WriteString(`,"latest":`)
	buf.Write(latestJSON)
	buf.WriteString(`,"kubernetes":`)
	buf.Write(kubernetesJSON)
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// LoadVersions reads and parses versions.json from path.
func LoadVersions(path string) (*VersionsFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var v VersionsFile
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, fmt.Errorf("schema: parsing %s: %w", path, err)
	}
	return &v, nil
}

// SaveVersions writes v to path as indented JSON, matching the formatting
// convention (2-space indent, trailing newline) already used in the repo.
func SaveVersions(path string, v *VersionsFile) error {
	return saveIndented(path, v)
}
