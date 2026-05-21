package pathutil

// DirOf returns the directory component of path.
// Returns "." if path contains no separator.
func DirOf(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[:i]
		}
	}

	return "."
}
