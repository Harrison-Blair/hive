package tui

import (
	"fmt"

	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// agentsLegendWave is the end of the legend of agents.md rule 3. It is an
// example, as the lit line of the legend is.
const agentsLegendWave = "   wave B 2/4 = landed tasks"

// projectWave gives the wave progress of the pod title of agents.md rule 2:
// wave W L/T, with the current wave of the active goal of the project, its
// landed tasks, and its tasks. It is empty when the project has no active
// goal or no open wave.
func projectWave(goals []records.Goal, tasks map[string]records.Task, project string) string {
	goal := projectGoal(goals, project)
	if goal == nil || goal.Status != "active" {
		return ""
	}
	for _, wave := range goal.Waves {
		if wave.Status == "done" {
			continue
		}
		landed := 0
		for _, id := range wave.Tasks {
			if task, ok := tasks[id]; ok && taskLanded(task) {
				landed++
			}
		}
		return fmt.Sprintf("wave %s %d/%d", wave.ID, landed, len(wave.Tasks))
	}
	return ""
}

// agentsTitleFacts gives the facts of the title row of agents.md rule 1:
// the project pods, the cards of the home pod, and the h hint while h hides
// a card. The one-pod view shows the pod and esc pods while esc returns to
// the overview (rule 5).
func (m *Model) agentsTitleFacts() string {
	pods := m.agentsPods()
	home := 0
	for _, pod := range pods {
		if pod.homeColumn {
			home = len(pod.nodes())
		}
	}
	facts := fmt.Sprintf("· %s · %d home", plural(projectPods(pods), "project"), home)
	if m.inPodView(pods) {
		facts = "· " + pods[m.podIndex(pods)].group.id + " · one pod"
		if m.graph.podView {
			facts += " · esc pods"
		}
	}
	if hidden := m.hiddenCards(); hidden > 0 {
		facts += fmt.Sprintf(" · h shows %d stopped", hidden)
	}
	return facts
}
