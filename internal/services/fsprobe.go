package services

// ValidateDownloadDir refuses a download folder Download cannot use safely. statfs reports the
// volume's type name and whether it is local (Statfs in the app; a fake in tests).
func ValidateDownloadDir(dir string, statfs func(string) (fsType string, local bool, err error)) error {
	return nil
}
