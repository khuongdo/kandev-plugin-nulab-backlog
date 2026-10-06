// Package pkgverify checks a built plugin package (BR5.2). It replaces
// Kandev's plugin-package-verify, which does not exist at the pinned
// v0.96.0. The checksum rules mirror Kandev's internal pkgtar package.
package pkgverify

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"maps"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	checksumsName    = "checksums.txt"
	checksumsSigName = "checksums.txt.sig"
	// maxEntry bounds one archive entry; the executables are about 16 MiB.
	maxEntry = 128 << 20
)

// executables is the exact runtime.executables map (BR5.1).
var executables = map[string]string{
	"linux-amd64":   "server/plugin-linux-amd64",
	"linux-arm64":   "server/plugin-linux-arm64",
	"darwin-amd64":  "server/plugin-darwin-amd64",
	"darwin-arm64":  "server/plugin-darwin-arm64",
	"windows-amd64": "server/plugin-windows-amd64.exe",
}

// nulabAssetURL matches a URL on a Nulab or Backlog domain. The logo is
// inline SVG, so the bundle must never load an asset from these hosts
// (NFR3.10). Bare host names such as the input placeholder are allowed.
var nulabAssetURL = regexp.MustCompile(`(?i)(?:https?:)?//(?:[a-z0-9-]+\.)*(?:nulab|nulab-inc|backlog|backlogtool)\.(?:com|jp)\b`)

// Expect is the identity the manifest must carry.
type Expect struct {
	ID      string
	Version string
}

// Verify fails unless the archive matches its dist checksums line, every file
// matches the in-archive checksums.txt with nothing unlisted or missing, the
// required files are present, the UI bundle references no Nulab or Backlog
// asset URL, and the manifest has the expected identity and exactly the five
// executables.
func Verify(archivePath, distChecksumsPath string, want Expect) error {
	raw, err := os.ReadFile(archivePath) //nolint:gosec // G304: the path is the operator's own build output
	if err != nil {
		return fmt.Errorf("read package: %w", err)
	}
	if err := checkDistChecksum(raw, filepath.Base(archivePath), distChecksumsPath); err != nil {
		return err
	}
	files, err := readArchive(raw)
	if err != nil {
		return err
	}
	if err := checkContents(files); err != nil {
		return err
	}
	for _, name := range append([]string{"manifest.yaml", "ui/bundle.js"}, values(executables)...) {
		if _, ok := files[name]; !ok {
			return fmt.Errorf("missing required file: %s", name)
		}
	}
	if nulabAssetURL.Match(files["ui/bundle.js"]) {
		return errors.New("ui/bundle.js references a Nulab or Backlog asset URL")
	}
	return checkManifest(files["manifest.yaml"], want)
}

func checkDistChecksum(raw []byte, base, distChecksumsPath string) error {
	data, err := os.ReadFile(distChecksumsPath) //nolint:gosec // G304: the path is the operator's own build output
	if err != nil {
		return fmt.Errorf("read %s: %w", distChecksumsPath, err)
	}
	sums, err := parseChecksums(data)
	if err != nil {
		return err
	}
	wantSum, ok := sums[base]
	if !ok {
		return fmt.Errorf("%s is not listed in %s", base, distChecksumsPath)
	}
	if hexSum(raw) != wantSum {
		return fmt.Errorf("package checksum mismatch: %s", base)
	}
	return nil
}

// readArchive returns every regular file in the .tar.gz by clean path.
func readArchive(raw []byte) (map[string][]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("open package: %w", err)
	}
	defer func() { _ = gz.Close() }()
	tr := tar.NewReader(gz)
	files := map[string][]byte{}
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return files, nil
		}
		if err != nil {
			return nil, fmt.Errorf("read package: %w", err)
		}
		if hdr.Typeflag == tar.TypeDir {
			continue
		}
		name, err := cleanPath(hdr.Name)
		if err != nil {
			return nil, err
		}
		if hdr.Typeflag != tar.TypeReg {
			return nil, fmt.Errorf("unsupported entry type: %s", name)
		}
		data, err := io.ReadAll(io.LimitReader(tr, maxEntry+1))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", name, err)
		}
		if len(data) > maxEntry {
			return nil, fmt.Errorf("entry too large: %s", name)
		}
		files[name] = data
	}
}

func cleanPath(name string) (string, error) {
	clean := path.Clean(strings.TrimPrefix(name, "./"))
	if clean == "." || path.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, "../") {
		return "", fmt.Errorf("unsafe path in package: %q", name)
	}
	return clean, nil
}

// checkContents applies the in-archive checksums.txt written by plugin-pack.
func checkContents(files map[string][]byte) error {
	data, ok := files[checksumsName]
	if !ok {
		return errors.New("missing checksums.txt in package")
	}
	sums, err := parseChecksums(data)
	if err != nil {
		return err
	}
	for name, content := range files {
		if name == checksumsName || name == checksumsSigName {
			continue
		}
		wantSum, listed := sums[name]
		if !listed {
			return fmt.Errorf("unlisted file: %s", name)
		}
		if hexSum(content) != wantSum {
			return fmt.Errorf("checksum mismatch: %s", name)
		}
	}
	for name := range sums {
		if _, ok := files[name]; !ok {
			return fmt.Errorf("listed file missing: %s", name)
		}
	}
	return nil
}

// parseChecksums reads "<sha256>  <path>" lines (sha256sum format).
func parseChecksums(data []byte) (map[string]string, error) {
	sums := map[string]string{}
	sc := bufio.NewScanner(bytes.NewReader(data))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			return nil, fmt.Errorf("malformed checksums.txt line: %q", line)
		}
		sums[strings.TrimPrefix(fields[1], "*")] = strings.ToLower(fields[0])
	}
	return sums, sc.Err()
}

func checkManifest(data []byte, want Expect) error {
	var m struct {
		ID      string `yaml:"id"`
		Version string `yaml:"version"`
		Runtime struct {
			Executables map[string]string `yaml:"executables"`
		} `yaml:"runtime"`
	}
	if err := yaml.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("invalid manifest: %w", err)
	}
	if m.ID != want.ID || m.Version != want.Version {
		return fmt.Errorf("manifest is %s@%s, expected %s@%s", m.ID, m.Version, want.ID, want.Version)
	}
	if !maps.Equal(m.Runtime.Executables, executables) {
		return errors.New("manifest must list exactly the 5 executables of BR5.1")
	}
	return nil
}

func hexSum(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

func values(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out
}

// Run is the verifypkg command. It returns 0 on success, 1 when verification
// fails, and 2 on a usage error.
func Run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("verifypkg", flag.ContinueOnError)
	fs.SetOutput(stderr)
	archive := fs.String("archive", "", "path to the plugin package (.tar.gz)")
	sums := fs.String("checksums", "", "path to dist/checksums.txt")
	id := fs.String("expected-id", "", "expected manifest id")
	version := fs.String("expected-version", "", "expected manifest version")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *archive == "" || *sums == "" || *id == "" || *version == "" {
		_, _ = fmt.Fprintln(stderr, "verifypkg: -archive, -checksums, -expected-id and -expected-version are required")
		return 2
	}
	if err := Verify(*archive, *sums, Expect{ID: *id, Version: *version}); err != nil {
		_, _ = fmt.Fprintln(stderr, "verifypkg:", err)
		return 1
	}
	_, _ = fmt.Fprintf(stdout, "verifypkg: OK %s (%s@%s)\n", *archive, *id, *version)
	return 0
}
