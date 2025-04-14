// SPDX-License-Identifier: Apache-2.0

package archiver

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// isPathWithinBoundary checks if a path is within a directory.
func isPathWithinBoundary(path, dir string) bool {
	path = filepath.Clean(path)
	dir = filepath.Clean(dir)

	return strings.HasPrefix(path, dir+string(os.PathSeparator)) || path == dir
}

// filterRedundantPaths removes paths that are already covered by other paths in the list.
func filterRedundantPaths(paths []string) ([]string, error) {
	if len(paths) <= 1 {
		return paths, nil
	}

	type pathInfo struct {
		original string // original path as provided
		abs      string // absolute path for comparison
		isDir    bool   // whether it's a directory
	}

	// collect information about all paths
	infos := make([]pathInfo, 0, len(paths))

	for _, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path for %s: %w", path, err)
		}

		fi, err := os.Stat(absPath)
		if err != nil {
			return nil, fmt.Errorf("failed to stat %s: %w", absPath, err)
		}

		info := pathInfo{
			original: path,
			abs:      absPath,
			isDir:    fi.IsDir(),
		}

		// ensure directory paths end with separator for proper prefix matching
		if info.isDir && !strings.HasSuffix(info.abs, string(filepath.Separator)) {
			info.abs += string(filepath.Separator)
		}

		infos = append(infos, info)
	}

	// sort by path length (shortest first) so parent directories come before children
	sort.Slice(infos, func(i, j int) bool {
		return len(infos[i].abs) < len(infos[j].abs)
	})

	// filter out redundant paths
	var result []string

	for i, info := range infos {
		redundant := false

		// check if this path is covered by any previous path in the sorted list
		for j := range i {
			prevInfo := infos[j]

			// if the previous path is a directory and the current path starts with it,
			// then the current path is redundant
			if prevInfo.isDir && strings.HasPrefix(info.abs, prevInfo.abs) {
				redundant = true
				break
			}
		}

		if !redundant {
			result = append(result, info.original)
		}
	}

	return result, nil
}
