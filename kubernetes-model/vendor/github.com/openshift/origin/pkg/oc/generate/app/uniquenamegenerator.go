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
package app

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/golang/glog"
	"github.com/openshift/origin/pkg/api/apihelpers"
	kvalidation "k8s.io/apimachinery/pkg/util/validation"
)

// the opposite of kvalidation.DNS1123LabelFmt
var invalidNameCharactersRegexp = regexp.MustCompile("[^-a-z0-9]")

// A UniqueNameGenerator is able to generate unique names from a given original
// name.
type UniqueNameGenerator interface {
	Generate(NameSuggester) (string, error)
}

// NewUniqueNameGenerator creates a new UniqueNameGenerator with the given
// original name.
func NewUniqueNameGenerator(name string) UniqueNameGenerator {
	return &uniqueNameGenerator{name, map[string]int{}}
}

type uniqueNameGenerator struct {
	originalName string
	names        map[string]int
}

// Generate returns a name that is unique within the set of names of this unique
// name generator. If the generator's original name is empty, a new name will be
// suggested.
func (ung *uniqueNameGenerator) Generate(suggester NameSuggester) (string, error) {
	name := ung.originalName
	if len(name) == 0 {
		var ok bool
		name, ok = suggester.SuggestName()
		if !ok {
			return "", ErrNameRequired
		}
	}
	return ung.ensureValidName(name)
}

// ensureValidName returns a new name based on the name given that is unique in
// the set of names of this unique name generator.
func (ung *uniqueNameGenerator) ensureValidName(name string) (string, error) {
	names := ung.names

	// Ensure that name meets length requirements
	if len(name) < 2 {
		return "", fmt.Errorf("invalid name: %s", name)
	}

	if !IsParameterizableValue(name) {

		// Make all names lowercase
		name = strings.ToLower(name)

		// Remove everything except [-0-9a-z]
		name = invalidNameCharactersRegexp.ReplaceAllString(name, "")

		// Remove leading hyphen(s) that may be introduced by the previous step
		name = strings.TrimLeft(name, "-")

		if len(name) > kvalidation.DNS1123SubdomainMaxLength {
			glog.V(4).Infof("Trimming %s to maximum allowable length (%d)\n", name, kvalidation.DNS1123SubdomainMaxLength)
			name = name[:kvalidation.DNS1123SubdomainMaxLength]
		}
	}

	count, existing := names[name]
	if !existing {
		names[name] = 0
		return name, nil
	}
	count++
	names[name] = count
	newName := apihelpers.GetName(name, strconv.Itoa(count), kvalidation.DNS1123SubdomainMaxLength)
	return newName, nil
}
