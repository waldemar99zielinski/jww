package migrator

import "regexp"

var filenamePattern = regexp.MustCompile(`^(\d{5})_([a-z0-9_]+)\.(up|down)\.sql$`)

// parseFilename splits "<version>_<name>.<up|down>.sql" into its parts. ok is
// false for any file that doesn't match this shape (e.g. README.md).
func parseFilename(name string) (version, migrationName, direction string, ok bool) {
	match := filenamePattern.FindStringSubmatch(name)
	if match == nil {
		return "", "", "", false
	}
	return match[1], match[2], match[3], true
}
