package land

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/Harrison-Blair/hive/internal/lib/gitx"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/run"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// landPR pushes the task branch and opens a pull request.
func landPR(ctx context.Context, registry *store.Store, logger *slog.Logger, home string, options Options, task records.Task, project records.Project) output.Envelope {
	repo := project.Path
	base := project.Base
	branch := task.Worktree.Branch
	fail := func(phase, message string) output.Envelope {
		logFailure(logger, project.ID, task.ID, phase, message)
		return failedEnvelope(phase, message)
	}
	retry := task.Delivery != nil && task.Delivery.Commit != ""
	commit := ""
	pushed := false
	prURL := ""
	effects := []output.Effect{}
	if retry {
		commit = task.Delivery.Commit
		pushed = task.Delivery.Pushed
		prURL = task.Delivery.PRURL
	}

	if !pushed {
		pushCtx, cancel := context.WithTimeout(ctx, bound(options.PushTimeout, pushTimeout))
		pushResult, pushErr := gitx.Push(pushCtx, repo, "origin", "-u", branch)
		cancel()
		if pushErr != nil || pushResult.ExitCode != 0 {
			if commit == "" {
				commit, _ = gitx.Tip(ctx, repo, branch)
			}
			var message string
			if pushErr != nil {
				message = "push failed: " + oneLine(pushErr.Error())
			} else {
				message = "push failed: " + gitOutput(pushResult)
			}
			delivery := records.Delivery{Mode: "pr", Target: base, Grant: options.Grant, Commit: commit, Pushed: false, Evidence: options.Evidence, At: now()}
			return deliveryEnvelope(registry, logger, task, project, delivery, "partial", "push", message, "", effects)
		}
		var err error
		commit, err = gitx.Tip(ctx, repo, branch)
		if err != nil {
			return fail("push", "read branch tip: "+oneLine(err.Error()))
		}
		pushed = true
		effects = append(effects, output.Effect{Resource: "push", Path: branch, Action: "pushed"})
		logEffect(logger, project.ID, task.ID, "push", branch, "pushed")
	}

	if prURL == "" && retry {
		url, found, _ := prView(ctx, repo, branch)
		if found {
			prURL = url
		}
	}
	if prURL == "" {
		body, err := briefBody(home, task.ID, options.Evidence)
		if err != nil {
			return fail("brief", err.Error())
		}
		bodyFile, err := writeBodyFile(home, task.ID, body)
		if err != nil {
			return fail("brief", err.Error())
		}
		defer func() { _ = os.Remove(bodyFile) }()
		args := []string{"pr", "create", "--base", base, "--head", branch, "--title", task.Title, "--body-file", bodyFile}
		if options.Draft {
			args = append(args, "--draft")
		}
		ghCtx, cancel := context.WithTimeout(ctx, ghTimeout)
		ghResult, ghErr := run.Command(ghCtx, repo, "gh", args...)
		cancel()
		if ghErr != nil || ghResult.ExitCode != 0 {
			var message string
			if ghErr != nil {
				message = "gh pr create failed: " + oneLine(ghErr.Error())
			} else {
				message = "gh pr create failed: " + gitOutput(ghResult)
			}
			delivery := records.Delivery{Mode: "pr", Target: base, Grant: options.Grant, Commit: commit, Pushed: pushed, Evidence: options.Evidence, At: now()}
			return deliveryEnvelope(registry, logger, task, project, delivery, "partial", "pull request", message, "", effects)
		}
		prURL = firstURL(ghResult.Stdout)
		effects = append(effects, output.Effect{Resource: "pull request", Path: prURL, Action: "created"})
		logEffect(logger, project.ID, task.ID, "pull request", prURL, "created")
	}

	delivery := records.Delivery{Mode: "pr", Target: base, Grant: options.Grant, Commit: commit, Pushed: pushed, PRURL: prURL, Evidence: options.Evidence, At: now()}
	return deliveryEnvelope(registry, logger, task, project, delivery, "ok", "complete", "", "", effects)
}

// prView reads the URL of an existing pull request for the branch.
func prView(ctx context.Context, repo, branch string) (string, bool, error) {
	result, err := run.Command(ctx, repo, "gh", "pr", "view", branch, "--json", "url")
	if err != nil {
		return "", false, err
	}
	if result.ExitCode != 0 {
		return "", false, nil
	}
	var payload struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal([]byte(result.Stdout), &payload); err != nil {
		return "", false, nil
	}
	if payload.URL == "" {
		return "", false, nil
	}
	return payload.URL, true, nil
}

// briefBody returns the pull request body from the brief first paragraph.
func briefBody(home, taskID, evidence string) (string, error) {
	data, err := os.ReadFile(briefPath(home, taskID))
	if err != nil {
		return "", fmt.Errorf("read brief: %s", oneLine(err.Error()))
	}
	body := firstParagraph(string(data))
	if evidence != "" {
		body += "\n\nEvidence: " + evidence
	}
	return body, nil
}

// firstParagraph returns the first block of non-empty lines.
func firstParagraph(text string) string {
	lines := strings.Split(text, "\n")
	paragraph := []string{}
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			if len(paragraph) > 0 {
				break
			}
			continue
		}
		paragraph = append(paragraph, strings.TrimRight(line, "\r"))
	}
	return strings.TrimSpace(strings.Join(paragraph, "\n"))
}

// writeBodyFile writes the body below the hive tmp directory.
func writeBodyFile(home, taskID, body string) (string, error) {
	dir := filepath.Join(home, ".hive", "tmp")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("write body: %s", oneLine(err.Error()))
	}
	file, err := os.CreateTemp(dir, "land-"+taskID+"-*.md")
	if err != nil {
		return "", fmt.Errorf("write body: %s", oneLine(err.Error()))
	}
	if _, err := file.WriteString(body); err != nil {
		_ = file.Close()
		_ = os.Remove(file.Name())
		return "", fmt.Errorf("write body: %s", oneLine(err.Error()))
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(file.Name())
		return "", fmt.Errorf("write body: %s", oneLine(err.Error()))
	}
	return file.Name(), nil
}

// firstURL returns the first http token of one gh output.
func firstURL(text string) string {
	for _, field := range strings.Fields(text) {
		if strings.HasPrefix(field, "http") {
			return field
		}
	}
	return strings.TrimSpace(text)
}
