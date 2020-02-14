package types

import (
	"github.com/blang/semver"
	"github.com/flavio/stale-container/pkg/stale_container"
)

type CheckResponse struct {
	Image          string `json:"image"`
	Constraint     string `json:"constraint"`
	CurrentVersion string `json:"current_version"`
	NextVersion    string `json:"next_version"`
	Stale          bool   `json:"stale"`
}

func NewCheckResponse(image stale_container.Image, constraint string, nextVer semver.Version) CheckResponse {
	return CheckResponse{
		Image:          image.String(),
		Constraint:     constraint,
		Stale:          nextVer.GT(image.TagVersion),
		CurrentVersion: image.Tag,
		NextVersion:    nextVer.String(),
	}
}
