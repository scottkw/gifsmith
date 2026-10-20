package services

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"gifsmith/internal/settings"
)

// fakeReg is the Registrar for tests: it records calls in order and fails Register for fail combos.
type fakeReg struct {
	mu    sync.Mutex
	calls []string
	cbs   map[string]func()
	fail  map[string]bool
}

func (r *fakeReg) Register(a string, cb func()) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, "register "+a)
	if r.fail[a] {
		return errors.New("RegisterEventHotKey failed (OSStatus -9878)")
	}
	if _, ok := r.cbs[a]; ok {
		return fmt.Errorf("global shortcut %q is already registered", a)
	}
	if r.cbs == nil {
		r.cbs = map[string]func(){}
	}
	r.cbs[a] = cb
	return nil
}

func (r *fakeReg) Unregister(a string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, "unregister "+a)
	if _, ok := r.cbs[a]; !ok {
		return fmt.Errorf("global shortcut %q is not registered", a)
	}
	delete(r.cbs, a)
	return nil
}

// fire runs a as Carbon would: the callback on its own goroutine, here waited for.
func (r *fakeReg) fire(t *testing.T, a string) {
	t.Helper()
	r.mu.Lock()
	cb := r.cbs[a]
	r.mu.Unlock()
	if cb == nil {
		t.Fatalf("%s is not registered", a)
	}
	done := make(chan struct{})
	go func() { cb(); close(done) }()
	<-done
}

func (r *fakeReg) take() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	c := r.calls
	r.calls = nil
	return c
}

type hotkeyFixture struct {
	h      *Hotkeys
	reg    *fakeReg
	now    *time.Time
	fired  *[]string // actions run, by slot
	events *[]string // "name slot"
}

// newHotkeys starts a Hotkeys on a fake Registrar with b and forgets Start's calls.
func newHotkeys(t *testing.T, b settings.Hotkeys) hotkeyFixture {
	t.Helper()
	var mu sync.Mutex
	reg := &fakeReg{}
	now := time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)
	fired, events := []string{}, []string{}
	act := func(slot string) func() {
		return func() { mu.Lock(); fired = append(fired, slot); mu.Unlock() }
	}
	h := &Hotkeys{
		Reg:     reg,
		Actions: map[string]func(){"open": act("open"), "favorites": act("favorites"), "recents": act("recents")},
		Emit: func(n string, d ...any) {
			mu.Lock()
			defer mu.Unlock()
			events = append(events, strings.TrimSpace(fmt.Sprintln(append([]any{n}, d...)...)))
		},
		Now: func() time.Time { return now },
	}
	h.Start(b)
	reg.take()
	return hotkeyFixture{h: h, reg: reg, now: &now, fired: &fired, events: &events}
}

