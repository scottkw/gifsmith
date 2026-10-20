package services

import (
	"log"
	"os"
	"path/filepath"
)

// ValidateDownloadDir refuses a download folder Download cannot use safely. statfs reports the
// volume's type name and whether it is local (Statfs in the app; a fake in tests). The os.Link
// probe is the authoritative check: Download's no-overwrite claim needs hard links (Save links
// the finished temp file into place), which exFAT, FAT and some other volumes lack.
func ValidateDownloadDir(dir string, statfs func(string) (fsType string, local bool, err error)) error {
	refuse := func(kind string) error { return &SettingsError{Kind: kind, Folder: filepath.Base(dir)} }
	if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
		return refuse("folder_missing")
	}
	typ, local, err := statfs(dir)
	switch {
	case err != nil:
		log.Printf("download folder: statfs: %v", err) // never a pass
		return refuse("folder_unwritable")
	case typ == "msdos" || typ == "exfat":
		return refuse("folder_fat")
	case !local:
		return refuse("folder_network")
	}
	f, err := os.CreateTemp(dir, ".gifsmith-probe-*")
	if err != nil {
		return refuse("folder_unwritable")
	}
	tmp := f.Name()
	_ = f.Close() // empty file
	defer func() { _ = os.Remove(tmp); _ = os.Remove(tmp + ".l") }()
	if err := os.Link(tmp, tmp+".l"); err != nil {
		return refuse("folder_fat")
	}
	return nil
}
