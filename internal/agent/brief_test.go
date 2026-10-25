package agent

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/guide"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

func TestPiSpawnDeliversInstructionsInSafeOrder(t *testing.T) {
	_, fake, _ := spawnProfileTask(t, "pi")
	args := startArgs(t, fake)
	if !reflect.DeepEqual(args, []string{"--model", "m"}) {
		t.Errorf("Pi start arguments must contain only the model; got %d arguments", len(args))
	}
	for index, arg := range args {
		if strings.ContainsAny(arg, "\r\n") {
			t.Errorf("Pi start argument %d contains multiline text", index)
		}
	}
	prompt := promptText(t, fake)
	profile := strings.TrimSpace(readProfile(t, "builder"))
	guideText := strings.TrimSpace(guide.Text())
	reminder := "Read `profiles/builder.md` and run `hive guide` after a context reset."
	wantPrefix := "⬢ " + profile + "\n\n" + guideText + "\n\n" + reminder + "\n\nYou are builder-1."
	if !strings.Contains(prompt, profile) {
		t.Error("Pi first prompt misses the builder profile")
	}
	if !strings.Contains(prompt, guideText) {
		t.Error("Pi first prompt misses the guide")
	}
	if !strings.HasPrefix(prompt, wantPrefix) {
		t.Error("Pi first prompt must start with the profile, guide, reset reminder, and facts in that order")
	}
}

func TestFirstPromptStartsWithHiveMarker(t *testing.T) {
	got := firstPrompt(records.Agent{Name: "builder-1"}, nil, "")
	if !strings.HasPrefix(got, "⬢ ") {
		t.Fatalf("prompt = %q; want the hive marker prefix", got)
	}
}

func TestFirstPromptGivesFactsThenSpawnerText(t *testing.T) {
	record := records.Agent{
		Name:     "builder-1",
		Parent:   "manager-hive-g1",
		Project:  "hive",
		Task:     "154-awr-spawn",
		Worktree: "/home/user/projects/hive/worktrees/154-awr-spawn",
	}
	task := &records.Task{ID: "154-awr-spawn", Project: "hive", Feature: "agents-without-roles"}
	got := firstPrompt(record, task, "Do the task.")
	want := "⬢ You are builder-1. Your parent is manager-hive-g1. Your project is hive. " +
		"Your task is 154-awr-spawn. Your worktree is /home/user/projects/hive/worktrees/154-awr-spawn. " +
		"Your brief is .hive/spec/agents-without-roles/tasks/154-awr-spawn/brief.md. " +
		"Your evidence is .hive/spec/agents-without-roles/tasks/154-awr-spawn/evidence/.\n\nDo the task."
	if got != want {
		t.Fatalf("prompt = %q; want %q", got, want)
	}
}

func TestFirstPromptSkipsEmptyFacts(t *testing.T) {
	got := firstPrompt(records.Agent{Name: "general-1"}, nil, "")
	want := "⬢ You are general-1."
	if got != want {
		t.Fatalf("prompt = %q; want %q", got, want)
	}
}

func TestFirstPromptUsesTheTaskBriefWithoutFeature(t *testing.T) {
	record := records.Agent{Name: "builder-1", Task: "1-simple"}
	task := &records.Task{ID: "1-simple"}
	got := firstPrompt(record, task, "")
	if !strings.Contains(got, "Your brief is .hive/state/briefs/1-simple.md.") {
		t.Fatalf("prompt = %q; want the task brief path", got)
	}
	if strings.Contains(got, "Your evidence is") {
		t.Fatalf("prompt = %q; want no evidence path for a task with no feature", got)
	}
}
