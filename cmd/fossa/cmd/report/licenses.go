package report

import (
	"fmt"
	"text/template"

	"github.com/fossas/fossa-cli/api/fossa"
	"github.com/fossas/fossa-cli/cmd/fossa/flags"
	"github.com/fossas/fossa-cli/log"
	"github.com/urfave/cli"
)

const defaultLicenceReportTemplate = `# 3rd-Party Software License Notice
Generated by fossa-cli (https://github.com/fossas/fossa-cli).
This software includes the following software and licenses:
{{range $license, $deps := .}}
========================================================================
{{$license}}
========================================================================
The following software have components provided under the terms of this license:
{{range $i, $dep := $deps}}
- {{$dep.Project.Title}} (from {{$dep.Project.URL}})
{{- end}}
{{end}}
`

/*
BAD:  https://app.fossa.io/api/revisions?locator=git%2Bgithub.com%2Fprometheus%2Fclient_model%2Fgo%2499fa1f4be8e564e8a6b613da7fa6f46c9edafc6c
GOOD: https://app.fossa.io/api/revisions/git%2Bgithub.com%2Fprometheus%2Fclient_model%2Fgo%2499fa1f4be8e564e8a6b613da7fa6f46c9edafc6c
      -301-> https://app.fossa.io/api/revisions/git%2Bgithub.com%2Fprometheus%2Fclient_model%2499fa1f4be8e564e8a6b613da7fa6f46c9edafc6c/
*/
var licensesCmd = cli.Command{
	Name:      "licenses",
	Usage:     "Generate licenses report",
	Flags: flags.WithGlobalFlags(flags.WithAPIFlags(flags.WithModulesFlags([]cli.Flag{
		cli.StringFlag{Name: flags.Short(Output), Destination: &outputFlag, Value: "-", Usage: "Output file for report"},
		cli.StringFlag{Name: flags.Short(Template), Destination: &templateFlag, Usage: "process report via template prior to sending it to output"},
		cli.BoolFlag{Name: flags.Short(Unknown), Usage: "license report including unkown (warning this is SLOW)"},
	}))),
	Before: prepareReportCtx,
	Action: generateLicenses,
}

func generateLicenses(ctx *cli.Context) (err error) {
	defer log.StopSpinner()
	revs := make(fossa.Revisions, 0)
	for _, module := range analyzed {
		if ctx.Bool(Unknown) {
			totalDeps := len(module.Deps)
			i := 0
			for _, dep := range module.Deps {
				i++
				log.ShowSpinner(fmt.Sprintf("Fetching Licence Info (%d/%d): %s", i+1, totalDeps, dep.ID.Name))
				rev, err := fossa.FetchRevisionForPackage(dep)
				if err != nil {
					println(err.Error())
					continue
				}
				revs = append(revs, rev)
			}
		} else {
			log.ShowSpinner("Fetching Licence Info")
			revs, err = fossa.FetchRevisionForDeps(module.Deps)
			if err != nil {
				log.Logger.Fatalf("Could not fetch revisions: %s", err.Error())
			}
		}
	}

	// TODO: Make sure Revisions is a set
	depsByLicence := make(map[string]fossa.Revisions, 0)
	for _, rev := range revs {
		for _, licence := range rev.Licenses {
			if _, ok := depsByLicence[licence.LicenseId]; !ok {
				depsByLicence[licence.LicenseId] = make(fossa.Revisions, 0)
			}
			depsByLicence[licence.LicenseId] = append(depsByLicence[licence.LicenseId], rev)
		}
	}

	tmpl, err := template.New("base").Parse(defaultLicenceReportTemplate)
	if err != nil {
		log.Logger.Fatalf("Could not parse template data: %s", err.Error())
	}

	if templateFlag != "" {
		tmpl, err = template.ParseFiles(templateFlag)
		if err != nil {
			log.Logger.Fatalf("Could not parse template data: %s", err.Error())
		}
	}
	log.StopSpinner()

	return outputReport(outputFlag, tmpl, depsByLicence)
}