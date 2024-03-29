package hosts

import (
	"log"
	"os"
	"strings"
	"sync"
)

const HOST_FILE = "/etc/hosts"

type HostsFile struct {
	mu sync.Mutex
}

type ChangeFunc func(line string) *string
type AppendFunc func() *string

func (h *HostsFile) read() []string {
	h.mu.Lock()
	file, err := os.ReadFile(HOST_FILE)
	if err != nil {
		log.Fatal(err)
	}
	return strings.Split(string(file), "\n")
}

func (h *HostsFile) write(lines []string) {
	out := strings.Join(lines, "\n")
	err := os.WriteFile(HOST_FILE, []byte(out), 0644)
	if err != nil {
		log.Fatal(err)
	}
	h.mu.Unlock()
}

func (h *HostsFile) change(alterLine ChangeFunc, lastLine AppendFunc) {
	lines := h.read()

	i := 0
	for _, line := range lines {
		new_line := alterLine(line)
		if new_line != nil { // removes
			lines[i] = *new_line
			i++
		}
	}
	lines = lines[:i]

	last := lastLine()
	if last != nil {
		lines = append(lines, *last)
	}

	h.write(lines)
}

func (f *HostsFile) Add(entry Entry) {
	end := true
	f.change(func(line string) *string {
		if strings.HasPrefix(line, entry.IpAddress+" ") {
			end = false
			return entry.line()
		} else {
			return &line
		}
	}, func() *string {
		if end {
			return entry.line()
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
