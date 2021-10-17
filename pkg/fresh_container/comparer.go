package fresh_container

import (
	"github.com/blang/semver"
	"strings"
	"fmt"
)

func NextTag(curTag, constraint string, tagPrefix string, tags []string) (string, error) {
	trimmedTag:=strings.TrimPrefix(curTag, tagPrefix)
	if tagPrefix!="" && trimmedTag==curTag {
		err:=fmt.Errorf(
				"The current tag '%s' didn't start with the tag prefix '%s'.",
				curTag,
				tagPrefix)
		return "", err
	}

	curVer, err := semver.Parse(trimmedTag)
	if err != nil {
		return "", err
	}

	constraintRange, err := semver.ParseRange(constraint)
	if err != nil {
		return "", err
	}

	versions, err := TagsToVersions(tags, tagPrefix, false)
	if err != nil {
		return "", err
	}

	nextVer := NextVersion(curVer, constraintRange, tagPrefix, versions)

	return nextVer.String(), nil
}

func NextVersion(curVer semver.Version, constraintRange semver.Range, tagPrefix string, versions semver.Versions) semver.Version {
	nextVer := curVer
	for _, v := range versions {
		if constraintRange(v) {
			if v.GTE(nextVer) && samePre(v, nextVer) {
				nextVer = v
			}
		}
	}

	return nextVer
}

func samePre(v1, v2 semver.Version) bool {
	if len(v1.Pre) != len(v2.Pre) {
		return false
	}

	for i := 0; i < len(v1.Pre); i++ {
		if v1.Pre[i].Compare(v2.Pre[i]) != 0 {
			return false
		}
	}

	return true
}
