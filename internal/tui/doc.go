// Package tui is the hive terminal interface that bare `hive` opens.
//
// tui.go defines the root model, its lifecycle, and the program runner.
// shell.go defines the screen interface, the palette, and the list of current screens.
// frame.go defines the screen registry, the frame, the focus, the key routes, and the help overlay.
// chat_screen.go loads and renders chat and coordinator inbox data, and handles Chat input.
// agents.go loads and renders agent topology, and handles pane actions.
// goals_queue.go loads and renders goal and queue records.
// task_screen.go loads and renders task records and evidence.
// friction.go loads dogfood friction entries and draws the Friction list and details.
// health_screen.go loads and renders doctor checks.
// header.go loads and renders project, agent, supervisor, task, and clock data.
// logs_screen.go loads, filters, and renders log events, and handles Logs keys.
// tick.go schedules refreshes and carries background load results.
// demo.go loads a demo fixture, feeds the screens from it, and draws the too-small notice.
package tui
