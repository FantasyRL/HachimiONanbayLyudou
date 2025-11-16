package application

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ===== 辅助：构造目录树（纯 Go, depth/ignore 简化） =====
func buildTreeText(root string, maxDepth int, ignores []string) (string, error) {
	root = filepath.Clean(root)
	info, err := os.Stat(root)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", root)
	}
	var lines []string
	prefix := ""
	var walk func(string, int) error
	walk = func(dir string, depth int) error {
		if depth > maxDepth {
			return nil
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return err
		}
		// 过滤忽略
		filter := func(name string) bool {
			for _, g := range ignores {
				ok, _ := filepath.Match(g, name)
				if ok {
					return true
				}
			}
			return false
		}
		for i, e := range entries {
			name := e.Name()
			if filter(name) {
				continue
			}
			isLast := i == len(entries)-1
			conn := "├── "
			nextPrefix := prefix + "│   "
			if isLast {
				conn = "└── "
				nextPrefix = prefix + "    "
			}
			lines = append(lines, prefix+conn+name)
			if e.IsDir() {
				old := prefix
				prefix = nextPrefix
				_ = walk(filepath.Join(dir, name), depth+1)
				prefix = old
			}
		}
		return nil
	}
	lines = append(lines, filepath.Base(root))
	if err := walk(root, 1); err != nil {
		return "", err
	}
	return strings.Join(lines, "\n"), nil
}
