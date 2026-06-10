package config

import (
	"bytes"
	"testing"

	"github.com/spf13/viper"
)

func viperFromYAML(t *testing.T, yaml string) *viper.Viper {
	t.Helper()
	v := viper.New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(bytes.NewBufferString(yaml)); err != nil {
		t.Fatalf("failed to parse YAML: %v", err)
	}
	return v
}

// --- Sources deserialization ---

func TestGetSources_Basic(t *testing.T) {
	yaml := `
sources:
  - sourcetype: sitemap
    name: samplesearth
    url: https://samples.earth/sitemap.xml
    active: true
    domain: https://samples.earth
`
	v := viperFromYAML(t, yaml)
	sources, err := GetSources(v)
	if err != nil {
		t.Fatalf("GetSources: %v", err)
	}
	if len(sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(sources))
	}
	s := sources[0]
	if s.Name != "samplesearth" {
		t.Errorf("Name = %q, want samplesearth", s.Name)
	}
	if s.SourceType != "sitemap" {
		t.Errorf("SourceType = %q, want sitemap", s.SourceType)
	}
	if s.URL != "https://samples.earth/sitemap.xml" {
		t.Errorf("URL = %q", s.URL)
	}
	if !s.Active {
		t.Error("expected Active=true")
	}
}

func TestGetSources_Defaults(t *testing.T) {
	yaml := `
sources:
  - name: minimal
    url: https://example.org/sitemap.xml
`
	v := viperFromYAML(t, yaml)
	sources, err := GetSources(v)
	if err != nil {
		t.Fatalf("GetSources: %v", err)
	}
	s := sources[0]
	if s.SourceType != "sitemap" {
		t.Errorf("default SourceType = %q, want sitemap", s.SourceType)
	}
	if s.AcceptContentType != "application/ld+json, text/html" {
		t.Errorf("default AcceptContentType = %q", s.AcceptContentType)
	}
}

func TestGetSources_FixContextOption_String(t *testing.T) {
	yaml := `
sources:
  - name: withcontext
    url: https://example.org/sitemap.xml
    fixcontextoption: https
`
	v := viperFromYAML(t, yaml)
	sources, err := GetSources(v)
	if err != nil {
		t.Fatalf("GetSources: %v", err)
	}
	s := sources[0]
	if s.FixContextOption != Https {
		t.Errorf("FixContextOption = %q, want %q", s.FixContextOption, Https)
	}
}

func TestGetSources_FixContextOption_AllValues(t *testing.T) {
	cases := []struct {
		yaml string
		want ContextOption
	}{
		{`fixcontextoption: strict`, Strict},
		{`fixcontextoption: https`, Https},
		{`fixcontextoption: http`, Http},
		{`fixcontextoption: standardizedHttps`, StandardizedHttps},
		{`fixcontextoption: standardizedHttp`, StandardizedHttp},
	}
	for _, tc := range cases {
		t.Run(tc.want, func(t *testing.T) {
			full := "sources:\n  - name: test\n    url: https://example.org\n    " + tc.yaml
			v := viperFromYAML(t, full)
			sources, err := GetSources(v)
			if err != nil {
				t.Fatalf("GetSources: %v", err)
			}
			if sources[0].FixContextOption != tc.want {
				t.Errorf("got %q, want %q", sources[0].FixContextOption, tc.want)
			}
		})
	}
}

func TestGetActiveSources(t *testing.T) {
	yaml := `
sources:
  - name: active1
    url: https://a.example.org
    active: true
  - name: inactive
    url: https://b.example.org
    active: false
  - name: active2
    url: https://c.example.org
    active: true
`
	v := viperFromYAML(t, yaml)
	active, err := GetActiveSources(v)
	if err != nil {
		t.Fatalf("GetActiveSources: %v", err)
	}
	if len(active) != 2 {
		t.Fatalf("expected 2 active sources, got %d", len(active))
	}
}

func TestGetSourceByName(t *testing.T) {
	sources := []Sources{
		{Name: "alpha"},
		{Name: "beta"},
		{Name: "gamma"},
	}

	s, err := GetSourceByName(sources, "beta")
	if err != nil {
		t.Fatalf("GetSourceByName: %v", err)
	}
	if s.Name != "beta" {
		t.Errorf("got %q", s.Name)
	}

	_, err = GetSourceByName(sources, "nonexistent")
	if err == nil {
		t.Error("expected error for missing source")
	}
}

