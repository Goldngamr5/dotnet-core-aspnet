package main

import (
	"os"

	dotnetcoreaspnet "github.com/Goldngamr5/dotnet-core-aspnet"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/draft"
	"github.com/paketo-buildpacks/packit/v2/postal"
)

func main() {
	buildpackYMLParser := dotnetcoreaspnet.NewBuildpackYMLParser()
	logEmitter := dotnetcoreaspnet.NewLogEmitter(os.Stdout)
	entryResolver := draft.NewPlanner()
	dependencyManager := postal.NewService(cargo.NewTransport())
	dotnetRootLinker := dotnetcoreaspnet.NewDotnetRootLinker()

	packit.Run(
		dotnetcoreaspnet.Detect(buildpackYMLParser),
		dotnetcoreaspnet.Build(
			entryResolver,
			dependencyManager,
			dotnetRootLinker,
			logEmitter,
			chronos.DefaultClock,
		),
	)
}
