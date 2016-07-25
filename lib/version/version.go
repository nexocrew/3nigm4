//
// 3nigm4 crypto package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 21/03/2016
//
package versionmng

// Go standard libraries
import (
	"fmt"
	"os"
	"strconv"
)

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
	ENV_MAJOR_VERSION = "3NIGM4_MAJOR_VERSION"
	ENV_MINOR_VERSION = "3NIGM4_MINOR_VERSION"
	ENV_PATCH_VERSION = "3NIGM4_PATCH_VERSION"
	ENV_RELEASE_TYPE  = "3NIGM4_RELEASE_TYPE"
)

var v *VersionManager // Singleton pattern static variable

// Constant version reference this is
// subordinated to environment variables.
const (
	kMajorVersion = 1
	kMinorVersion = 0
	kPatchVersion = 0
	kReleaseType  = "beta"
)

// Returns a version manager instance using a
// singleton pattern approach. If nil initializes a
// new manager using constant version values.
// Accepts environment variables to build time
// define version components.
func V() *VersionManager {
	if v == nil {
		tv := VersionManager{}
		// setup vars
		env := os.Getenv(ENV_MAJOR_VERSION)
		if env != "" {
			tv.major, _ = strconv.Atoi(env)
		} else {
			tv.major = kMajorVersion
		}
		env = os.Getenv(ENV_MINOR_VERSION)
		if env != "" {
			tv.minor, _ = strconv.Atoi(env)
		} else {
			tv.minor = kMinorVersion
		}
		env = os.Getenv(ENV_PATCH_VERSION)
		if env != "" {
			tv.patch, _ = strconv.Atoi(env)
		} else {
			tv.patch = kPatchVersion
		}
		env = os.Getenv(ENV_RELEASE_TYPE)
		if env != "" {
			tv.releaseType = env
		} else {
			tv.releaseType = kReleaseType
		}
		v = &tv
	}
	return v
}

// Returns a string composing all version components.
func (v *VersionManager) VersionString() string {
	if v.releaseType == "" {
		return fmt.Sprintf("%d.%d.%d", v.major, v.minor, v.patch)
	} else {
		return fmt.Sprintf("%d.%d.%d_%s", v.major, v.minor, v.patch, v.releaseType)
	}
}
