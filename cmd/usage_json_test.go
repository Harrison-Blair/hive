package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestUsageErrorsUnderJSONPrintOneEnvelope(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"unknown command", []string{"--json", "bogus"}, `unknown command \"bogus\" for \"hive\"`},
		{"missing arguments", []string{"agent", "send", "--json"}, "accepts 2 arg(s), received 0"},
		{"missing argument of register", []string{"project", "register", "--json"}, "accepts 1 arg(s), received 0"},
		{"extra argument", []string{"agent", "stop", "--json=true", "a", "b"}, "accepts 1 arg(s), received 2"},
		{"unknown flag", []string{"agent", "spawn", "--bogus", "--json"}, "unknown flag: --bogus"},
		{"empty register path", []string{"--json", "project", "register", ""}, "PATH requires a value"},
		{"last json flag true", []string{"--json=false", "--json", "bogus"}, `unknown command \"bogus\" for \"hive\"`},
		{"removed init", []string{"--json", "init"}, `unknown command \"init\" for \"hive\"`},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			testHome(t)
			var stdout, stderr bytes.Buffer
			status := executeWithStreams(test.args, strings.NewReader(""), &stdout, &stderr)
			want := `{"outcome":"failed","phase":"usage","effects":[],"error":"` + test.want + `","result":null}` + "\n"
			if status != 2 || stdout.String() != want || stderr.Len() != 0 {
				t.Fatalf("status = %d, stdout = %q, stderr = %q; want 2, %q and no stderr", status, stdout.String(), stderr.String(), want)
			}
		})
	}
}

func TestUsageErrorsWithoutJSONPrintOneLine(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"no json flag", []string{"agent", "send"}, "accepts 2 arg(s), received 0"},
		{"json false", []string{"--json=false", "bogus"}, `unknown command "bogus" for "hive"`},
		{"removed init", []string{"init"}, `unknown command "init" for "hive"`},
		{"json after the terminator", []string{"agent", "stop", "--", "a", "--json"}, "accepts 1 arg(s), received 2"},
		{"empty register path", []string{"project", "register", ""}, "PATH requires a value"},
		{"last json flag false", []string{"--json", "--json=false", "bogus"}, `unknown command "bogus" for "hive"`},
		{"last json flag invalid", []string{"--json", "--json=x", "bogus"}, `invalid argument "x" for "--json" flag: strconv.ParseBool: parsing "x": invalid syntax`},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			testHome(t)
			var stdout, stderr bytes.Buffer
			status := executeWithStreams(test.args, strings.NewReader(""), &stdout, &stderr)
			if status != 2 || stdout.Len() != 0 || stderr.String() != "hive: "+test.want+"\n" {
				t.Fatalf("status = %d, stdout = %q, stderr = %q; want 2 and one line %q", status, stdout.String(), stderr.String(), test.want)
			}
		})
	}
}
