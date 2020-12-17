package integration_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testLayerReuse(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		docker occam.Docker
		pack   occam.Pack

		imageIDs     map[string]struct{}
		containerIDs map[string]struct{}

		name   string
		source string
	)

	it.Before(func() {
		var err error
		name, err = occam.RandomName()
		Expect(err).NotTo(HaveOccurred())

		docker = occam.NewDocker()
		pack = occam.NewPack()
		imageIDs = map[string]struct{}{}
		containerIDs = map[string]struct{}{}
	})

	it.After(func() {
		for id := range containerIDs {
			Expect(docker.Container.Remove.Execute(id)).To(Succeed())
		}

		for id := range imageIDs {
			Expect(docker.Image.Remove.Execute(id)).To(Succeed())
		}

		Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
		Expect(os.RemoveAll(source)).To(Succeed())
	})

	context("an app is rebuilt and aspnet dependency is unchanged", func() {
		it("reuses a layer from a previous build", func() {
			var (
				err         error
				logs        fmt.Stringer
				firstImage  occam.Image
				secondImage occam.Image
			)

			source, err = occam.Source(filepath.Join("testdata", "default_app"))
			Expect(err).NotTo(HaveOccurred())

			build := pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					buildpack,
					buildPlanBuildpack,
				)

			firstImage, logs, err = build.Execute(name, source)
			Expect(err).NotTo(HaveOccurred())

			imageIDs[firstImage.ID] = struct{}{}

			Expect(firstImage.Buildpacks).To(HaveLen(2))

			Expect(firstImage.Buildpacks[0].Key).To(Equal(buildpackInfo.Buildpack.ID))
			Expect(firstImage.Buildpacks[0].Layers).To(HaveKey("dotnet-core-aspnet"))

			Expect(logs.String()).To(ContainSubstring("  Executing build process"))

			// Second pack build
			secondImage, logs, err = build.Execute(name, source)
			Expect(err).NotTo(HaveOccurred())

			imageIDs[secondImage.ID] = struct{}{}

			Expect(secondImage.Buildpacks).To(HaveLen(2))

			Expect(secondImage.Buildpacks[0].Key).To(Equal(buildpackInfo.Buildpack.ID))
			Expect(secondImage.Buildpacks[0].Layers).To(HaveKey("dotnet-core-aspnet"))

			Expect(logs.String()).NotTo(ContainSubstring("  Executing build process"))
			Expect(logs.String()).To(ContainSubstring(fmt.Sprintf("  Reusing cached layer /layers/%s/dotnet-core-aspnet", strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"))))

			Expect(secondImage.Buildpacks[0].Layers["dotnet-core-aspnet"].Metadata["built_at"]).To(Equal(firstImage.Buildpacks[0].Layers["dotnet-core-aspnet"].Metadata["built_at"]))
		})
	})

	context("an app is rebuilt and requirement changes", func() {
		it("does not reuse a layer from the previous build", func() {
			var (
				err         error
				logs        fmt.Stringer
				firstImage  occam.Image
				secondImage occam.Image
			)

			source, err = occam.Source(filepath.Join("testdata", "default_app"))
			Expect(err).NotTo(HaveOccurred())

			build := pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					buildpack,
					buildPlanBuildpack,
				)

			firstImage, logs, err = build.Execute(name, source)
			Expect(err).NotTo(HaveOccurred())

			imageIDs[firstImage.ID] = struct{}{}

			Expect(firstImage.Buildpacks).To(HaveLen(2))

			Expect(firstImage.Buildpacks[0].Key).To(Equal(buildpackInfo.Buildpack.ID))
			Expect(firstImage.Buildpacks[0].Layers).To(HaveKey("dotnet-core-aspnet"))

			Expect(logs.String()).To(ContainSubstring("  Executing build process"))

			// Second pack build
			err = ioutil.WriteFile(filepath.Join(source, "plan.toml"), []byte(`[[requires]]
			name = "dotnet-aspnetcore"

				[requires.metadata]
					launch = true
					version = "2.1.*"
			`), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			secondImage, logs, err = build.Execute(name, source)
			Expect(err).NotTo(HaveOccurred())

			imageIDs[secondImage.ID] = struct{}{}

			Expect(secondImage.Buildpacks).To(HaveLen(2))

			Expect(secondImage.Buildpacks[0].Key).To(Equal(buildpackInfo.Buildpack.ID))
			Expect(secondImage.Buildpacks[0].Layers).To(HaveKey("dotnet-core-aspnet"))

			Expect(logs.String()).To(ContainSubstring("  Executing build process"))
			Expect(logs.String()).NotTo(ContainSubstring("Reusing cached layer"))

			Expect(secondImage.Buildpacks[0].Layers["dotnet-core-aspnet"].Metadata["built_at"]).NotTo(Equal(firstImage.Buildpacks[0].Layers["dotnet-core-aspnet"].Metadata["built_at"]))
		})
	})
}