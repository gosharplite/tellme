package harness

import (
	"os/exec"
	"strings"
)

// networkCapableClosurePackages must not appear in the binary's dependency
// closure. `net` is deliberately excluded: spf13/pflag (the mandated flag
// library) links `net` for IP flag parsing but never performs I/O, and the
// dialing-symbol check below covers real network capability. `net/http` is the
// HTTP client — its presence would be a genuine indicator.
var networkCapableClosurePackages = []string{"net/http"}

// networkDialPatterns are symbols whose presence in the linked binary indicates
// the ability to perform network I/O.
var networkDialPatterns = []string{
	"net.Dial",
	"net.(*Dialer).Dial",
	"net.(*Dialer).DialContext",
	"net.Listen",
	"net.ListenPacket",
	"net.LookupHost",
	"net.LookupIP",
	"net.LookupAddr",
	"net.Resolve",
	"net/http.",
	"crypto/tls.(*Conn).Handshake",
}

// NetworkCapabilityViolation reports the first network-capability indicator found
// in the tellme binary — a network-capable package in its dependency closure, or
// a dialing/listening symbol in the linked binary — or "" when none is present.
// It is the build-graph capability guard (research.md Decision 5, backstop
// witness). Using `go tool nm` on the built binary measures real capability
// rather than transitive package presence, which `spf13/pflag` would otherwise
// false-positive on `net`.
func NetworkCapabilityViolation() (string, error) {
	if _, err := BinaryPath(); err != nil {
		return "", err
	}

	closure, err := packageClosure()
	if err != nil {
		return "", err
	}
	for _, pkg := range networkCapableClosurePackages {
		if closure[pkg] {
			return "package " + pkg, nil
		}
	}

	bin, err := BinaryPath()
	if err != nil {
		return "", err
	}
	nm, err := exec.Command("go", "tool", "nm", bin).Output()
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(nm), "\n") {
		for _, pattern := range networkDialPatterns {
			if strings.Contains(line, pattern) {
				return strings.TrimSpace(line), nil
			}
		}
	}
	return "", nil
}

// packageClosure returns the set of packages linked into ./cmd/tellme.
func packageClosure() (map[string]bool, error) {
	cmd := exec.Command("go", "list", "-deps", "./cmd/tellme")
	cmd.Dir = repoRoot()
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	set := make(map[string]bool)
	for _, pkg := range strings.Fields(string(out)) {
		set[pkg] = true
	}
	return set, nil
}
