#!/bin/bash
#
# Copyright (C) 2015 Red Hat, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#         http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#


set -e

if ! git diff-index --quiet HEAD -- || test $(git ls-files --exclude-standard --others | wc -l) != 0; then
  echo "You can't have any staged files in git when updating packages."
  echo "Either commit them or unstage them to continue."
  exit 1
fi

echo "This command will update Origin with the latest stable_proposed branch or tag"
echo "in your OpenShift fork of Kubernetes."
echo
echo "This command is destructive and will alter the contents of your GOPATH projects"
echo
echo "Hit ENTER to continue or CTRL+C to cancel"
read

export GOOS=linux

echo "Restoring Origin dependencies ..."
make clean
godep restore
git fetch --tags

pushd $GOPATH/src/k8s.io/kubernetes > /dev/null
echo "Fetching latest ..."
git fetch
git fetch --tags
popd > /dev/null

pushd $GOPATH/src/k8s.io/kubernetes > /dev/null
git checkout stable_proposed
echo "Restoring any newer Kubernetes dependencies ..."
rm -rf _output
godep restore
popd > /dev/null

echo "Restore complete, update any packages which must diverge from Kubernetes now"
echo
echo "Hit ENTER to continue"
read

echo "Clearing old versions ..."
git rm -r Godeps

echo "Saving dependencies ..."
if ! godep save ./... ; then
  echo "ERROR: Unable to save new dependencies. If packages are listed as not found, try fetching them via 'go get'"
  exit 1
fi
git add .
echo "SUCCESS: Added all new dependencies, review Godeps/Godeps.json"
echo "  To check upstreams, run: git log -E --grep=\"^UPSTREAM:|^bump\" --oneline"
