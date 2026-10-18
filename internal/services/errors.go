package services

import (
	"encoding/json"
	"errors"
)

// ActionError is the error ActionService returns to the webview.
// Its text is fixed per kind, never a URL, path or transport error; JS reads err.cause.kind.
type ActionError struct {
	Kind string `json:"kind"`
}

var actionText = map[string]string{
	"busy":              "another action is still running", // App.svelte compares this text
	"download_failed":   "the GIF download failed",
	"not_media":         "the download is not a GIF",
	"too_large":         "the GIF is too large to copy",
	"not_allowed":       "this GIF's address is not allowed",
	"no_link":           "this item has no link to copy",
	"no_page":           "this item has no page to open",
	"open_failed":       "could not open the page",
	"copy_failed":       "could not write to the clipboard",
	"folder_unwritable": "the download folder is not writable",
	"nothing_to_show":   "there is no download to show",
	"crop_failed":       "could not crop this GIF",
	"crop_too_large":    "the GIF is too large to crop",
	"library_read":      "a saved list could not be read",
	"library_write":     "a saved list could not be written",
	"invalid_item":      "this item cannot be saved",
}

func (e *ActionError) Error() string {
	if t, ok := actionText[e.Kind]; ok {
		return t
	}
	return "action failed"
}

func actionErr(kind string) error { return &ActionError{Kind: kind} }

// ErrBusy is returned when another action is still running (each bound call has its own goroutine).
var ErrBusy = actionErr("busy")

// MarshalError makes an *ActionError err.cause in JS; nil leaves every other error to Wails.
// main.go tries provider.MarshalError first.
func MarshalError(err error) []byte {
	var ae *ActionError
	if !errors.As(err, &ae) {
		return nil
	}
	b, _ := json.Marshal(ae) // cannot fail for this struct
	return b
}

// SettingsError is the error SettingsService returns to the webview: fixed text per kind.
type SettingsError struct {
	Kind string `json:"kind"`
	Slot string `json:"slot,omitempty"`
}

func (e *SettingsError) Error() string { return "settings failed" } // RED skeleton
