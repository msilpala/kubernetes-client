/**
 * Copyright (C) 2015 Red Hat, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *         http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package builds

import (
	"time"

	g "github.com/onsi/ginkgo"
	o "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/util/wait"

	exutil "github.com/openshift/origin/test/extended/util"
)

var _ = g.Describe("[Feature:Builds][Conformance] remove all builds when build configuration is removed", func() {
	defer g.GinkgoRecover()
	var (
		buildFixture = exutil.FixturePath("testdata", "builds", "test-build.yaml")
		oc           = exutil.NewCLI("cli-remove-build", exutil.KubeConfigPath())
	)

	g.Context("", func() {
		g.BeforeEach(func() {
			exutil.DumpDockerInfo()
		})

		g.JustBeforeEach(func() {
			g.By("waiting for builder service account")
			err := exutil.WaitForBuilderAccount(oc.KubeClient().Core().ServiceAccounts(oc.Namespace()))
			o.Expect(err).NotTo(o.HaveOccurred())
			oc.Run("create").Args("-f", buildFixture).Execute()
		})

		g.AfterEach(func() {
			if g.CurrentGinkgoTestDescription().Failed {
				exutil.DumpPodStates(oc)
				exutil.DumpPodLogsStartingWith("", oc)
			}
		})

		g.Describe("oc delete buildconfig", func() {
			g.It("should start builds and delete the buildconfig", func() {
				var (
					err    error
					builds [4]string
				)

				g.By("starting multiple builds")
				for i := range builds {
					stdout, _, err := exutil.StartBuild(oc, "sample-build", "-o=name")
					o.Expect(err).NotTo(o.HaveOccurred())
					builds[i] = stdout
				}

				g.By("deleting the buildconfig")
				err = oc.Run("delete").Args("bc/sample-build").Execute()
				o.Expect(err).NotTo(o.HaveOccurred())

				g.By("waiting for builds to clear")
				err = wait.Poll(3*time.Second, 3*time.Minute, func() (bool, error) {
					out, err := oc.Run("get").Args("builds").Output()
					o.Expect(err).NotTo(o.HaveOccurred())
					if out == "No resources found." {
						return true, nil
					}
					return false, nil
				})
				if err == wait.ErrWaitTimeout {
					g.Fail("timed out waiting for builds to clear")
				}
			})

		})
	})
})