func TestCanonical(t *testing.T) {
	h := &Hotkeys{}
	for in, want := range map[string]string{
		"Ctrl+Option+Cmd+G":   "Cmd+Ctrl+Option+G",
		"cmd+shift+k":         "Cmd+Shift+K",
		"Cmd+F13":             "Cmd+F13",
		"option+ctrl+space":   "Ctrl+Option+SPACE",
		"Cmd+Up":              "Cmd+UP",
		"Cmd+Shift+Option+5":  "Cmd+Option+Shift+5",
		"Ctrl+=":              "Ctrl+=",
		"Cmd+`":               "Cmd+`",
		"Cmd+Ctrl+Option+G":   "Cmd+Ctrl+Option+G",
		"Ctrl+Option+Cmd+Tab": "Cmd+Ctrl+Option+TAB",
	} {
		if got, err := h.Canonical(in); err != nil || got != want {
			t.Errorf("Canonical(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	for _, in := range []string{"", "Cmd+", "Cmd+Ctrl", "Cmd+plusx", "Cmd++", "Super+G", "Cmd+F21", "Cmd+F0", "Cmd+é", "Cmd+Backspace", "Hyper+G"} {
		_, err := h.Canonical(in)
		wantSettingsErr(t, err, "hotkey_invalid")
	}
}

func TestBindRefusals(t *testing.T) {
	f := newHotkeys(t, settings.Hotkeys{})
	wantSettingsErr(t, f.h.Bind("open", "Shift+G"), "hotkey_no_modifier")
	wantSettingsErr(t, f.h.Bind("open", "G"), "hotkey_no_modifier")
	wantSettingsErr(t, f.h.Bind("open", "Cmd+Q"), "hotkey_system")
	wantSettingsErr(t, f.h.Bind("favorites", "Ctrl+Space"), "hotkey_system")
	wantSettingsErr(t, f.h.Bind("favorites", "Shift+Cmd+3"), "hotkey_system")
	wantSettingsErr(t, f.h.Bind("open", "Cmd+"), "hotkey_invalid")
	wantSettingsErr(t, f.h.Bind("nope", "Cmd+Ctrl+J"), "hotkey_invalid")
	if c := f.reg.take(); len(c) != 0 {
		t.Errorf("Registrar calls = %v, want none", c)
	}
}

func TestBindDuplicate(t *testing.T) {
	f := newHotkeys(t, settings.Hotkeys{})
	err := f.h.Bind("favorites", "Ctrl+Option+Cmd+G")
	var se *SettingsError
	if !errors.As(err, &se) || se.Kind != "hotkey_duplicate" || se.Slot != "open" {
		t.Fatalf("err = %#v, want hotkey_duplicate with Slot open", err)
	}
	if c := f.reg.take(); len(c) != 0 {
		t.Errorf("Registrar calls = %v, want none", c)
	}
}

func TestBindRebind(t *testing.T) {
	f := newHotkeys(t, settings.Hotkeys{})
	if err := f.h.Bind("open", "Cmd+Ctrl+Option+H"); err != nil {
		t.Fatal(err)
	}
	want := []string{"unregister Cmd+Ctrl+Option+G", "register Cmd+Ctrl+Option+H"}
	if c := f.reg.take(); !reflect.DeepEqual(c, want) {
		t.Errorf("calls = %v, want %v", c, want)
	}
	if got := f.h.Bindings().Open; got != "Cmd+Ctrl+Option+H" {
		t.Errorf("Open = %q", got)
	}
	f.reg.fire(t, "Cmd+Ctrl+Option+H")
	if !reflect.DeepEqual(*f.fired, []string{"open"}) {
		t.Errorf("fired = %v, want [open]", *f.fired)
	}
}

func TestBindRollback(t *testing.T) {
	f := newHotkeys(t, settings.Hotkeys{})
	f.reg.fail = map[string]bool{"Cmd+Ctrl+Option+J": true}
	wantSettingsErr(t, f.h.Bind("open", "Ctrl+Option+Cmd+J"), "hotkey_register_failed")
	want := []string{"unregister Cmd+Ctrl+Option+G", "register Cmd+Ctrl+Option+J", "register Cmd+Ctrl+Option+G"}
	if c := f.reg.take(); !reflect.DeepEqual(c, want) {
		t.Errorf("calls = %v, want %v", c, want)
	}
	if got := f.h.Bindings().Open; got != DefaultOpenHotkey {
		t.Errorf("Open = %q, want the old binding", got)
	}
	f.reg.fire(t, DefaultOpenHotkey) // the old binding still works
}

func TestBindRestoreFailsLogs(t *testing.T) {
	f := newHotkeys(t, settings.Hotkeys{})
	buf := captureLog(t)
	f.reg.fail = map[string]bool{"Cmd+Ctrl+Option+J": true, DefaultOpenHotkey: true}
	wantSettingsErr(t, f.h.Bind("open", "Cmd+Ctrl+Option+J"), "hotkey_register_failed")
	if !strings.Contains(buf.String(), "hotkey restore failed") {
		t.Errorf("log = %q, want a restore failure line", buf.String())
	}
	if got := f.h.Bindings().Open; got != "" {
		t.Errorf("Open = %q, want cleared after a failed restore", got)
	}
}

func TestBindFailedNew(t *testing.T) {
	f := newHotkeys(t, settings.Hotkeys{})
	f.reg.fail = map[string]bool{"Cmd+Ctrl+Option+F": true}
	wantSettingsErr(t, f.h.Bind("favorites", "Cmd+Ctrl+Option+F"), "hotkey_register_failed_new")
	if got := f.h.Bindings().Favorites; got != "" {
		t.Errorf("Favorites = %q, want empty", got)
	}
	if c := f.reg.take(); !reflect.DeepEqual(c, []string{"register Cmd+Ctrl+Option+F"}) {
		t.Errorf("calls = %v", c)
	}
}

func TestBindSameIsNoop(t *testing.T) {
	f := newHotkeys(t, settings.Hotkeys{})
	if err := f.h.Bind("open", "Ctrl+Option+Cmd+G"); err != nil {
		t.Fatal(err)
	}
	if c := f.reg.take(); len(c) != 0 {
		t.Errorf("calls = %v, want none", c)
	}
}

func TestStartRegistersDefault(t *testing.T) {
	buf := captureLog(t)
	reg := &fakeReg{}
	h := &Hotkeys{Reg: reg, Actions: map[string]func(){}, Now: time.Now}
	h.Start(settings.Hotkeys{})
	if c := reg.take(); !reflect.DeepEqual(c, []string{"register Cmd+Ctrl+Option+G"}) {
		t.Errorf("calls = %v", c)
	}
	if !strings.Contains(buf.String(), "hotkey registered: Cmd+Ctrl+Option+G") {
		t.Errorf("log = %q", buf.String())
	}
	if b := h.Bindings(); b != (settings.Hotkeys{Open: DefaultOpenHotkey}) {
		t.Errorf("Bindings = %+v", b)
	}
}

func TestStartRegistersStored(t *testing.T) {
	reg := &fakeReg{}
	h := &Hotkeys{Reg: reg, Actions: map[string]func(){}, Now: time.Now}
	h.Start(settings.Hotkeys{Open: "Cmd+Ctrl+Option+H", Favorites: "Cmd+Ctrl+Option+F"})
	want := []string{"register Cmd+Ctrl+Option+H", "register Cmd+Ctrl+Option+F"}
	if c := reg.take(); !reflect.DeepEqual(c, want) {
		t.Errorf("calls = %v, want %v", c, want)
	}
}

// settingsWithHotkeys is the SettingsService fixture with a started Hotkeys on a fake Registrar.
func settingsWithHotkeys(t *testing.T) (settingsFixture, *fakeReg) {
	t.Helper()
	f := newSettings(t, "")
	reg := &fakeReg{}
	f.svc.Hotkeys = &Hotkeys{Reg: reg, Actions: map[string]func(){}, Now: time.Now}
	f.svc.Hotkeys.Start(f.svc.Store.Get().Hotkeys)
	reg.take()
	return f, reg
}

func TestSetHotkeyStores(t *testing.T) {
	f, _ := settingsWithHotkeys(t)
	ctx := context.Background()
	snap, err := f.svc.SetHotkey(ctx, "open", "Ctrl+Option+Cmd+H")
	if err != nil {
		t.Fatal(err)
	}
	if snap.Hotkeys.Open != "Cmd+Ctrl+Option+H" {
		t.Errorf("snapshot hotkeys = %+v", snap.Hotkeys)
	}
	if got := f.svc.Store.Get().Hotkeys.Open; got != "Cmd+Ctrl+Option+H" {
		t.Errorf("stored Open = %q, want canonical", got)
	}
	if ev := *f.events; len(ev) != 1 || ev[0] != "settings:changed" {
		t.Errorf("events = %v", ev)
	}
	if len(*f.data) == 0 {
		t.Fatal("no event payload")
	}
	if s, ok := (*f.data)[0].(Snapshot); !ok || s.Hotkeys.Open != "Cmd+Ctrl+Option+H" {
		t.Errorf("event payload = %#v", (*f.data)[0])
	}
	// Back to the default stores "" (Open "" = default).
	if _, err := f.svc.SetHotkey(ctx, "open", DefaultOpenHotkey); err != nil {
		t.Fatal(err)
	}
	if got := f.svc.Store.Get().Hotkeys.Open; got != "" {
		t.Errorf("stored Open = %q, want empty for the default", got)
	}
}

func TestSetHotkeyErrorKeepsStore(t *testing.T) {
	f, reg := settingsWithHotkeys(t)
	ctx := context.Background()
	reg.fail = map[string]bool{"Cmd+Ctrl+Option+J": true}
	snap, err := f.svc.SetHotkey(ctx, "open", "Cmd+Ctrl+Option+J")
	wantSettingsErr(t, err, "hotkey_register_failed")
	if snap.Hotkeys.Open != DefaultOpenHotkey {
		t.Errorf("snapshot Open = %q, want the old binding", snap.Hotkeys.Open)
	}
	if got := f.svc.Store.Get().Hotkeys; got != (settings.Hotkeys{}) {
		t.Errorf("stored = %+v, want unchanged", got)
	}
	if len(*f.events) != 0 {
		t.Errorf("events = %v, want none", *f.events)
	}
}

func TestMarshalHotkeyDuplicate(t *testing.T) {
	b := MarshalError(&SettingsError{Kind: "hotkey_duplicate", Slot: "favorites"})
	if string(b) != `{"kind":"hotkey_duplicate","slot":"favorites"}` {
		t.Errorf("MarshalError = %s", b)
	}
}

func TestArmTestReportsWithinFiveSeconds(t *testing.T) {
	f := newHotkeys(t, settings.Hotkeys{Favorites: "Cmd+Ctrl+Option+F"})
	if err := f.h.ArmTest("favorites"); err != nil {
		t.Fatal(err)
	}
	*f.now = f.now.Add(2 * time.Second)
	f.reg.fire(t, "Cmd+Ctrl+Option+F")
	if !reflect.DeepEqual(*f.events, []string{"hotkey:test favorites"}) || !reflect.DeepEqual(*f.fired, []string{"favorites"}) {
		t.Fatalf("events %v, fired %v; want the test event, then the action", *f.events, *f.fired)
	}
	f.reg.fire(t, "Cmd+Ctrl+Option+F") // disarmed after one report
	if err := f.h.ArmTest("favorites"); err != nil {
		t.Fatal(err)
	}
	*f.now = f.now.Add(6 * time.Second)
	f.reg.fire(t, "Cmd+Ctrl+Option+F")
	if len(*f.events) != 1 || len(*f.fired) != 3 {
		t.Errorf("events %v, fired %v; a late press reports nothing but still acts", *f.events, *f.fired)
	}
}

func TestArmTestUnbound(t *testing.T) {
	f := newHotkeys(t, settings.Hotkeys{})
	wantSettingsErr(t, f.h.ArmTest("recents"), "hotkey_invalid")
	wantSettingsErr(t, f.h.ArmTest("nope"), "hotkey_invalid")
}

func TestSuspendResumeIdempotent(t *testing.T) {
	f := newHotkeys(t, settings.Hotkeys{Favorites: "Cmd+Ctrl+Option+F"})
	f.h.Suspend()
	f.h.Suspend()
	f.h.Resume()
	f.h.Resume()
	want := []string{"unregister Cmd+Ctrl+Option+G", "unregister Cmd+Ctrl+Option+F", "register Cmd+Ctrl+Option+G", "register Cmd+Ctrl+Option+F"}
	if c := f.reg.take(); !reflect.DeepEqual(c, want) {
		t.Errorf("calls = %v, want %v", c, want)
	}
}

func TestSuspendedComboIsCapturedNotFired(t *testing.T) {
	f := newHotkeys(t, settings.Hotkeys{})
	f.h.Suspend()
	f.reg.mu.Lock()
	_, live := f.reg.cbs[DefaultOpenHotkey]
	f.reg.mu.Unlock()
	if live {
		t.Fatal("the Open hotkey is still registered while suspended")
	}
	f.h.Resume() // what the Settings window's hide / lost-focus hook does
	f.reg.fire(t, DefaultOpenHotkey)
}

func TestSetHotkeyResumesFirst(t *testing.T) {
	f, reg := settingsWithHotkeys(t)
	ctx := context.Background()
	f.svc.SuspendHotkeys(ctx)
	reg.take()
	if _, err := f.svc.SetHotkey(ctx, "favorites", "Cmd+Ctrl+Option+F"); err != nil {
		t.Fatal(err)
	}
	want := []string{"register Cmd+Ctrl+Option+G", "register Cmd+Ctrl+Option+F"}
	if c := reg.take(); !reflect.DeepEqual(c, want) {
		t.Errorf("calls = %v, want %v", c, want)
	}
}

func TestBindClear(t *testing.T) {
	f := newHotkeys(t, settings.Hotkeys{Recents: "Cmd+Ctrl+Option+R"})
	if err := f.h.Bind("recents", ""); err != nil {
		t.Fatal(err)
	}
	if c := f.reg.take(); !reflect.DeepEqual(c, []string{"unregister Cmd+Ctrl+Option+R"}) {
		t.Errorf("calls = %v", c)
	}
	if f.h.Bindings().Recents != "" {
		t.Error("Recents still bound")
	}
	wantSettingsErr(t, f.h.Bind("open", ""), "hotkey_invalid")
}

func TestSetHotkeyOpenUpdatesTray(t *testing.T) {
	f, _ := settingsWithHotkeys(t)
	var got []string
	f.svc.OnOpenHotkey = func(a string) { got = append(got, a) }
	ctx := context.Background()
	if _, err := f.svc.SetHotkey(ctx, "favorites", "Cmd+Ctrl+Option+F"); err != nil {
		t.Fatal(err)
	}
	if _, err := f.svc.SetHotkey(ctx, "open", "Cmd+Ctrl+Option+H"); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []string{"Cmd+Ctrl+Option+H"}) {
		t.Errorf("OnOpenHotkey calls = %v", got)
	}
}

func TestTestHotkey(t *testing.T) {
	f, _ := settingsWithHotkeys(t)
	ctx := context.Background()
	if err := f.svc.TestHotkey(ctx, "open"); err != nil {
		t.Fatal(err)
	}
	wantSettingsErr(t, f.svc.TestHotkey(ctx, "favorites"), "hotkey_invalid")
}
