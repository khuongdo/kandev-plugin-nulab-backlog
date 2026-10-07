package pkgverify

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const goodManifest = `id: "nulab-backlog"
version: "0.0.1"
runtime:
  type: binary
  executables:
    linux-amd64: server/plugin-linux-amd64
    linux-arm64: server/plugin-linux-arm64
    darwin-amd64: server/plugin-darwin-amd64
    darwin-arm64: server/plugin-darwin-arm64
`

var want = Expect{ID: "nulab-backlog", Version: "0.0.1"}

// goodFiles is a package that passes every check.
func goodFiles() map[string]string {
	return map[string]string{
		"manifest.yaml":              goodManifest,
		"ui/bundle.js":               "export {};",
		"server/plugin-linux-amd64":  "bin-la",
		"server/plugin-linux-arm64":  "bin-lr",
		"server/plugin-darwin-amd64": "bin-da",
		"server/plugin-darwin-arm64": "bin-dr",
	}
}

func sum(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// checksumsFor writes plugin-pack's format: "<sha256>  <path>" per line.
func checksumsFor(files map[string]string) string {
	names := make([]string, 0, len(files))
	for n := range files {
		names = append(names, n)
	}
	sort.Strings(names)
	var b strings.Builder
	for _, n := range names {
		fmt.Fprintf(&b, "%s  %s\n", sum(files[n]), n)
	}
	return b.String()
}

// pkg builds dist/<name> and dist/checksums.txt. mutate runs on the archive
// entries after the in-archive checksums.txt is computed.
type pkg struct {
	files    map[string]string // entries checksummed by checksums.txt
	extra    map[string]string // entries added after checksumming
	skipSums bool              // omit the in-archive checksums.txt
	mutate   func(entries map[string]string)
}

func (p pkg) write(t *testing.T) (archive, distSums string) {
	t.Helper()
	entries := map[string]string{}
	for k, v := range p.files {
		entries[k] = v
	}
	if !p.skipSums {
		entries["checksums.txt"] = checksumsFor(p.files)
	}
	for k, v := range p.extra {
		entries[k] = v
	}
	if p.mutate != nil {
		p.mutate(entries)
	}
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	names := make([]string, 0, len(entries))
	for n := range entries {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		require.NoError(t, tw.WriteHeader(&tar.Header{Name: n, Mode: 0o644, Size: int64(len(entries[n])), Typeflag: tar.TypeReg}))
		_, err := tw.Write([]byte(entries[n]))
		require.NoError(t, err)
	}
	require.NoError(t, tw.Close())
	require.NoError(t, gz.Close())

	dir := t.TempDir()
	archive = filepath.Join(dir, "nulab-backlog-0.0.1.tar.gz")
	require.NoError(t, os.WriteFile(archive, buf.Bytes(), 0o600))
	distSums = filepath.Join(dir, "checksums.txt")
	line := fmt.Sprintf("%s  %s\n", sum(buf.String()), filepath.Base(archive))
	require.NoError(t, os.WriteFile(distSums, []byte(line), 0o600))
	return archive, distSums
}

func TestVerifyAcceptsAGoodPackage(t *testing.T) {
	archive, sums := pkg{files: goodFiles()}.write(t)
	require.NoError(t, Verify(archive, sums, want))
}

func TestVerifyRejects(t *testing.T) {
	without := func(name string) map[string]string {
		f := goodFiles()
		delete(f, name)
		return f
	}
	withWindowsExe := func() map[string]string {
		f := goodFiles()
		f["server/plugin-windows-amd64.exe"] = "bin-wa"
		return f
	}
	withManifest := func(m string) map[string]string {
		f := goodFiles()
		f["manifest.yaml"] = m
		return f
	}
	cases := []struct {
		name    string
		pkg     pkg
		expect  Expect
		errText string
	}{
		{name: "a file whose content differs from its checksum", pkg: pkg{files: goodFiles(), mutate: func(e map[string]string) { e["ui/bundle.js"] += "// tampered" }}, errText: "checksum mismatch: ui/bundle.js"},
		{name: "a listed file that is missing", pkg: pkg{files: goodFiles(), mutate: func(e map[string]string) { delete(e, "ui/bundle.js") }}, errText: "listed file missing: ui/bundle.js"},
		{name: "an unlisted file", pkg: pkg{files: goodFiles(), extra: map[string]string{"server/backdoor": "x"}}, errText: "unlisted file: server/backdoor"},
		{name: "no in-archive checksums.txt", pkg: pkg{files: goodFiles(), skipSums: true}, errText: "missing checksums.txt"},
		{name: "a malformed checksums line", pkg: pkg{files: goodFiles(), mutate: func(e map[string]string) { e["checksums.txt"] += "garbage\n" }}, errText: "malformed checksums.txt line"},
		{name: "a path that escapes the package", pkg: pkg{files: goodFiles(), extra: map[string]string{"../evil": "x"}}, errText: "unsafe path"},
		{name: "missing manifest.yaml", pkg: pkg{files: without("manifest.yaml")}, errText: "missing required file: manifest.yaml"},
		{name: "missing ui/bundle.js", pkg: pkg{files: without("ui/bundle.js")}, errText: "missing required file: ui/bundle.js"},
		{name: "a missing executable", pkg: pkg{files: without("server/plugin-darwin-arm64")}, errText: "missing required file: server/plugin-darwin-arm64"},
		{name: "a wrong plugin id", pkg: pkg{files: goodFiles()}, expect: Expect{ID: "other", Version: "0.0.1"}, errText: `manifest is nulab-backlog@0.0.1, expected other@0.0.1`},
		{name: "a wrong version", pkg: pkg{files: goodFiles()}, expect: Expect{ID: "nulab-backlog", Version: "9.9.9"}, errText: "expected nulab-backlog@9.9.9"},
		{name: "a manifest missing an executable", pkg: pkg{files: withManifest(strings.Replace(goodManifest, "    darwin-arm64: server/plugin-darwin-arm64\n", "", 1))}, errText: "manifest must list exactly the four executables"},
		// Regression for "Plugin install failed: 502": Windows is no longer packaged.
		{name: "a manifest that still lists windows-amd64", pkg: pkg{files: withManifest(goodManifest + "    windows-amd64: server/plugin-windows-amd64.exe\n")}, errText: "manifest must list exactly the four executables"},
		{name: "a package containing the Windows executable", pkg: pkg{files: withWindowsExe()}, errText: "unexpected executable: server/plugin-windows-amd64.exe"},
		{name: "an invalid manifest", pkg: pkg{files: withManifest("id: [")}, errText: "invalid manifest"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			archive, sums := tc.pkg.write(t)
			expect := tc.expect
			if expect == (Expect{}) {
				expect = want
			}
			err := Verify(archive, sums, expect)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.errText)
		})
	}
}

