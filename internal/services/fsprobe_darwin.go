//go:build darwin

package services

import "syscall"

// mntLocal is MNT_LOCAL (golang.org/x/sys/unix zerrors_darwin_arm64.go); stdlib syscall lacks it.
const mntLocal = 0x1000

// Statfs reports the volume type name ("apfs", "exfat", "smbfs"…) and whether it is local.
func Statfs(dir string) (fsType string, local bool, err error) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(dir, &st); err != nil {
		return "", false, err
	}
	b := make([]byte, 0, len(st.Fstypename))
	for _, c := range st.Fstypename {
		if c == 0 {
			break
		}
		b = append(b, byte(c))
	}
	return string(b), st.Flags&mntLocal != 0, nil
}
