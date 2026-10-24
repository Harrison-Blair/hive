// Package tui is the hive terminal interface that bare `hive` opens.
//
// tui.go defines the root model, its lifecycle, and the program runner.
// shell.go defines the screen interface, the palette, and the list of current screens.
// frame.go defines the screen registry, the frame, the focus, and the key routes.
// help.go draws the Help overlay and handles its keys.
// chat.go loads and renders chat and coordinator inbox data, and handles Chat input.
// mouse.go handles mouse clicks and wheel steps on each tab.
// agents.go loads and renders agent topology, and handles pane actions.
// goals_queue.go loads and renders goal and queue records.
// work_tasks.go loads task records and files, and draws the Work Tasks view.
// friction.go loads dogfood friction entries and handles the keys of the Friction screen.
// friction_view.go draws the Friction list, the filter line, and the details box.
// health_screen.go loads and renders doctor checks.
// header.go loads and renders project, agent, supervisor, task, and clock data.
// logs_screen.go loads, filters, and renders log events, and handles Logs keys.
// logs_strip.go buckets the kept log events into the Logs density strip.
// logs_fields.go gives the label and value rows of the Logs field panel.
// tick.go schedules refreshes and carries background load results.
// demo.go loads a demo fixture, feeds the screens from it, and draws the too-small notice.
// gframe.go draws the golden-frames status rows and the notice line, and handles the chords.
// status_command.go runs the status command and keeps the first line of its output.
package tui
