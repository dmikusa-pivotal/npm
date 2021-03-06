package npm_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/npm/npm"
	"github.com/paketo-buildpacks/npm/npm/fakes"
	"github.com/paketo-buildpacks/packit/scribe"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuildProcessResolver(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		cacheDir   string
		workingDir string

		executable *fakes.Executable
		summer     *fakes.Summer

		resolver npm.BuildProcessResolver

		buffer *bytes.Buffer
	)

	it.Before(func() {
		var err error

		cacheDir, err = ioutil.TempDir("", "cache")
		Expect(err).NotTo(HaveOccurred())

		workingDir, err = ioutil.TempDir("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		executable = &fakes.Executable{}
		summer = &fakes.Summer{}

		buffer = bytes.NewBuffer(nil)
		logger := scribe.NewLogger(buffer)

		resolver = npm.NewBuildProcessResolver(executable, summer, &logger)
	})
	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
		Expect(os.RemoveAll(cacheDir)).To(Succeed())
	})

	context("the build process is install", func() {
		context("there is no node_modules, package-lock.json, or cache", func() {
			it("returns an InstallBuildProcess", func() {
				buildProcess, err := resolver.Resolve(workingDir, cacheDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(buildProcess).To(Equal(npm.NewInstallBuildProcess(executable, scribe.NewLogger(os.Stderr))))

				Expect(buffer.String()).To(ContainSubstring("Selected NPM build process: 'npm install'"))
			})
		})

		context("there is no node_modules, package-lock.json but there is a cache", func() {
			it.Before(func() {
				Expect(os.MkdirAll(filepath.Join(workingDir, "npm-cache"), os.ModePerm)).To(Succeed())

				err := ioutil.WriteFile(filepath.Join(workingDir, "npm-cache", "some-cache-file"), []byte("some-content"), 0644)
				Expect(err).NotTo(HaveOccurred())
			})
			it("returns an InstallBuildProcess", func() {
				buildProcess, err := resolver.Resolve(workingDir, cacheDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(buildProcess).To(Equal(npm.NewInstallBuildProcess(executable, scribe.NewLogger(os.Stderr))))

				contents, err := ioutil.ReadFile(filepath.Join(cacheDir, "npm-cache", "some-cache-file"))
				Expect(err).NotTo(HaveOccurred())
				Expect(string(contents)).To(Equal("some-content"))
			})
		})
	})

	context("the build process is rebuild", func() {
		context("there is no package-lock.json or cache but there is node_modules", func() {
			it.Before(func() {
				Expect(os.MkdirAll(filepath.Join(workingDir, "node_modules"), os.ModePerm)).To(Succeed())
			})

			it("returns a RebuildBuildProcess", func() {
				buildProcess, err := resolver.Resolve(workingDir, cacheDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(buildProcess).To(Equal(npm.NewRebuildBuildProcess(executable, summer, scribe.NewLogger(os.Stderr))))

				Expect(buffer.String()).To(ContainSubstring("Selected NPM build process: 'npm rebuild'"))
			})
		})

		context("there is no package-lock.json but both node_modules and a cache are present", func() {
			it.Before(func() {
				Expect(os.MkdirAll(filepath.Join(workingDir, "node_modules"), os.ModePerm)).To(Succeed())

				Expect(os.MkdirAll(filepath.Join(workingDir, "npm-cache"), os.ModePerm)).To(Succeed())

				err := ioutil.WriteFile(filepath.Join(workingDir, "npm-cache", "some-cache-file"), []byte("some-content"), 0644)
				Expect(err).NotTo(HaveOccurred())
			})

			it("returns a RebuildBuildProcess", func() {
				buildProcess, err := resolver.Resolve(workingDir, cacheDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(buildProcess).To(Equal(npm.NewRebuildBuildProcess(executable, summer, scribe.NewLogger(os.Stderr))))

				contents, err := ioutil.ReadFile(filepath.Join(cacheDir, "npm-cache", "some-cache-file"))
				Expect(err).NotTo(HaveOccurred())
				Expect(string(contents)).To(Equal("some-content"))
			})
		})

		context("there is no cache but node_modules and a package-lock.json are present", func() {
			it.Before(func() {
				Expect(os.MkdirAll(filepath.Join(workingDir, "node_modules"), os.ModePerm)).To(Succeed())

				err := ioutil.WriteFile(filepath.Join(workingDir, "package-lock.json"), []byte("some-content"), 0644)
				Expect(err).NotTo(HaveOccurred())
			})

			it("returns a RebuildBuildProcess", func() {
				buildProcess, err := resolver.Resolve(workingDir, cacheDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(buildProcess).To(Equal(npm.NewRebuildBuildProcess(executable, summer, scribe.NewLogger(os.Stderr))))
			})
		})
	})

	context("the build process is ci", func() {
		context("there is no node_modules or cache but there is a package-lock.json", func() {
			it.Before(func() {
				err := ioutil.WriteFile(filepath.Join(workingDir, "package-lock.json"), []byte("some-content"), 0644)
				Expect(err).NotTo(HaveOccurred())
			})

			it("returns a CIBuildProcess", func() {
				buildProcess, err := resolver.Resolve(workingDir, cacheDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(buildProcess).To(Equal(npm.NewCIBuildProcess(executable, summer, scribe.NewLogger(os.Stderr))))

				Expect(buffer.String()).To(ContainSubstring("Selected NPM build process: 'npm ci'"))
			})
		})

		context("there is no node_modules but the package-lock.json and cache are present", func() {
			it.Before(func() {
				err := ioutil.WriteFile(filepath.Join(workingDir, "package-lock.json"), []byte("some-content"), 0644)
				Expect(err).NotTo(HaveOccurred())

				Expect(os.MkdirAll(filepath.Join(workingDir, "npm-cache"), os.ModePerm)).To(Succeed())

				err = ioutil.WriteFile(filepath.Join(workingDir, "npm-cache", "some-cache-file"), []byte("some-content"), 0644)
				Expect(err).NotTo(HaveOccurred())
			})

			it("returns a CIBuildProcess", func() {
				buildProcess, err := resolver.Resolve(workingDir, cacheDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(buildProcess).To(Equal(npm.NewCIBuildProcess(executable, summer, scribe.NewLogger(os.Stderr))))

				contents, err := ioutil.ReadFile(filepath.Join(cacheDir, "npm-cache", "some-cache-file"))
				Expect(err).NotTo(HaveOccurred())
				Expect(string(contents)).To(Equal("some-content"))
			})
		})

		context("the package-lock.json, node_modules, and cache are present", func() {
			it.Before(func() {
				Expect(os.MkdirAll(filepath.Join(workingDir, "node_modules"), os.ModePerm)).To(Succeed())

				err := ioutil.WriteFile(filepath.Join(workingDir, "package-lock.json"), []byte("some-content"), 0644)
				Expect(err).NotTo(HaveOccurred())

				Expect(os.MkdirAll(filepath.Join(workingDir, "npm-cache"), os.ModePerm)).To(Succeed())

				err = ioutil.WriteFile(filepath.Join(workingDir, "npm-cache", "some-cache-file"), []byte("some-content"), 0644)
				Expect(err).NotTo(HaveOccurred())
			})

			it("returns a CIBuildProcess", func() {
				buildProcess, err := resolver.Resolve(workingDir, cacheDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(buildProcess).To(Equal(npm.NewCIBuildProcess(executable, summer, scribe.NewLogger(os.Stderr))))

				contents, err := ioutil.ReadFile(filepath.Join(cacheDir, "npm-cache", "some-cache-file"))
				Expect(err).NotTo(HaveOccurred())
				Expect(string(contents)).To(Equal("some-content"))
			})
		})
	})

	context("output cases", func() {
		context("when there is a package-lock.json", func() {
			it.Before(func() {
				err := ioutil.WriteFile(filepath.Join(workingDir, "package-lock.json"), []byte("some-content"), 0644)
				Expect(err).NotTo(HaveOccurred())
			})

			it("outputs that package-lock.json was found", func() {
				_, err := resolver.Resolve(workingDir, cacheDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(buffer.String()).To(MatchRegexp(`package-lock.json\s+-> "Found"`))
			})
		})

		context("when there is a node_modules directory", func() {
			it.Before(func() {
				Expect(os.MkdirAll(filepath.Join(workingDir, "node_modules"), os.ModePerm)).To(Succeed())
			})

			it("outputs that package-lock.json was found", func() {
				_, err := resolver.Resolve(workingDir, cacheDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(buffer.String()).To(MatchRegexp(`node_modules\s+-> "Found"`))
			})
		})

		context("when there is a node_modules directory", func() {
			it.Before(func() {
				Expect(os.MkdirAll(filepath.Join(workingDir, "npm-cache"), os.ModePerm)).To(Succeed())
			})

			it("outputs that package-lock.json was found", func() {
				_, err := resolver.Resolve(workingDir, cacheDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(buffer.String()).To(MatchRegexp(`npm-cache\s+-> "Found"`))
			})
		})
	})

	context("failure cases", func() {
		var resolver npm.BuildProcessResolver

		it.Before(func() {
			var err error
			cacheDir, err = ioutil.TempDir("", "layer")
			Expect(err).NotTo(HaveOccurred())

			workingDir, err = ioutil.TempDir("", "working-dir")
			Expect(err).NotTo(HaveOccurred())

			logger := scribe.NewLogger(bytes.NewBuffer(nil))
			resolver = npm.NewBuildProcessResolver(executable, summer, &logger)
		})

		it.After(func() {
			Expect(os.RemoveAll(workingDir)).To(Succeed())
			Expect(os.RemoveAll(cacheDir)).To(Succeed())
		})

		context("Resolve", func() {
			context("when the working directory is unreadable", func() {
				it.Before(func() {
					Expect(os.Chmod(workingDir, 0000)).To(Succeed())
				})

				it.After(func() {
					Expect(os.Chmod(workingDir, os.ModePerm)).To(Succeed())
				})

				it("returns an error", func() {
					_, err := resolver.Resolve(workingDir, cacheDir)
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})

			context("npm-cache exists and is unreadable", func() {
				var npmCacheItemPath string

				it.Before(func() {
					npmCacheItemPath = filepath.Join(workingDir, "npm-cache", "some-cache-dir")
					Expect(os.MkdirAll(npmCacheItemPath, os.ModePerm)).To(Succeed())
					Expect(os.Chmod(npmCacheItemPath, 0000)).To(Succeed())
				})

				it.After(func() {
					Expect(os.Chmod(npmCacheItemPath, os.ModePerm)).To(Succeed())
				})

				it("fails", func() {
					_, err := resolver.Resolve(workingDir, cacheDir)
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})
		})
	})
}
