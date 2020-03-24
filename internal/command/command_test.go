package command

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	backupRunKubectlFunc := runKubectl
	backupRunCommandWithFzf := runCommandWithFzf
	defer func() {
		runKubectl = backupRunKubectlFunc
		runCommandWithFzf = backupRunCommandWithFzf
	}()
	os.Exit(m.Run())
}

func TestGetFzfOption(t *testing.T) {
	testCases := []struct {
		name           string
		previewCommand string
		envVars        map[string]string
		want           string
		wantErr        error
	}{
		{
			name:           "no env vars",
			previewCommand: "kubectl describe pods {{1}}",
			want:           fmt.Sprintf("--inline-info --layout reverse --preview '%s' --preview-window down:70%% --header-lines 1 --bind %s", "kubectl describe pods {{1}}", defaultFzfBindOption),
		},
		{
			name:           "all correct env vars",
			previewCommand: "kubectl describe pods {{1}}",
			envVars: map[string]string{
				envNameFzfOption:     fmt.Sprintf("--preview '$KUBECTL_FZF_FZF_PREVIEW_OPTION' --bind $%s", envNameFzfBindOption),
				envNameFzfBindOption: "ctrl-k:kill-line",
			},
			want: fmt.Sprintf("--preview '%s' --bind %s", "kubectl describe pods {{1}}", "ctrl-k:kill-line"),
		},
		{
			name:           "no env vars",
			previewCommand: "unused preview command",
			envVars: map[string]string{
				envNameFzfOption:     "--inline-info",
				envNameFzfBindOption: "unused",
			},
			want: "--inline-info",
		},
		{
			name:           "invalid env vars in KUBECTL_FZF_FZF_OPTION",
			previewCommand: "unused preview command",
			envVars: map[string]string{
				envNameFzfOption:     "--inline-info $UNKNOWN_ENV_NAME",
				envNameFzfBindOption: "unused",
			},
			want:    "",
			wantErr: fmt.Errorf("%s has invalid environment variables: UNKNOWN_ENV_NAME", envNameFzfOption),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				for k := range tc.envVars {
					require.NoError(t, os.Unsetenv(k))
				}
			}()
			for k, v := range tc.envVars {
				require.NoError(t, os.Setenv(k, v))
			}
			got, gotErr := getFzfOption(tc.previewCommand)
			assert.Equal(t, tc.want, got)
			assert.Equal(t, tc.wantErr, gotErr)
		})
	}
}

func TestCommandFromTemplate(t *testing.T) {
	testCases := []struct {
		name         string
		templateName string
		command      string
		data         map[string]interface{}
		want         string
		wantIsErr    bool
	}{
		{
			name:         "template",
			templateName: "template",
			command:      "kubectl {{ .command }} {{ .resource }}",
			data: map[string]interface{}{
				"command":  "get",
				"resource": "pods",
			},
			want:      "kubectl get pods",
			wantIsErr: false,
		},
		{
			name:         "no template",
			templateName: "",
			command:      "{{ .name }}",
			data: map[string]interface{}{
				"name": "fzf",
			},
			want:      "fzf",
			wantIsErr: false,
		},
		{
			name:         "invalid command",
			templateName: "template",
			command:      "{{ .name }",
			data: map[string]interface{}{
				"name": "name",
			},
			want:      "",
			wantIsErr: true,
		},
		{
			name:         "wrong parameter",
			templateName: "template",
			command:      "wrong {{ .name }}",
			data: map[string]interface{}{
				"unknown": "unknown",
			},
			want:      "wrong ",
			wantIsErr: false,
		},
		{
			name:         "no parameter",
			templateName: "template",
			command:      "no {{ .name }}",
			data:         nil,
			want:         "no ",
			wantIsErr:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, gotErr := commandFromTemplate(tc.templateName, tc.command, tc.data)
			assert.Equal(t, tc.want, got)
			assert.Equal(t, tc.wantIsErr, gotErr != nil)
		})
	}
}
