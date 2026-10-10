//go:build !darwin

package probe

func readDirSizes(dir string) (int64, []string) {
	return genericReadDirSizes(dir)
}
