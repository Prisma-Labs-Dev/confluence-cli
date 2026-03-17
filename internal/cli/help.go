package cli

import (
	"fmt"
	"io"
	"strings"
)

func maybeWriteHelp(args []string, stdout io.Writer) bool {
	if !containsHelpFlag(args) {
		return false
	}

	text, ok := helpText(stripHelpFlags(args))
	if !ok {
		return false
	}

	discardWrite(fmt.Fprint(stdout, text))
	return true
}

func containsHelpFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return true
		}
	}
	return false
}

func stripHelpFlags(args []string) []string {
	filtered := make([]string, 0, len(args))
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			continue
		}
		filtered = append(filtered, arg)
	}
	return filtered
}

func helpHint(command string) string {
	if strings.TrimSpace(command) == "" {
		return "Run `confluence --help` for usage."
	}
	return fmt.Sprintf("Run `confluence %s --help` for usage.", command)
}

func helpText(path []string) (string, bool) {
	switch strings.Join(path, " ") {
	case "":
		return rootHelp(), true
	case "spaces":
		return spacesHelp(), true
	case "spaces list":
		return spacesListHelp(), true
	case "pages":
		return pagesHelp(), true
	case "pages list":
		return pagesListHelp(), true
	case "pages get":
		return pagesGetHelp(), true
	case "pages tree":
		return pagesTreeHelp(), true
	case "pages search":
		return pagesSearchHelp(), true
	case "auth":
		return authHelp(), true
	case "auth login":
		return authLoginHelp(), true
	case "version":
		return versionHelp(), true
	default:
		return "", false
	}
}
