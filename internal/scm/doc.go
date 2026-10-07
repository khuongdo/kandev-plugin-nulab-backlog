// Package scm is the provider-neutral source control context for the cloud
// services GitHub, GitLab and Bitbucket (FR1.1, FR1.2): their settings, the
// Backlog project to repository mapping (FR3), pull request lists, saved
// queries and watches (FR4), and links to Kandev tasks and Backlog issues
// (FR5). Backlog Git stays in internal/git (FR1.3). It never imports
// pluginsdk; internal/plugin adapts it to Kandev (C2).
package scm
