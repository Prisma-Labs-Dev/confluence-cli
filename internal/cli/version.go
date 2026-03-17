package cli

import "fmt"

type VersionCmd struct{}

func (cmd *VersionCmd) Run(app *App) error {
	if app.IsPlain() {
		discardWrite(fmt.Fprintf(app.Stdout, "confluence %s\n", app.Version))
		return nil
	}
	return renderJSON(app.Stdout, itemEnvelope(VersionInfo{Version: app.Version}, "version", []string{"version"}))
}
