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

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	exutil "github.com/openshift/origin/test/extended/util"
)

var _ = g.Describe("[Feature:Builds][Conformance] imagechangetriggers", func() {
	defer g.GinkgoRecover()

	var (
		buildFixture = exutil.FixturePath("testdata", "builds", "test-imagechangetriggers.yaml")
		oc           = exutil.NewCLI("imagechangetriggers", exutil.KubeConfigPath())
	)

	g.Context("", func() {
		g.BeforeEach(func() {
			exutil.DumpDockerInfo()
		})

		g.JustBeforeEach(func() {
			g.By("waiting for builder service account")
			err := exutil.WaitForBuilderAccount(oc.AdminKubeClient().Core().ServiceAccounts(oc.Namespace()))
			o.Expect(err).NotTo(o.HaveOccurred())
		})

		g.AfterEach(func() {
			if g.CurrentGinkgoTestDescription().Failed {
				exutil.DumpPodStates(oc)
				exutil.DumpPodLogsStartingWith("", oc)
			}
		})

		g.It("imagechangetriggers should trigger builds of all types", func() {
			err := oc.AsAdmin().Run("create").Args("-f", buildFixture).Execute()
			o.Expect(err).NotTo(o.HaveOccurred())

			err = wait.Poll(time.Second, 30*time.Second, func() (done bool, err error) {
				for _, build := range []string{"bc-docker-1", "bc-jenkins-1", "bc-source-1", "bc-custom-1"} {
					_, err := oc.BuildClient().Build().Builds(oc.Namespace()).Get(build, metav1.GetOptions{})
					if err == nil {
						continue
					}
					if kerrors.IsNotFound(err) {
						return false, nil
					}
					return false, err
				}
				return true, nil
			})
			o.Expect(err).NotTo(o.HaveOccurred())
		})
	})
})
