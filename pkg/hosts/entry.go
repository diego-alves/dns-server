package hosts

import "strings"

type Entry struct {
	Source    string
	IpAddress string
	Hostnames []string
}

func NewEntry(ip string, hosts []string, source string) *Entry {
	entry := new(Entry)
	entry.IpAddress = ip
	entry.Hostnames = hosts
	entry.Source = source
	return entry
}

func (e *Entry) line() *string {
	if len(e.Hostnames) > 0 {
		line := strings.Join(append([]string{e.IpAddress}, e.Hostnames...), " ")
		return &line
	}
	return nil
}
