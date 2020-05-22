package integration_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testBuildpackYAML(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		pack   occam.Pack
		docker occam.Docker
	)

	it.Before(func() {
		pack = occam.NewPack().WithVerbose()
		docker = occam.NewDocker()
	})

	context("when the buildpack is run with pack build", func() {
		var (
			image     occam.Image
			container occam.Container
			name      string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
		})

		it("builds with the settings in buildpack.yml", func() {
			var err error
			var logs fmt.Stringer
			image, logs, err = pack.WithNoColor().Build.
				WithNoPull().
				WithBuildpacks(buildpack, buildPlanBuildpack).
				Execute(name, filepath.Join("testdata", "buildpack_yml_app"))
			Expect(err).ToNot(HaveOccurred(), logs.String)

			container, err = docker.Container.Run.WithCommand("python3 server.py").Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())

			Eventually(container).Should(BeAvailable(), ContainerLogs(container.ID))

			response, err := http.Get(fmt.Sprintf("http://localhost:%s", container.HostPort()))
			Expect(err).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusOK))

			content, err := ioutil.ReadAll(response.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("hello world"))

			// buildpackVersion, err := GetGitVersion()
			// Expect(err).ToNot(HaveOccurred())
			//
			// Expect(logs).To(ContainLines(
			// 	fmt.Sprintf("Go Compiler Buildpack %s", buildpackVersion),
			// 	"  Resolving Go version",
			// 	"    Candidate version sources (in priority order):",
			// 	"      buildpack.yml -> \"1.14.*\"",
			// 	"      <unknown>     -> \"\"",
			// 	"",
			// 	MatchRegexp(`    Selected Go version \(using buildpack\.yml\): 1\.14\.\d+`),
			// 	"",
			// 	"  Executing build process",
			// 	MatchRegexp(`    Installing Go 1\.14\.\d+`),
			// 	MatchRegexp(`      Completed in \d+\.\d+`),
			// ))
		})
	})
}
