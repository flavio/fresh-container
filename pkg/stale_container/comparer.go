package stale_container

import (
	"github.com/blang/semver"
)

func NextTag(curTag, constraint string, tags []string) (string, error) {
	curVer, err := semver.Parse(curTag)
	if err != nil {
		return "", err
	}

	constraintRange, err := semver.ParseRange(constraint)
	if err != nil {
		return "", err
	}

	versions, err := TagsToVersions(tags, false)
	if err != nil {
		return "", err
	}

	nextVer := NextVersion(curVer, constraintRange, versions)

	return nextVer.String(), nil
}

func NextVersion(curVer semver.Version, constraintRange semver.Range, versions semver.Versions) semver.Version {
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
