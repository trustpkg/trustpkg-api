package databaseService

import "testing"

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    Command
		wantErr bool
	}{
		{name: "default startup", want: Command{Action: actionStartup}},
		{name: "help", args: []string{"help"}, want: Command{Action: ActionHelp}},
		{name: "first init", args: []string{"first-init"}, want: Command{Action: actionInit}},
		{name: "migrate up", args: []string{"migrate", "up"}, want: Command{Action: actionUp}},
		{name: "migrate down", args: []string{"migrate", "down"}, want: Command{Action: actionDown}},
		{name: "migration version", args: []string{"migrate", "version"}, want: Command{Action: actionVersion}},
		{name: "force version", args: []string{"migrate", "force", "13"}, want: Command{Action: actionForce, Version: 13}},
		{name: "force unset version", args: []string{"migrate", "force", "-1"}, want: Command{Action: actionForce, Version: -1}},
		{name: "unknown action", args: []string{"migrate", "status"}, wantErr: true},
		{name: "unexpected argument", args: []string{"migrate", "up", "extra"}, wantErr: true},
		{name: "missing force version", args: []string{"migrate", "force"}, wantErr: true},
		{name: "invalid force version", args: []string{"migrate", "force", "latest"}, wantErr: true},
		{name: "force version below unset", args: []string{"migrate", "force", "-2"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCommand(tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseCommand() error = %v, wantErr %t", err, tt.wantErr)
			}
			if err == nil && got != tt.want {
				t.Fatalf("ParseCommand() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
