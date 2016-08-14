//
// 3nigm4 versionmng package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 21/03/2016
//

// Package versionmng manage the software version globally, it's
// used to have a single point of management to assign the
// 3nigm4 framework version to all it's components.
package versionmng

// Go standard libraries
import (
	"fmt"
	"os"
	"strconv"
)

// VersionManager struture holding all required
// version related data.
type VersionManager struct {
	major       int    // major version number;
	minor       int    // minor version number;
	patch       int    // patch version number;
	releaseType string // string component: available alpha, beta e gm.
}

// Environment defined keys: should be used
// in automated environment (CI tools) to
// change compiled version id.
const (
	envMajorVersion = "3NIGM4_MAJOR_VERSION"
	envMinorVersion = "3NIGM4_MINOR_VERSION"
	envPatchVersion = "3NIGM4_PATCH_VERSION"
	envReleaseType  = "3NIGM4_RELEASE_TYPE"
)

var v *VersionManager // Singleton pattern static variable

// Constant version reference this is
// subordinated to environment variables.
const (
	majorVersion = 1
	minorVersion = 0
	patchVersion = 0
	releaseType  = "beta"
)

// V returns a version manager instance using a
// singleton pattern approach. If nil initializes a
// new manager using constant version values.
// Accepts environment variables to build time
// define version components.
func V() *VersionManager {
	if v == nil {
		tv := VersionManager{}
		// setup vars
		env := os.Getenv(envMajorVersion)
		if env != "" {
			tv.major, _ = strconv.Atoi(env)
		} else {
			tv.major = majorVersion
		}
		env = os.Getenv(envMinorVersion)
		if env != "" {
			tv.minor, _ = strconv.Atoi(env)
		} else {
			tv.minor = minorVersion
		}
		env = os.Getenv(envPatchVersion)
		if env != "" {
			tv.patch, _ = strconv.Atoi(env)
		} else {
			tv.patch = patchVersion
		}
		env = os.Getenv(envReleaseType)
		if env != "" {
			tv.releaseType = env
		} else {
			tv.releaseType = releaseType
		}
		v = &tv
	}
	return v
}

// VersionString returns a string composing all version
// components.
func (v *VersionManager) VersionString() string {
	if v.releaseType == "" {
		return fmt.Sprintf("%d.%d.%d", v.major, v.minor, v.patch)
	}
	return fmt.Sprintf("%d.%d.%d_%s", v.major, v.minor, v.patch, v.releaseType)
}
