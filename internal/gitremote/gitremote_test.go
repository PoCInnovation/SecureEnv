package gitremote_test

import (
	"os"
	"os/exec"
	"testing"

	"github.com/PoCInnovation/SecureEnv/internal/domain"
	"github.com/PoCInnovation/SecureEnv/internal/gitremote"
)

func TestProjectNameFromURL(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"git@github.com:PoCInnovation/SecureEnv.git":       "PoCInnovation_SecureEnv",
		"https://github.com/PoCInnovation/SecureEnv":       "PoCInnovation_SecureEnv",
		"https://github.com/PoCInnovation/SecureEnv.git/":  "PoCInnovation_SecureEnv",
		"ssh://git@gitlab.com:2222/group/sub/repo.git":     "group_sub_repo",
		"https://user:pass@example.com/team/my%20repo.git": "team_my-repo",
		"git@github.com:owner/repo.name.git":               "owner_repo.name",
		"file:///srv/git/project.git":                      "srv_git_project",
		"/home/me/repos/local-repo":                        "home_me_repos_local-repo",
	}
	for url, want := range tests {
		t.Run(url, func(t *testing.T) {
			t.Parallel()

			got, err := gitremote.ProjectName(url)
			if err != nil {
				t.Fatalf("ProjectName(%q) error: %v", url, err)
			}
			if got != want {
				t.Fatalf("ProjectName(%q) = %q, want %q", url, got, want)
			}
			if _, err := domain.NewProjectName(got); err != nil {
				t.Fatalf("derived name is not a valid project name: %v", err)
			}
		})
	}
}

func TestProjectNameFromInvalidURL(t *testing.T) {
	t.Parallel()

	for _, url := range []string{"", "   ", "https://github.com/", "git@github.com:"} {
		if got, err := gitremote.ProjectName(url); err == nil {
			t.Errorf("ProjectName(%q) = %q, want an error", url, got)
		}
	}
}

func TestOriginURL(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init", "-q")

	if _, err := gitremote.OriginURL(t.Context(), dir); err == nil {
		t.Fatal("OriginURL() without origin should fail")
	}

	run("remote", "add", "upstream", "git@github.com:other/fork.git")
	run("remote", "add", "origin", "git@github.com:PoCInnovation/SecureEnv.git")

	got, err := gitremote.OriginURL(t.Context(), dir)
	if err != nil {
		t.Fatalf("OriginURL() error: %v", err)
	}
	if got != "git@github.com:PoCInnovation/SecureEnv.git" {
		t.Fatalf("OriginURL() = %q", got)
	}
}
