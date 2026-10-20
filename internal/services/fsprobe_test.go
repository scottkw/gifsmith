package services

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func fakeStatfs(fsType string, local bool) func(string) (string, bool, error) {
	return func(string) (string, bool, error) { return fsType, local, nil }
}

func wantFolderErr(t *testing.T, err error, kind, folder string) {
	t.Helper()
	var se *SettingsError
	if !errors.As(err, &se) || se.Kind != kind || se.Folder != folder {
		t.Fatalf("err = %#v, want *SettingsError %s with folder %q", err, kind, folder)
	}
}

func TestValidateDownloadDirAccepts(t *testing.T) {
	dir := t.TempDir()
	if err := ValidateDownloadDir(dir, fakeStatfs("apfs", true)); err != nil {
		t.Fatalf("ValidateDownloadDir(apfs) = %v, want nil", err)
	}
	if left, _ := os.ReadDir(dir); len(left) != 0 {
		t.Errorf("probe files left behind: %v", left)
	}
}

func TestValidateDownloadDirMissing(t *testing.T) {
	dir := t.TempDir()
	wantFolderErr(t, ValidateDownloadDir(filepath.Join(dir, "gone"), fakeStatfs("apfs", true)), "folder_missing", "gone")
	file := filepath.Join(dir, "plain")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	wantFolderErr(t, ValidateDownloadDir(file, fakeStatfs("apfs", true)), "folder_missing", "plain")
}

func TestValidateDownloadDirFAT(t *testing.T) {
	for _, typ := range []string{"exfat", "msdos"} {
		dir := filepath.Join(t.TempDir(), "Stick")
		if err := os.Mkdir(dir, 0o700); err != nil {
			t.Fatal(err)
		}
		wantFolderErr(t, ValidateDownloadDir(dir, fakeStatfs(typ, true)), "folder_fat", "Stick")
	}
}

func TestValidateDownloadDirNetwork(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "Share")
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	wantFolderErr(t, ValidateDownloadDir(dir, fakeStatfs("smbfs", false)), "folder_network", "Share")
}

func TestValidateDownloadDirUnwritable(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root writes into a 0555 folder")
	}
	dir := filepath.Join(t.TempDir(), "Locked")
	if err := os.Mkdir(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	wantFolderErr(t, ValidateDownloadDir(dir, fakeStatfs("apfs", true)), "folder_unwritable", "Locked")
}

func TestValidateDownloadDirStatfsError(t *testing.T) {
	dir := t.TempDir()
	statfs := func(string) (string, bool, error) { return "", false, errors.New("statfs failed") }
	wantFolderErr(t, ValidateDownloadDir(dir, statfs), "folder_unwritable", filepath.Base(dir))
}
