package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSource_AddListShow(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LNAV_CLI_CONFIG_DIR", dir)

	out := &bytes.Buffer{}
	root := NewRootCmd()
	root.SetOut(out)
	root.SetErr(out)
	root.SetArgs([]string{"source", "add", "app", "--paths", "/var/log/app.log,/var/log/app.err.log"})
	if err := root.Execute(); err != nil {
		t.Fatalf("add: %v", err)
	}
	if !strings.Contains(out.String(), "added") {
		t.Fatalf("unexpected add output: %s", out.String())
	}

	out.Reset()
	root = NewRootCmd()
	root.SetOut(out)
	root.SetErr(out)
	root.SetArgs([]string{"source", "ls"})
	if err := root.Execute(); err != nil {
		t.Fatalf("ls: %v", err)
	}
	if !strings.Contains(out.String(), "app") {
		t.Fatalf("expected 'app' in ls output: %s", out.String())
	}

	out.Reset()
	root = NewRootCmd()
	root.SetOut(out)
	root.SetErr(out)
	root.SetArgs([]string{"source", "show", "app"})
	if err := root.Execute(); err != nil {
		t.Fatalf("show: %v", err)
	}
	if !strings.Contains(out.String(), "/var/log/app.log") {
		t.Fatalf("expected path in show output: %s", out.String())
	}

	if _, err := os.Stat(filepath.Join(dir, "sources.yaml")); err != nil {
		t.Fatalf("sources.yaml not created: %v", err)
	}
}

func TestSource_Rm(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LNAV_CLI_CONFIG_DIR", dir)

	out := &bytes.Buffer{}
	root := NewRootCmd()
	root.SetOut(out)
	root.SetErr(out)
	root.SetArgs([]string{"source", "add", "app", "--paths", "/tmp/a.log"})
	if err := root.Execute(); err != nil {
		t.Fatalf("add: %v", err)
	}

	out.Reset()
	root = NewRootCmd()
	root.SetOut(out)
	root.SetErr(out)
	root.SetArgs([]string{"source", "rm", "app"})
	if err := root.Execute(); err != nil {
		t.Fatalf("rm: %v", err)
	}
	if !strings.Contains(out.String(), "removed") {
		t.Fatalf("unexpected rm output: %s", out.String())
	}

	out.Reset()
	root = NewRootCmd()
	root.SetOut(out)
	root.SetErr(out)
	root.SetArgs([]string{"source", "show", "app"})
	if err := root.Execute(); err == nil {
		t.Fatalf("expected error after rm, got: %s", out.String())
	}
}
