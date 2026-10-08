# Requirements Analysis Questions — Plugin install failed: 502

Context: the root cause is confirmed (see `aidlc/spaces/default/codekb/kandev-plugin-nulab-backlog/code-quality-assessment.md#known-issue-plugin-install-502`). The Kandev server (v0.97.0) stops reading a request after 30 seconds (`server.readTimeout`, env `KANDEV_SERVER_READTIMEOUT`). Uploading the 29.5 MB `nulab-backlog-0.4.1.tar.gz` through the web UI over Tailscale takes longer than that, the upload is cut, and `tailscale serve` reports 502. The package is large because it carries 5 server binaries (linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64), each about 15 MB before compression, already stripped.

## Question 1
What should this fix change in the plugin repository?

A. Shrink the package by shipping fewer platform binaries (manifest, Makefile and package verifier change together), and release a patch version
B. Documentation only: README install guide recommends installing by release URL (the server downloads the package itself, no 30 s upload limit) and adds a troubleshooting note for the 502 with `KANDEV_SERVER_READTIMEOUT`
C. Both A and B
D. Nothing in the repository; I only need to get the plugin installed on my server
X. Other (please specify)

[Answer]: C

## Question 2
If the package is shrunk: which server platforms should the package keep? (Kandev rejects a package that has no binary for the server's own platform, so a dropped platform can no longer install the plugin.)

A. Linux only: linux-amd64 and linux-arm64 (about 12 MB package)
B. Linux and macOS: drop only windows-amd64 (about 24 MB package)
C. linux-amd64 only (about 6 MB package)
D. Not applicable (the package is not shrunk)
X. Other (please specify)

[Answer]: B

## Question 3
Your own Kandev server: should I also change its settings, outside the repository, so large uploads stop failing?

A. Yes: set `KANDEV_SERVER_READTIMEOUT=120` in `~/.config/systemd/user/kandev.service` and restart the service
B. No: I will install by release URL (or wait for the smaller package) and leave the server as is
X. Other (please specify)

[Answer]: B

## Follow-up Question 4
Dropping only Windows (Question 2 = B) gives a package of about 24 MB. Your failed upload took more than 30 s for 29.5 MB, so your link was slower than about 1 MB/s; a 24 MB upload needs about 0.8 MB/s to finish in 30 s and may still fail on the same link. Which outcome do you want from the smaller package?

A. Keep Linux + macOS (about 24 MB); browser upload is "better but not guaranteed", and the documented install-by-URL path is the reliable fix
B. Switch to Linux only (about 12 MB) so browser upload works on links down to about 0.4 MB/s; macOS users then cannot install the package
C. Keep Linux + macOS and also improve compression of the package if the Kandev packer allows it (measured, no guarantee)
X. Other (please specify)

[Answer]: A

## Consolidated Summary Confirmation

- Question 1: C — shrink the package by shipping fewer platform binaries and release a patch version, and document install by release URL plus a 502 / `KANDEV_SERVER_READTIMEOUT` troubleshooting note in the README.
- Question 2: B — keep Linux and macOS (linux-amd64, linux-arm64, darwin-amd64, darwin-arm64); drop windows-amd64.
- Question 3: B — do not change the local Kandev server; the user installs by release URL or with the smaller package.
- Follow-up Question 4: A — accept a package of about 24 MB; browser upload is better but not guaranteed on slow links, and install by URL is the documented reliable path.

Does this all look correct before I generate the requirements artifact?

- Looks correct
- Request changes

[Answer]: Looks correct
