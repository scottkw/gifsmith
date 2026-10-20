package services

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"

	"gifsmith/internal/update"
)

// plistString returns the <string> value that follows <key>key</key> in a plist dict.
func plistString(t *testing.T, path, key string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	dec := xml.NewDecoder(bytes.NewReader(data))
	dec.Strict = false // the DOCTYPE carries an external DTD
	var last, cur string
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		switch tok := tok.(type) {
		case xml.StartElement:
			cur = tok.Name.Local
		case xml.CharData:
			s := strings.TrimSpace(string(tok))
			if s == "" {
				continue
			}
			if cur == "string" && last == key {
				return s
			}
			if cur == "key" {
				last = s
			} else {
				last = ""
			}
		}
	}
	t.Fatalf("%s: no string for key %s", path, key)
	return ""
}

func TestVersionMatchesInfoPlist(t *testing.T) {
	for _, key := range []string{"CFBundleShortVersionString", "CFBundleVersion"} {
		if got := plistString(t, "../../build/darwin/Info.plist", key); got != Version {
			t.Errorf("Info.plist %s = %q, want services.Version %q", key, got, Version)
		}
	}
}

// The release asset name (Taskfile DMG_NAME) must be the one the updater looks for.
func TestVersionDMGAssetName(t *testing.T) {
	data, err := os.ReadFile("../../build/darwin/Taskfile.yml")
	if err != nil {
		t.Fatal(err)
	}
	vars := map[string]string{}
	for _, m := range regexp.MustCompile(`(?m)^  (VERSION|DMG_NAME): "([^"]+)"$`).FindAllStringSubmatch(string(data), -1) {
		vars[m[1]] = m[2]
	}
	if vars["VERSION"] != Version {
		t.Errorf("Taskfile VERSION = %q, want services.Version %q", vars["VERSION"], Version)
	}
	name := strings.ReplaceAll(vars["DMG_NAME"], "{{.VERSION}}", vars["VERSION"])
	if want := update.DMGAssetName(Version); name != want {
		t.Errorf("Taskfile DMG_NAME = %q, want update.DMGAssetName = %q", name, want)
	}
}

func TestVersionInSnapshot(t *testing.T) {
	f := newSettings(t, "")
	b, err := json.Marshal(f.svc.Snapshot(context.Background()))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if m["version"] != Version {
		t.Errorf("snapshot version = %v, want %q", m["version"], Version)
	}
}

// A2 credits (UI-SPEC copy): every name must be in LICENSE's extension notice and in About.
var creditNames = []string{"Joseph Schmitt", "tm.wrnr", "j3lte", "samuelkraft", "ridemountainpig", "spookywy",
	"seanparkross", "YourMCGeek", "thomas", "tim_gailey", "clins1994", "0xdhrv", "Olli0103", "alexi.build"}

func TestLicenseCredits(t *testing.T) {
	license, err := os.ReadFile("../../LICENSE")
	if err != nil {
		t.Fatal(err)
	}
	about, err := os.ReadFile("../../frontend/src/settings/AboutPane.svelte")
	if err != nil {
		t.Error(err)
	}
	for _, n := range creditNames {
		if !bytes.Contains(license, []byte(n)) {
			t.Errorf("LICENSE lacks contributor %q", n)
		}
		if !bytes.Contains(about, []byte(n)) {
			t.Errorf("About credits lack contributor %q", n)
		}
	}
}
