package source

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoad_Empty_NoFile(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "sources.yaml"))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(cfg.Sources) != 0 {
		t.Fatalf("expected empty, got %+v", cfg)
	}
}

func TestResolve_Alias_Paths(t *testing.T) {
	cfg := &Config{Sources: map[string]Source{
		"nginx-prod": {Paths: []string{"/var/log/nginx/*.log"}},
	}}
	got, err := cfg.Resolve([]string{"nginx-prod"})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	want := Resolved{Files: []string{"/var/log/nginx/*.log"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got=%+v want=%+v", got, want)
	}
}

func TestResolve_Alias_Command(t *testing.T) {
	cfg := &Config{Sources: map[string]Source{
		"k8s-api": {Command: "kubectl logs -n kube-system deploy/x"},
	}}
	got, err := cfg.Resolve([]string{"k8s-api"})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if got.StdinCmd != "kubectl logs -n kube-system deploy/x" || len(got.Files) != 0 {
		t.Fatalf("unexpected: %+v", got)
	}
}

func TestResolve_PathPassthrough(t *testing.T) {
	cfg := &Config{}
	got, err := cfg.Resolve([]string{"/var/log/app.log", "/tmp/other.log"})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if !reflect.DeepEqual(got.Files, []string{"/var/log/app.log", "/tmp/other.log"}) {
		t.Fatalf("unexpected: %+v", got)
	}
}

func TestAddSave_Roundtrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sources.yaml")
	cfg := &Config{Sources: map[string]Source{}}
	cfg.Sources["app"] = Source{Paths: []string{"/var/log/app.log"}}
	if err := Save(path, cfg); err != nil {
		t.Fatalf("save: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("stat: %v", err)
	}
	reloaded, err := Load(path)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if !reflect.DeepEqual(reloaded.Sources["app"].Paths, []string{"/var/log/app.log"}) {
		t.Fatalf("roundtrip mismatch: %+v", reloaded)
	}
}
