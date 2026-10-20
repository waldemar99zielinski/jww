package migrator

import (
	"fmt"
	"io/fs"
	"sort"
)

// Load reads every "<version>_<name>.{up,down}.sql" pair from fsys and
// returns them sorted by version. It errors if any version is missing its up
// or down half.
func Load(fsys fs.FS) ([]Migration, error) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("reading migrations dir: %w", err)
	}

	type half struct{ up, down string }
	halves := make(map[string]*half)
	names := make(map[string]string)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		version, name, direction, ok := parseFilename(entry.Name())
		if !ok {
			continue // not a migration file, e.g. README.md
		}

		if existing, ok := names[version]; ok && existing != name {
			return nil, fmt.Errorf("migrator: version %s has two different names: %q and %q", version, existing, name)
		}
		names[version] = name

		content, err := fs.ReadFile(fsys, entry.Name())
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", entry.Name(), err)
		}

		h, ok := halves[version]
		if !ok {
			h = &half{}
			halves[version] = h
		}
		if direction == "up" {
			h.up = string(content)
		} else {
			h.down = string(content)
		}
	}

	versions := make([]string, 0, len(halves))
	for v := range halves {
		versions = append(versions, v)
	}
	sort.Strings(versions)

	migrations := make([]Migration, 0, len(versions))
	for _, v := range versions {
		h := halves[v]
		if h.up == "" {
			return nil, fmt.Errorf("migrator: version %s (%s) is missing its .up.sql file", v, names[v])
		}
		if h.down == "" {
			return nil, fmt.Errorf("migrator: version %s (%s) is missing its .down.sql file", v, names[v])
		}
		migrations = append(migrations, Migration{
			Version:  v,
			Name:     names[v],
			Up:       h.up,
			Down:     h.down,
			Checksum: checksum(h.up, h.down),
		})
	}
	return migrations, nil
}
