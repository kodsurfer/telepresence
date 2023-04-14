//go:build !windows
// +build !windows

package dns

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/telepresenceio/telepresence/v2/pkg/ioutil"
)

type resolveFile struct {
	port        int
	domain      string
	nameservers []string
	search      []string
	options     []string
}

func readResolveFile(fileName string) (*resolveFile, error) {
	fl, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer fl.Close()
	sc := bufio.NewScanner(fl)
	rf := resolveFile{}
	line := 0

	onlyOne := func(key string) error {
		return fmt.Errorf("%q must have a value at %s line %d", key, fileName, line)
	}

	for sc.Scan() {
		line++
		txt := strings.TrimSpace(sc.Text())
		if len(txt) == 0 || strings.HasPrefix(txt, "#") {
			continue
		}
		fields := strings.Fields(txt)
		fc := len(fields)
		if fc == 0 {
			continue
		}
		key := fields[0]
		if fc == 1 {
			return nil, fmt.Errorf("%q must have a value at %s line %d", key, fileName, line)
		}
		value := fields[1]
		switch key {
		case "port":
			if fc != 2 {
				return nil, onlyOne(key)
			}
			rf.port, err = strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("%q is not a valid integer at %s line %d", key, fileName, line)
			}
		case "domain":
			if fc != 2 {
				return nil, onlyOne(key)
			}
			rf.domain = value
		case "nameserver":
			if fc != 2 {
				return nil, onlyOne(key)
			}
			rf.nameservers = append(rf.nameservers, value)
		case "search":
			rf.search = fields[1:]
		case "options":
			rf.options = fields[1:]
		default:
			// This reader doesn't do options just yet
			return nil, fmt.Errorf("%q is not a recognized key at %s line %d", key, fileName, line)
		}
	}
	return &rf, nil
}

func (r *resolveFile) String() string {
	var buf strings.Builder
	_, _ = r.WriteTo(&buf)
	return buf.String()
}

func (r *resolveFile) WriteTo(buf io.Writer) (int64, error) {
	n := ioutil.Println(buf, "# Generated by telepresence\n")
	if r.port > 0 {
		n += ioutil.Printf(buf, "port %d\n", r.port)
	}
	if r.domain != "" {
		n += ioutil.Printf(buf, "domain %s\n", r.domain)
	}
	for _, ns := range r.nameservers {
		n += ioutil.Printf(buf, "nameserver %s\n", ns)
	}
	if len(r.search) > 0 {
		n += ioutil.Printf(buf, "search %s\n", strings.Join(r.search, " "))
	}
	if len(r.options) > 0 {
		n += ioutil.Printf(buf, "options %s\n", strings.Join(r.options, " "))
	}
	return int64(n), nil
}
