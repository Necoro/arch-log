package arch

import "github.com/Necoro/arch-log/pkg/entries"

func GetEntries(pkg string) ([]entries.Entry, error) {
	return nil, entries.ErrNotFound
}