func TestSourceToNabuPrefix(t *testing.T) {
	sources := []Sources{
		{Name: "src1", SourceType: "sitemap"},
		{Name: "src2", SourceType: "googledrive"},
	}
	prefixes := SourceToNabuPrefix(sources, false)
	if len(prefixes) != 2 {
		t.Fatalf("expected 2 prefixes, got %d", len(prefixes))
	}
	if prefixes[0] != "summoned/src1" {
		t.Errorf("prefix[0] = %q, want summoned/src1", prefixes[0])
	}

	milled := SourceToNabuPrefix(sources, true)
	if milled[0] != "milled/src1" {
		t.Errorf("milled prefix[0] = %q, want milled/src1", milled[0])
	}
}

// --- MinIO config ---

func TestGetMinioConfig(t *testing.T) {
	yaml := `
minio:
  address: minio.example.org
  port: 9000
  ssl: true
  accesskey: AKTEST
  secretkey: SKTEST
  bucket: testbucket
  region: us-west-2
`
	v := viperFromYAML(t, yaml)
	cfg, err := GetMinioConfig(v)
	if err != nil {
		t.Fatalf("GetMinioConfig: %v", err)
	}
	if cfg.Address != "minio.example.org" {
		t.Errorf("Address = %q", cfg.Address)
	}
	if cfg.Port != 9000 {
		t.Errorf("Port = %d", cfg.Port)
	}
	if !cfg.Ssl {
		t.Error("expected Ssl=true")
	}
	if cfg.Accesskey != "AKTEST" {
		t.Errorf("Accesskey = %q", cfg.Accesskey)
	}
	if cfg.Bucket != "testbucket" {
		t.Errorf("Bucket = %q", cfg.Bucket)
	}
}

func TestGetBucketName(t *testing.T) {
	yaml := `
minio:
  address: localhost
  port: 9000
  bucket: mybucket
`
	v := viperFromYAML(t, yaml)
	name, err := GetBucketName(v)
	if err != nil {
		t.Fatalf("GetBucketName: %v", err)
	}
	if name != "mybucket" {
		t.Errorf("bucket = %q, want mybucket", name)
	}
}

// --- Endpoints ---

func TestGetEndPointsConfig(t *testing.T) {
	yaml := `
endpoints:
  - service: blazegraph
    baseurl: http://localhost:9999/blazegraph/namespace/kb
    type: blazegraph
    authenticate: false
    modes:
      - action: bulk
        suffix: /sparql
        accept: text/x-nquads
        method: POST
      - action: query
        suffix: /sparql
        accept: application/sparql-results+json
        method: GET
`
	v := viperFromYAML(t, yaml)
	eps, err := GetEndPointsConfig(v)
	if err != nil {
		t.Fatalf("GetEndPointsConfig: %v", err)
	}
	if len(eps) != 1 {
		t.Fatalf("expected 1 endpoint, got %d", len(eps))
	}
	if eps[0].Service != "blazegraph" {
		t.Errorf("Service = %q", eps[0].Service)
	}
	if len(eps[0].Modes) != 2 {
		t.Fatalf("expected 2 modes, got %d", len(eps[0].Modes))
	}
}

func TestGetEndpoint_Resolves(t *testing.T) {
	yaml := `
endpoints:
  - service: blazegraph
    baseurl: http://localhost:9999/blazegraph/namespace/kb
    type: blazegraph
    authenticate: false
    username: admin
    password: secret
    modes:
      - action: bulk
        suffix: /sparql
        accept: text/x-nquads
        method: POST
`
	v := viperFromYAML(t, yaml)
	sm, err := GetEndpoint(v, "blazegraph", "bulk")
	if err != nil {
		t.Fatalf("GetEndpoint: %v", err)
	}
	if sm.URL != "http://localhost:9999/blazegraph/namespace/kb/sparql" {
		t.Errorf("URL = %q", sm.URL)
	}
	if sm.Method != "POST" {
		t.Errorf("Method = %q", sm.Method)
	}
	if sm.Accept != "text/x-nquads" {
		t.Errorf("Accept = %q", sm.Accept)
	}
	if sm.Type != "blazegraph" {
		t.Errorf("Type = %q", sm.Type)
	}
}

func TestGetEndpoint_SingleEndpointDefaultService(t *testing.T) {
	yaml := `
endpoints:
  - service: graphdb
    baseurl: http://localhost:7200/repositories/test
    type: graphdb
    modes:
      - action: bulk
        suffix: /statements
        accept: text/x-nquads
        method: POST
`
	v := viperFromYAML(t, yaml)
	sm, err := GetEndpoint(v, "", "bulk")
	if err != nil {
		t.Fatalf("GetEndpoint with empty service: %v", err)
	}
	if sm.Service != "graphdb" {
		t.Errorf("Service = %q, expected default to graphdb", sm.Service)
	}
}

// --- Summoner ---

