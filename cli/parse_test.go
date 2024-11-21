package cli_test

import (
	"strings"
	"testing"

	"github.com/kb-sp/goverter"
	"github.com/kb-sp/goverter/cli"
	"github.com/kb-sp/goverter/config"
	"github.com/kb-sp/goverter/enum"
	"github.com/stretchr/testify/require"
)

func TestError(t *testing.T) {
	tests := []struct {
		args     []string
		contains string
	}{
		{[]string{}, "Error: invalid args"},
		{[]string{"goverter"}, "goverter gen [OPTIONS]"},
		{[]string{"goverter"}, "Error: missing command"},
		{[]string{"goverter", "-u"}, "Error: flag provided but not defined: -u"},
		{[]string{"goverter", "test"}, "Error: unknown command test"},
		{[]string{"goverter", "gen"}, "Error: missing PATTERN"},
		{[]string{"goverter", "gen", "-u"}, "Error: flag provided but not defined: -u"},
		{[]string{"goverter", "gen", "-g"}, "Error: flag needs an argument: -g"},
	}

	for _, test := range tests {
		test := test
		t.Run(strings.Join(test.args, " "), func(t *testing.T) {
			_, err := cli.Parse(test.args)
			require.ErrorContains(t, err, test.contains)
		})
	}
}

func TestHelp(t *testing.T) {
	tests := [][]string{
		{"goverter", "help"},
		{"goverter", "-h"},
		{"goverter", "--help"},
		{"goverter", "gen", "-h"},
		{"goverter", "gen", "--help"},
	}

	for _, test := range tests {
		test := test
		t.Run(strings.Join(test, " "), func(t *testing.T) {
			cmd, err := cli.Parse(test)
			require.NoError(t, err)
			require.IsType(t, &cli.Help{}, cmd)
		})
	}
}

func TestVersion(t *testing.T) {
	cmd, err := cli.Parse([]string{"goverter", "version"})
	require.NoError(t, err)
	require.IsType(t, &cli.Version{}, cmd)
}

func TestSuccess(t *testing.T) {
	actual, err := cli.Parse([]string{
		"goverter",
		"gen",
		"-cwd", "file/path",
		"-build-tags", "",
		"-output-constraint", "",
		"-g", "g1",
		"-global", "g2",
		"-g", "g3 oops",
		"pattern1", "pattern2",
	})
	require.NoError(t, err)

	expected := &cli.Generate{&goverter.GenerateConfig{
		PackagePatterns:       []string{"pattern1", "pattern2"},
		WorkingDir:            "file/path",
		OutputBuildConstraint: "",
		BuildTags:             "",
		EnumTransformers:      map[string]enum.Transformer{},
		Global: config.RawLines{
			Location: "command line (-g, -global)",
			Lines:    []string{"g1", "g2", "g3 oops"},
		},
	}}
	require.Equal(t, expected, actual)
}

func TestDefault(t *testing.T) {
	actual, err := cli.Parse([]string{"goverter", "gen", "pattern"})
	require.NoError(t, err)

	expected := &cli.Generate{&goverter.GenerateConfig{
		PackagePatterns:       []string{"pattern"},
		WorkingDir:            "",
		OutputBuildConstraint: "!goverter",
		BuildTags:             "goverter",
		EnumTransformers:      map[string]enum.Transformer{},
		Global: config.RawLines{
			Location: "command line (-g, -global)",
			Lines:    nil,
		},
	}}
	require.Equal(t, expected, actual)
}
