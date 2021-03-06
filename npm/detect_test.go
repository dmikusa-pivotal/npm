package npm_test

import (
	"errors"
	"os"
	"testing"

	"github.com/paketo-buildpacks/npm/npm"
	"github.com/paketo-buildpacks/npm/npm/fakes"
	"github.com/paketo-buildpacks/packit"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		packageJSONParser *fakes.VersionParser
		detect            packit.DetectFunc
	)

	it.Before(func() {
		packageJSONParser = &fakes.VersionParser{}
		packageJSONParser.ParseVersionCall.Returns.Version = "1.2.3"

		detect = npm.Detect(packageJSONParser)
	})

	it("returns a plan that provides node_modules", func() {
		result, err := detect(packit.DetectContext{
			WorkingDir: "/working-dir",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Plan).To(Equal(packit.BuildPlan{
			Provides: []packit.BuildPlanProvision{
				{Name: npm.PlanDependencyNodeModules},
			},
			Requires: []packit.BuildPlanRequirement{
				{Name: npm.PlanDependencyNodeModules},
				{
					Name:    npm.PlanDependencyNode,
					Version: "1.2.3",
					Metadata: npm.BuildPlanMetadata{
						VersionSource: "package.json",
						Build:         true,
						Launch:        true,
					},
				},
			},
		}))

		Expect(packageJSONParser.ParseVersionCall.Receives.Path).To(Equal("/working-dir/package.json"))
	})

	context("when the package.json does not declare a node engine version", func() {
		it.Before(func() {
			packageJSONParser.ParseVersionCall.Returns.Version = ""
		})

		it("returns a plan that does not declare a node version", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: "/working-dir",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: npm.PlanDependencyNodeModules},
				},
				Requires: []packit.BuildPlanRequirement{
					{Name: npm.PlanDependencyNodeModules},
					{
						Name: npm.PlanDependencyNode,
						Metadata: npm.BuildPlanMetadata{
							Build:  true,
							Launch: true,
						},
					},
				},
			}))

			Expect(packageJSONParser.ParseVersionCall.Receives.Path).To(Equal("/working-dir/package.json"))
		})
	})

	context("when the package.json file does not exist", func() {
		it.Before(func() {
			_, err := os.Stat("no such file")
			packageJSONParser.ParseVersionCall.Returns.Err = err
		})

		it("fails detection", func() {
			_, err := detect(packit.DetectContext{
				WorkingDir: "/working-dir",
			})
			Expect(err).To(MatchError(packit.Fail))
		})
	})

	context("failure cases", func() {
		context("when the package.json parser fails", func() {
			it.Before(func() {
				packageJSONParser.ParseVersionCall.Returns.Err = errors.New("failed to parse package.json")
			})

			it("returns an error", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: "/working-dir",
				})
				Expect(err).To(MatchError("failed to parse package.json"))
			})
		})
	})
}
