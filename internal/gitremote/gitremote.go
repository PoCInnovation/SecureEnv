// Package gitremote derives a default project name from a git repository.
package gitremote

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"regexp"
	"strings"
)

// scpLike matches the scp-like syntax used by ssh remotes: user@host:path.
var scpLike = regexp.MustCompile(`^[\w.-]+@[\w.-]+:(.*)$`)

var invalidChars = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

// OriginURL returns the URL of the "origin" remote of the repository
// containing dir.
func OriginURL(ctx context.Context, dir string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "config", "--get", "remote.origin.url")
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			return "", fmt.Errorf("git: %s", msg)
		}
		return "", errors.New("git: no origin remote found")
	}
	return strings.TrimSpace(string(out)), nil
}

// ProjectName turns a remote URL into a project name made of the repository
// path segments joined with "_", e.g. git@github.com:owner/repo.git becomes
// owner_repo.
func ProjectName(remote string) (string, error) {
	repoPath, err := repositoryPath(strings.TrimSpace(remote))
	if err != nil {
		return "", err
	}

	repoPath = strings.TrimSuffix(strings.Trim(repoPath, "/"), ".git")
	var segments []string
	for _, segment := range strings.Split(repoPath, "/") {
		if cleaned := strings.Trim(invalidChars.ReplaceAllString(segment, "-"), "-."); cleaned != "" {
			segments = append(segments, cleaned)
		}
	}
	if len(segments) == 0 {
		return "", fmt.Errorf("git: cannot derive a project name from %q", remote)
	}
	return strings.Join(segments, "_"), nil
}

func repositoryPath(remote string) (string, error) {
	if match := scpLike.FindStringSubmatch(remote); match != nil {
		return match[1], nil
	}
	if strings.Contains(remote, "://") {
		parsed, err := url.Parse(remote)
		if err != nil {
			return "", fmt.Errorf("git: invalid remote URL: %w", err)
		}
		return parsed.Path, nil
	}
	return remote, nil
}
