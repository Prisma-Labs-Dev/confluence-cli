package cli

import "fmt"

type VersionCmd struct{}

func (cmd *VersionCmd) Run(app *App) error {
	if app.Plain {
		fmt.Fprintf(app.Stdout, "confluence %s\n", app.Version)
		return nil
	}
	return renderJSON(app.Stdout, struct {
		Version string `json:"version"`
	}{Version: app.Version})
}
