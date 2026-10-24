// Package setup brings a host to what another has: it scans what a host
// has (toolchains, agent CLIs, git), keeps what to bring as a manifest
// (enowx.toml), and plans the steps that bring a server to it, each
// checking before it changes anything. See docs/setup.md.
package setup

import (
	"bufio"
	"context"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/enowx/enowx/core/agent"
	"github.com/enowx/enowx/proto"
)

// tool is a program a scan looks for, and how it says its version.
type tool struct {
	name string
	args []string
}

// tools are the programs a scan looks for: toolchains, package managers,
// what the steps use, and the agents (from their presets).
var tools = []tool{
	{"git", []string{"--version"}},
	{"curl", []string{"--version"}},
	{"unzip", []string{"-v"}},
	{"make", []string{"--version"}},
	{"gcc", []string{"--version"}},
	{"mise", []string{"--version"}},
	{"node", []string{"--version"}},
	{"bun", []string{"--version"}},
	{"deno", []string{"--version"}},
	{"pnpm", []string{"--version"}},
	{"python3", []string{"--version"}},
	{"go", []string{"version"}},
	{"rustc", []string{"--version"}},
	{"java", []string{"-version"}},
	{"gh", []string{"--version"}},
	{"docker", []string{"--version"}},
	{"xpra", []string{"--version"}},
	// What runs a step as root: sudo, else doas (Alpine's).
	{"sudo", []string{"-V"}},
	{"doas", nil},
}

// scanTimeout bounds asking one program its version.
const scanTimeout = 5 * time.Second

var versionPattern = regexp.MustCompile(`\d+\.\d+(\.\d+)?`)

// Scan finds what this host has. Programs are looked for on PATH, which
// the daemon takes from a login shell, so version managers' shims count.
func Scan() proto.SetupScan {
	s := proto.SetupScan{OS: runtime.GOOS, Arch: runtime.GOARCH, Tools: map[string]string{}, Root: os.Geteuid() == 0}
	s.Distro, s.Release = osRelease()
	for _, pm := range []string{"apt-get", "dnf", "zypper", "pacman", "apk", "brew"} {
		if _, err := exec.LookPath(pm); err == nil {
			s.Packages = pm
			break
		}
	}
	all := append([]tool(nil), tools...)
	for _, p := range agent.Presets {
		if agent.Base(p.ID) == p.ID {
			all = append(all, tool{p.Command, []string{"--version"}})
		}
	}
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, t := range all {
		path, err := exec.LookPath(t.name)
		if err != nil {
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			v := version(path, t.args)
			mu.Lock()
			s.Tools[t.name] = v
			mu.Unlock()
		}()
	}
	wg.Wait()
	if _, ok := s.Tools["git"]; ok {
		s.GitName = output("git", "config", "--global", "user.name")
		s.GitEmail = output("git", "config", "--global", "user.email")
	}
	if _, ok := s.Tools["gh"]; ok {
		s.GitHub = scanCommand(context.Background(), "gh", "auth", "status").Run() == nil
	}
	return s
}

// version runs a program to read the version it says.
func version(path string, args []string) string {
	ctx, cancel := context.WithTimeout(context.Background(), scanTimeout)
	defer cancel()
	out, _ := scanCommand(ctx, path, args...).CombinedOutput()
	return versionPattern.FindString(string(out))
}

func output(name string, args ...string) string {
	ctx, cancel := context.WithTimeout(context.Background(), scanTimeout)
	defer cancel()
	out, _ := scanCommand(ctx, name, args...).Output()
	return strings.TrimSpace(string(out))
}

// osRelease reads the distribution and its version on Linux.
func osRelease() (id, release string) {
	f, err := os.Open("/etc/os-release")
	if err != nil {
		return "", ""
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		k, v, ok := strings.Cut(sc.Text(), "=")
		if !ok {
			continue
		}
		v = strings.Trim(v, `"`)
		switch k {
		case "ID":
			id = v
		case "VERSION_ID":
			release = v
		}
	}
	return id, release
}
