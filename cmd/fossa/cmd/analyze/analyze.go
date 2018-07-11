package analyze

import (
	"fmt"
	"text/template"

	"github.com/urfave/cli"

	"github.com/fossas/fossa-cli/analyzers"
	"github.com/fossas/fossa-cli/api/fossa"
	"github.com/fossas/fossa-cli/cmd/fossa/cmdutil"
	"github.com/fossas/fossa-cli/cmd/fossa/flags"
	"github.com/fossas/fossa-cli/config"
	"github.com/fossas/fossa-cli/log"
	"github.com/fossas/fossa-cli/module"
)

var Cmd = cli.Command{
	Name:      "analyze",
	Usage:     "Analyze built dependencies",
	Action:    Run,
	ArgsUsage: "MODULE",
	Flags:     flags.WithGlobalFlags(flags.WithAPIFlags(flags.WithModulesFlags(flags.WithAnalysisTemplateFlags([]cli.Flag{})))),
}

var _ cli.ActionFunc = Run

func Run(ctx *cli.Context) error {
	err := cmdutil.Init(ctx)
	if err != nil {
		log.Logger.Fatalf("Could not initialize: %s", err.Error())
	}

	modules, err := config.Modules()
	if err != nil {
		log.Logger.Fatalf("Could not parse modules: %s", err.Error())
	}
	if len(modules) == 0 {
		log.Logger.Fatal("No modules specified.")
	}

	analyzed, err := Modules(modules)
	if err != nil {
		log.Logger.Fatalf("Could not analyze modules: %s", err.Error())
		return err
	}

	normalized, err := fossa.Normalize(analyzed)
	if err != nil {
		log.Logger.Fatalf("Could not normalize output: %s", err.Error())
		return err
	}

	if ctx.String(flags.ShowOutput) != "" || ctx.String(flags.Template) != "" {
		var tmpl *template.Template
		if ctx.String(flags.Template) != "" {
			tmpl, err = template.ParseFiles(ctx.String(flags.Template))
			if err != nil {
				log.Logger.Fatalf("Could not parse template data: %s", err.Error())
			}

			if ctx.String(flags.ShowOutput) == "" {
				err = ctx.Set(flags.ShowOutput, "-")
				if err != nil {
					log.Logger.Fatalf("Could not set default output to STDOUT", err.Error())
				}
			}
		}
		return cmdutil.OutputData(ctx.String(flags.ShowOutput), tmpl, normalized)
	}

	return uploadAnalysis(normalized)
}

func Modules(modules []module.Module) (analyzed []module.Module, err error) {
	defer log.StopSpinner()
	for i, m := range modules {
		log.ShowSpinner(fmt.Sprintf("Analyzing module (%d/%d): %s", i+1, len(modules), m.Name))
		analyzer, err := analyzers.New(m.Type, m.Options)
		if err != nil {
			log.Logger.Warningf("Could not load analyzer: %s", err.Error())
			continue
		}
		built, err := analyzer.IsBuilt(m)
		if err != nil {
			log.Logger.Warningf("Could not determine whether module is built: %s", err.Error())
		}
		if !built {
			log.Logger.Warningf("Module does not appear to be built")
		}
		result, err := analyzer.Analyze(m)
		if err != nil {
			log.Logger.Fatalf("Could not analyze: %s", err.Error())
		}
		analyzed = append(analyzed, result)
	}
	log.StopSpinner()

	return analyzed, err
}

func uploadAnalysis(normalized []fossa.SourceUnit) error {
	fossa.MustInit(config.Endpoint(), config.APIKey())
	log.ShowSpinner("Uploading analysis...")
	locator, err := fossa.Upload(config.Fetcher(), config.Project(), config.Revision(), config.Title(), config.Branch(), normalized)
	log.StopSpinner()
	if err != nil {
		log.Logger.Fatalf("Error during upload: %s", err.Error())
		return err
	}
	log.Printf(cmdutil.FmtReportURL(locator))
	return nil
}
