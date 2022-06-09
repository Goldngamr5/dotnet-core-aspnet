package dotnetcoreaspnet

import (
	"os"
	"path/filepath"
	"fmt"

	"github.com/paketo-buildpacks/packit/v2"
)

//go:generate faux --interface VersionParser --output fakes/version_parser.go
type VersionParser interface {
	ParseVersion(path string) (version string, err error)
}

func Detect(buildpackYMLParser VersionParser) packit.DetectFunc {
	fmt.Println("Detect called")
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		fmt.Println("getting req")
		var requirements = []packit.BuildPlanRequirement{
			{
				Name: "dotnet-runtime",
				Metadata: map[string]interface{}{
					"build": true,
				},
			},
		}
		fmt.Println("checking for BP_DOTNET_FRAMEWORK_VERSION")
		// check if BP_DOTNET_FRAMEWORK_VERSION is set
		if version, ok := os.LookupEnv("BP_DOTNET_FRAMEWORK_VERSION"); ok {
			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: "dotnet-aspnetcore",
				Metadata: map[string]interface{}{
					"version-source": "BP_DOTNET_FRAMEWORK_VERSION",
					"version":        version,
				},
			})
		}
		fmt.Println("checking for buildpack.yml")
		// check if the version is set in the buildpack.yml
		version, err := buildpackYMLParser.ParseVersion(filepath.Join(context.WorkingDir, "buildpack.yml"))
		if err != nil {
			fmt.Println("returning error")
			return packit.DetectResult{}, err
		}
		fmt.Println("checking what ver")
		if version != "" {
			fmt.Println("requirements")
			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: "dotnet-aspnetcore",
				Metadata: map[string]interface{}{
					"version-source": "buildpack.yml",
					"version":        version,
				},
			})
		}
		fmt.Println("returning detect result")
		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: "dotnet-aspnetcore"},
				},
				Requires: requirements,
			},
		}, nil
	}
}