func TestReadSummonerConfig(t *testing.T) {
	yaml := `
summoner:
  after: "2024-01-01"
  mode: diff
  threads: 3
  delay: 5000
  headless: http://127.0.0.1:9222
  identifiertype: identifiersha
`
	v := viperFromYAML(t, yaml)
	sub := v.Sub("summoner")
	if sub == nil {
		t.Fatal("summoner subtree is nil")
	}
	s, err := ReadSummonerConfig(sub)
	if err != nil {
		t.Fatalf("ReadSummonerConfig: %v", err)
	}
	if s.Mode != "diff" {
		t.Errorf("Mode = %q, want diff", s.Mode)
	}
	if s.Threads != 3 {
		t.Errorf("Threads = %d, want 3", s.Threads)
	}
	if s.Delay != 5000 {
		t.Errorf("Delay = %d, want 5000", s.Delay)
	}
}

func TestReadSummmonerConfig_Compat(t *testing.T) {
	yaml := `
summoner:
  mode: full
  threads: 5
  headless: http://127.0.0.1:9222
`
	v := viperFromYAML(t, yaml)
	sub := v.Sub("summoner")
	s, err := ReadSummmonerConfig(sub)
	if err != nil {
		t.Fatalf("ReadSummmonerConfig (triple-m): %v", err)
	}
	if s.Mode != "full" {
		t.Errorf("Mode = %q", s.Mode)
	}
}

// --- Objects ---

func TestGetObjectsConfig(t *testing.T) {
	yaml := `
objects:
  bucket: gleaner
  domain: us-east-1
  prefix:
    - summoned/samplesearth
    - milled/samplesearth
`
	v := viperFromYAML(t, yaml)
	obj, err := GetObjectsConfig(v)
	if err != nil {
		t.Fatalf("GetObjectsConfig: %v", err)
	}
	if obj.Bucket != "gleaner" {
		t.Errorf("Bucket = %q", obj.Bucket)
	}
	if obj.Domain != "us-east-1" {
		t.Errorf("Domain = %q", obj.Domain)
	}
	if len(obj.Prefix) != 2 {
		t.Fatalf("expected 2 prefixes, got %d", len(obj.Prefix))
	}
}

// --- Nabu config read (from buffer) ---

func TestReadNabuConfig_FullRoundtrip(t *testing.T) {
	yaml := `
minio:
  address: localhost
  port: 9000
  ssl: false
  accesskey: testkey
  secretkey: testsecret
  bucket: gleaner
sparql:
  endpoint: http://localhost:9999/blazegraph/namespace/kb/sparql
  authenticate: false
endpoints:
  - service: blazegraph
    baseurl: http://localhost:9999/blazegraph/namespace/kb
    type: blazegraph
    authenticate: false
    modes:
      - action: bulk
        suffix: /sparql
        accept: text/x-nquads
        method: POST
objects:
  bucket: gleaner
  domain: us-east-1
  prefix:
    - summoned/samplesearth
sources:
  - name: samplesearth
    url: https://samples.earth/sitemap.xml
    sourcetype: sitemap
    active: true
    fixcontextoption: https
`
	v := viperFromYAML(t, yaml)

	// Verify all sections deserialize
	cfg, err := GetMinioConfig(v)
	if err != nil {
		t.Fatalf("MinIO: %v", err)
	}
	if cfg.Address != "localhost" {
		t.Errorf("MinIO.Address = %q", cfg.Address)
	}

	eps, err := GetEndPointsConfig(v)
	if err != nil {
		t.Fatalf("Endpoints: %v", err)
	}
	if len(eps) != 1 {
		t.Errorf("expected 1 endpoint, got %d", len(eps))
	}

	obj, err := GetObjectsConfig(v)
	if err != nil {
		t.Fatalf("Objects: %v", err)
	}
	if len(obj.Prefix) != 1 {
		t.Errorf("expected 1 prefix, got %d", len(obj.Prefix))
	}

	sources, err := GetSources(v)
	if err != nil {
		t.Fatalf("Sources: %v", err)
	}
	if len(sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(sources))
	}
	if sources[0].FixContextOption != "https" {
		t.Errorf("FixContextOption = %q, want https", sources[0].FixContextOption)
	}
}

// --- Utilities ---

func TestMoveToFront(t *testing.T) {
	cases := []struct {
		needle string
		input  []string
		first  string
	}{
		{"b", []string{"a", "b", "c"}, "b"},
		{"a", []string{"a", "b", "c"}, "a"},
		{"c", []string{"a", "b", "c"}, "c"},
		{"d", []string{"a", "b", "c"}, "d"},
	}
	for _, tc := range cases {
		result := MoveToFront(tc.needle, tc.input)
		if result[0] != tc.first {
			t.Errorf("MoveToFront(%q, %v)[0] = %q, want %q", tc.needle, tc.input, result[0], tc.first)
		}
	}
}
