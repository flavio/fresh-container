package stale_container

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/blang/semver"
	"github.com/flavio/stale-container/internal/config"
	"github.com/genuinetools/reg/registry"
	"github.com/genuinetools/reg/repoutils"
)

// Image is an extended struct that represents
// an Image
type Image struct {
	registry.Image
	TagVersion  semver.Version
	TagVersions semver.Versions
}

type ImageUpgradeEvaluationResponse struct {
	Image          string `json:"image"`
	Constraint     string `json:"constraint"`
	CurrentVersion string `json:"current_version"`
	NextVersion    string `json:"next_version"`
	Stale          bool   `json:"stale"`
}

func NewImage(image string) (Image, error) {
	img, err := registry.ParseImage(image)

	if err != nil {
		return Image{}, err
	}

	version, err := semver.Parse(img.Tag)
	if err != nil {
		return Image{}, err
	}

	return Image{
		TagVersion: version,
		Image:      img,
	}, nil
}

// FetchTags queries the registry that holds the image to
// assess the tags it has.
// The tags are automatically converted to semver.Version objects
// and stored into the `TagVersions` field.
// Note well: invalid tags are going to be ignored.
func (image *Image) FetchTags(ctx context.Context, cfg *config.Config) error {
	var err error

	// Create the registry client.
	r, err := createRegistryClient(ctx, image.Domain, cfg)
	if err != nil {
		return err
	}

	tags, err := r.Tags(ctx, image.Path)
	if err != nil {
		return err
	}
	sort.Strings(tags)

	return image.SetTagVersions(tags, true)
}

func (image *Image) SetTagVersions(tags []string, skipInvalid bool) error {
	var err error

	image.TagVersions, err = TagsToVersions(tags, skipInvalid)
	return err
}

func (image *Image) FullNameWithoutTag() string {
	return fmt.Sprintf("%s/%s", image.Domain, image.Path)
}

func (image *Image) EvalUpgrade(constraint string) (ImageUpgradeEvaluationResponse, error) {
	constraintRange, err := semver.ParseRange(constraint)
	if err != nil {
		return ImageUpgradeEvaluationResponse{}, err
	}

	nextVer := NextVersion(
		image.TagVersion,
		constraintRange,
		image.TagVersions,
	)

	return ImageUpgradeEvaluationResponse{
		Image:          image.FullNameWithoutTag(),
		Constraint:     constraint,
		Stale:          nextVer.GT(image.TagVersion),
		CurrentVersion: image.Tag,
		NextVersion:    nextVer.String(),
	}, nil
}

func createRegistryClient(ctx context.Context, domain string, config *config.Config) (*registry.Registry, error) {
	// Use the auth-url domain if provided.
	rc := config.GetRegistryConfig(domain)

	auth, err := repoutils.GetAuthConfig(rc.Username, rc.Password, rc.AuthDomain)
	if err != nil {
		return nil, err
	}

	// Prevent non-ssl unless explicitly forced
	if !rc.NonSSL && strings.HasPrefix(auth.ServerAddress, "http:") {
		return nil, fmt.Errorf("attempted to use insecure protocol! Use force-non-ssl option to force")
	}

	// Create the registry client.
	return registry.New(ctx, auth, registry.Opt{
		Domain:   domain,
		Insecure: rc.Insecure,
		SkipPing: rc.SkipPing,
		NonSSL:   rc.NonSSL,
	})
}
