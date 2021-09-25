package hosts

import (
	"log"
	"os"
	"strings"
	"sync"
)

type HostsFile struct {
	mu sync.Mutex
}

type ChangeFunc func(line string) *string
type AppendFunc func() *string

func (h *HostsFile) change(foo ChangeFunc, app AppendFunc) {
	h.mu.Lock() // Threading safe on modify host file
	defer h.mu.Unlock()

	file, err := os.ReadFile("hosts")
	if err != nil {
		log.Fatal(err)
	}

	i := 0
	lines := strings.Split(string(file), "\n")
	for _, line := range lines {
		new_line := foo(line)
		if new_line != nil {
			lines[i] = *new_line
			i++
		}
	}
	lines = lines[:i]

	last := app()
	if last != nil {
		lines = append(lines, *last)
	}

	out := strings.Join(lines, "\n")
	err = os.WriteFile("hosts", []byte(out), 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func (f *HostsFile) Add(entry []string) {
	end := true
	new_line := strings.Join(entry, " ")
	f.change(func(line string) *string {
		if strings.HasPrefix(line, entry[0]+" ") {
			end = false
			return &new_line
		} else {
			return &line
		}
	}, func() *string {
		if end {
			return &new_line
		} else {
			return nil
		}
	})
}

func (f *HostsFile) Remove(ip string) {
	f.change(func(line string) *string {
		if strings.HasPrefix(line, ip+" ") {
			return nil
		}
		return &line
	}, func() *string {
		return nil
	})
}
