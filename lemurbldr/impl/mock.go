// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"fmt"
	"path/filepath"
	"strings"

	"lemurbldr/manifest"
	"lemurbldr/util"
)

func fedoraMock(pkg string, arch string, srpmPath string) error {
	errPrefix := fmt.Sprintf("fedoraMock(%s-%s): ", pkg, arch)

	var mockArgs []string

	mockResultsDir := getMockResultsDir(pkg, arch)
	targetArg := "--target=" + arch
	cfgArg := "--root=" + getMockCfgPath(pkg, arch)
	mockArgs = append(mockArgs, fmt.Sprintf("--resultdir=%s", mockResultsDir), targetArg, cfgArg, srpmPath)
	if util.GlobalVar.Quiet {
		mockArgs = append(mockArgs, "--quiet")
	}

	mockErr := util.RunSystemCmd("mock", mockArgs...)
	if mockErr != nil {
		return fmt.Errorf("%smock %s on arch %s errored out with %s",
			errPrefix, pkg, arch, mockErr)
	}

	return nil
}

func filterAndCopy(errPrefix string,
	movePathMap map[string]string, srcPath string) error {
	for destPath, regexStr := range movePathMap {
		filenames, gmfdErr := util.GetMatchingFilenamesFromDir(errPrefix, srcPath, regexStr)
		if gmfdErr != nil {
			return gmfdErr
		}
		if err := util.CopyFilesToDir(errPrefix, filenames, srcPath, destPath, true); err != nil {
			return err
		}
	}
	return nil
}

// Build one SRPM
func mock(repo string, pkgSpec manifest.Package, arch string) error {
	pkg := pkgSpec.Name

	errPrefix := fmt.Sprintf("impl.mock(%s-%s): ", pkg, arch)

	pkgSrpmsDir := getPkgSrpmsDestDir(pkg)
	if err := util.CheckPath(pkgSrpmsDir, true, false); err != nil {
		return fmt.Errorf("%sDirectory %s not found, input .src.rpm is expected here",
			errPrefix, pkgSrpmsDir)
	}

	srpmNames, gmfdErr := util.GetMatchingFilenamesFromDir(errPrefix, pkgSrpmsDir, "")
	if gmfdErr != nil {
		return gmfdErr
	}
	if numMatched := len(srpmNames); numMatched != 1 {
		return fmt.Errorf("%sFound %d files in %s, expected (only) one .src.rpm file",
			errPrefix, numMatched, pkgSrpmsDir)
	}
	srpmName := srpmNames[0]
	srpmPath := filepath.Join(pkgSrpmsDir, srpmName)
	if !strings.HasSuffix(srpmName, ".src.rpm") {
		return fmt.Errorf("%sFile %s found, but expected a .src.rpm file here", errPrefix, srpmPath)
	}

	// These should be created but not cleaned up
	dirsToSetup := []string{getPkgWorkingDir(pkg)}
	if err := util.CreateDirs(errPrefix, dirsToSetup, false); err != nil {
		return err
	}

	// These should be cleaned up and re-created
	mockBaseDir := getMockBaseDir(pkg, arch)
	mockResultsDir := getMockResultsDir(pkg, arch)
	mockCfgDir := getMockCfgDir(pkg, arch)
	dirsToWipeAndRecreate := []string{getPkgRpmsDestDir(pkg), mockBaseDir, mockResultsDir, mockCfgDir}
	if err := util.CreateDirs(errPrefix, dirsToWipeAndRecreate, true); err != nil {
		return err
	}

	if mcgErr := createMockCfgFile(repo, pkgSpec, arch); mcgErr != nil {
		return mcgErr
	}

	if fmErr := fedoraMock(pkgSpec.Name, arch, srpmPath); fmErr != nil {
		return fmErr
	}

	pkgRpmsDestDir := getPkgRpmsDestDir(pkg)
	// move out logs, srpm from resultdir to logs and scratch respectively
	copyPathMap := make(map[string]string)
	copyPathMap[pkgRpmsDestDir] = "(?i).*\\.(noarch|i686|x86_64|aarch64)\\.rpm"
	copyErr := filterAndCopy(errPrefix, copyPathMap, mockResultsDir)
	if copyErr != nil {
		return copyErr
	}

	return nil
}

// Mock calls fedora mock to build the RPMS for the specified target
// from the already built SRPMs and places the results in {WorkingDir}/<pkg>/RPMS
func Mock(repo string, pkg string, arch string) error {
	if err := CheckEnv(); err != nil {
		return err
	}

	// Error out early if source is not available.
	if err := checkRepo(repo); err != nil {
		return err
	}

	repoManifest, loadManifestErr := manifest.LoadManifest(repo)
	if loadManifestErr != nil {
		return loadManifestErr
	}

	// These should be created but not cleaned up
	rpmsDestDir := getAllRpmsDestDir()
	if dirCreateErr := util.MaybeCreateDir("impl.Mock", rpmsDestDir); dirCreateErr != nil {
		return dirCreateErr
	}

	var pkgSpecified bool = (pkg != "")
	found := !pkgSpecified
	for _, pkgSpec := range repoManifest.Package {
		thisPkgName := pkgSpec.Name
		if pkgSpecified && (pkg != thisPkgName) {
			continue
		}
		found = true
		if err := mock(repo, pkgSpec, arch); err != nil {
			return err
		}
	}

	if !found {
		return fmt.Errorf("impl.Mock: Invalid package name %s specified", pkg)
	}

	return nil
}