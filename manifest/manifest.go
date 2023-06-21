// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package manifest

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"code.arista.io/eos/tools/eext/util"
)

// Repo spec
// mock cfg dnf.conf is generated from this
type Repo struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	UseBaseArch bool   `yaml:"use-base-arch"`
}

// Build spec
// mock cfg is generated for each target depending on this
type Build struct {
	Include []string `yaml:"include"`
	Repo    []Repo   `yaml:"repo"`
}

// UpstreamSrcSignature specifies detached signature file for tarball
// and specifies the public key to be used to verify the signature.
type UpstreamSrcSignature struct {
	DetachedSig string `yaml:"detached-sig"`
	PubKey      string `yaml:"public-key"`
	SkipCheck   bool   `yaml:"skip-check"`
}

// UpstreamSrc spec
// Lists each source bundle(tarball/srpm) and
// detached signature file for tarball.
type UpstreamSrc struct {
	Source    string               `yaml:"source"`
	Signature UpstreamSrcSignature `yaml:"signature"`
}

// Package spec
// In the general case, there will only be one package/
// But each git repo can have multiple packages if there is
// a dependency order to be maintained.
type Package struct {
	Name            string        `yaml:"name"`
	Subdir          bool          `yaml:"subdir"`
	RpmReleaseMacro string        `yaml:"release"`
	UpstreamSrc     []UpstreamSrc `yaml:"upstream-sources"`
	Type            string        `yaml:"type"`
	Build           Build         `yaml:"build"`
}

// Manifest spec
// This is loaded from eext.yaml
type Manifest struct {
	// In most cases there is only one package.
	Package []Package `yaml:"package"`
}

// LoadManifest loads the manifest file for the repo to memory and
// returns the data structure
func LoadManifest(repo string) (*Manifest, error) {
	repoDir := util.GetRepoDir(repo)

	yamlPath := filepath.Join(repoDir, "eext.yaml")
	yamlContents, readErr := ioutil.ReadFile(yamlPath)
	if readErr != nil {
		return nil, fmt.Errorf("manifest.LoadManifest: ioutil.ReadFile on %s returned %s", yamlPath, readErr)
	}

	var manifest Manifest
	parseErr := yaml.UnmarshalStrict(yamlContents, &manifest)
	if parseErr != nil {
		return nil, fmt.Errorf("manifest.LoadManifest: Error parsing yaml file %s: %s", yamlPath, parseErr)
	}
	return &manifest, nil
}
