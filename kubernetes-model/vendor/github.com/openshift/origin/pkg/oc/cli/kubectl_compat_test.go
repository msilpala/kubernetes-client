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
package cli

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/spf13/pflag"

	"k8s.io/apimachinery/pkg/util/sets"
	kcmd "k8s.io/kubernetes/pkg/kubectl/cmd"

	"github.com/openshift/origin/pkg/oc/cli/util/clientcmd"
)

// MissingCommands is the list of commands we're already missing.
// NEVER ADD TO THIS LIST
// TODO kill this list
var MissingCommands = sets.NewString(
	"namespace", "rolling-update",
	"cluster-info", "api-versions",
	"stop",

	// are on admin commands
	"cordon",
	"drain",
	"uncordon",
	"taint",
	"top",
	"certificate",

	// TODO commands to assess
	"apiversions",
	"clusterinfo",
	"resize",
	"rollingupdate",
	"run-container",
	"update",
	"alpha",
)

// WhitelistedCommands is the list of commands we're never going to have,
// defend each one with a comment
var WhitelistedCommands = sets.NewString()

func TestKubectlCompatibility(t *testing.T) {
	f := clientcmd.New(pflag.NewFlagSet("name", pflag.ContinueOnError))

	oc := NewCommandCLI("oc", "oc", &bytes.Buffer{}, ioutil.Discard, ioutil.Discard)
	kubectl := kcmd.NewKubectlCommand(f, nil, ioutil.Discard, ioutil.Discard)

kubectlLoop:
	for _, kubecmd := range kubectl.Commands() {
		for _, occmd := range oc.Commands() {
			if kubecmd.Name() == occmd.Name() {
				if MissingCommands.Has(kubecmd.Name()) {
					t.Errorf("%s was supposed to be missing", kubecmd.Name())
					continue
				}
				if WhitelistedCommands.Has(kubecmd.Name()) {
					t.Errorf("%s was supposed to be whitelisted", kubecmd.Name())
					continue
				}
				continue kubectlLoop
			}
		}
		if MissingCommands.Has(kubecmd.Name()) || WhitelistedCommands.Has(kubecmd.Name()) {
			continue
		}

		t.Errorf("missing %q,", kubecmd.Name())
	}
}

// this only checks one level deep for nested commands, but it does ensure that we've gotten several
// --validate flags.  Based on that we can reasonably assume we got them in the kube commands since they
// all share the same registration.
func TestValidateDisabled(t *testing.T) {
	f := clientcmd.New(pflag.NewFlagSet("name", pflag.ContinueOnError))

	oc := NewCommandCLI("oc", "oc", &bytes.Buffer{}, ioutil.Discard, ioutil.Discard)
	kubectl := kcmd.NewKubectlCommand(f, nil, ioutil.Discard, ioutil.Discard)

	for _, kubecmd := range kubectl.Commands() {
		for _, occmd := range oc.Commands() {
			if kubecmd.Name() == occmd.Name() {
				ocValidateFlag := occmd.Flags().Lookup("validate")
				if ocValidateFlag == nil {
					continue
				}

				if ocValidateFlag.Value.String() != "false" {
					t.Errorf("%s --validate is not defaulting to false", occmd.Name())
				}
			}
		}
	}

}