func TestVerifyRejectsAnArchiveThatDiffersFromDistChecksums(t *testing.T) {
	archive, sums := pkg{files: goodFiles()}.write(t)
	require.NoError(t, os.WriteFile(sums, []byte(sum("other")+"  nulab-backlog-0.0.1.tar.gz\n"), 0o600))
	require.ErrorContains(t, Verify(archive, sums, want), "package checksum mismatch")
}

func TestVerifyRejectsAnArchiveMissingFromDistChecksums(t *testing.T) {
	archive, sums := pkg{files: goodFiles()}.write(t)
	require.NoError(t, os.WriteFile(sums, []byte(sum("x")+"  another.tar.gz\n"), 0o600))
	require.ErrorContains(t, Verify(archive, sums, want), "not listed in")
}

func TestVerifyRejectsAFileThatIsNotGzip(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "p.tar.gz")
	require.NoError(t, os.WriteFile(archive, []byte("not gzip"), 0o600))
	sums := filepath.Join(dir, "checksums.txt")
	require.NoError(t, os.WriteFile(sums, []byte(sum("not gzip")+"  p.tar.gz\n"), 0o600))
	require.ErrorContains(t, Verify(archive, sums, want), "open package")
}

func TestVerifyReportsMissingInputFiles(t *testing.T) {
	dir := t.TempDir()
	require.ErrorContains(t, Verify(filepath.Join(dir, "none.tar.gz"), filepath.Join(dir, "checksums.txt"), want), "read")
}

func TestRunExitCodes(t *testing.T) {
	archive, sums := pkg{files: goodFiles()}.write(t)
	var out, errOut bytes.Buffer
	args := []string{"-archive", archive, "-checksums", sums, "-expected-id", "nulab-backlog", "-expected-version", "0.0.1"}
	require.Equal(t, 0, Run(args, &out, &errOut), errOut.String())
	require.Contains(t, out.String(), "OK")

	errOut.Reset()
	require.Equal(t, 1, Run([]string{"-archive", archive, "-checksums", sums, "-expected-id", "x", "-expected-version", "0.0.1"}, &out, &errOut))
	require.Contains(t, errOut.String(), "verifypkg:")

	require.Equal(t, 2, Run([]string{"-archive", archive}, &out, &errOut), "missing flags is a usage error")
	require.Equal(t, 2, Run([]string{"-nope"}, &out, &errOut))
}

func TestVerifyRejectsABundleLoadingANulabOrBacklogAsset(t *testing.T) {
	for _, bundle := range []string{
		`const logo = "https://nulab.com/images/media-assets/logos/backlog.png";`,
		`img.src = "//cdn.backlog.jp/logo.svg";`,
		`fetch("HTTPS://Example.BacklogTool.com/icon.png")`,
		`url(http://nulab-inc.com/x.svg)`,
		`"https://backlog.com/favicon.ico"`,
	} {
		t.Run(bundle, func(t *testing.T) {
			files := goodFiles()
			files["ui/bundle.js"] = bundle
			archive, sums := pkg{files: files}.write(t)
			err := Verify(archive, sums, want)
			require.Error(t, err)
			require.Contains(t, err.Error(), "ui/bundle.js references a Nulab or Backlog asset URL") // NFR3.10
		})
	}
}

func TestVerifyAcceptsABundleNamingBacklogHostsWithoutURLs(t *testing.T) {
	files := goodFiles()
	files["ui/bundle.js"] = `const placeholder = "myteam.backlog.com"; const help = "myteam.backlog.jp or myteam.backlogtool.com";`
	archive, sums := pkg{files: files}.write(t)
	require.NoError(t, Verify(archive, sums, want))
}
