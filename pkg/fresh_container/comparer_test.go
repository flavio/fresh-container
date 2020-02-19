package fresh_container

import (
	"testing"
)

func TestNextReleaseInvalidConstraint(t *testing.T) {
	_, err := NextTag("1.1", "> 1.0", []string{})
	if err == nil {
		t.Error("Expected failure parsing invalid constraint")
	}
}

func TestNextReleaseInvalidVersion(t *testing.T) {
	_, err := NextTag("1.1", "> 1.1.0", []string{})
	if err == nil {
		t.Error("Expected failure parsing invalid version")
	}
}

func TestNextReleaseInvalidVersions(t *testing.T) {
	_, err := NextTag("1.1.0", "> 1.1.0", []string{"1.1"})
	if err == nil {
		t.Error("Expected failure parsing invalid versions")
	}
}

type NextTagTestCase struct {
	CurTag      string
	ExpectedTag string
	Constraint  string
	Tags        []string
}

func TestNextTag(t *testing.T) {
	testCases := []NextTagTestCase{
		NextTagTestCase{
			CurTag:     "1.5.0",
			Constraint: ">= 1.5.0 < 1.6.0",
			Tags: []string{
				"1.4.0",
				"1.5.0",
				"1.5.1",
				"1.5.2",
				"1.5.6-alpine",
				"1.6.3",
			},
			ExpectedTag: "1.5.2",
		},
		NextTagTestCase{
			CurTag:     "1.5.0-alpine",
			Constraint: ">= 1.5.0-alpine < 1.6.0-alpine",
			Tags: []string{
				"1.4.0",
				"1.5.0",
				"1.5.0-alpine",
				"1.5.1",
				"1.5.2",
				"1.5.6-alpine",
				"1.6.3",
				"1.6.3-alpine",
			},
			ExpectedTag: "1.5.6-alpine",
		},
		NextTagTestCase{
			CurTag:     "1.5.0-alpine",
			Constraint: ">= 1.5.0-alpine < 1.6.0-alpine",
			Tags: []string{
				"1.4.0",
				"1.5.0",
				"1.5.0-alpine",
				"1.5.1",
				"1.5.2",
				"1.5.6-alpine",
				"1.5.8-data-alpine",
				"1.6.3",
				"1.6.3-alpine",
			},
			ExpectedTag: "1.5.6-alpine",
		},
		NextTagTestCase{
			CurTag:     "1.5.0",
			Constraint: ">= 1.5.0 < 1.6.0",
			Tags: []string{
				"1.4.0",
				"1.5.0",
				"1.5.6-alpine",
				"1.5.8-data-alpine",
				"1.6.3",
				"1.6.3-alpine",
			},
			ExpectedTag: "1.5.0",
		},
		NextTagTestCase{
			CurTag:     "1.5.0",
			Constraint: ">= 1.5.0 < 2.0.0",
			Tags: []string{
				"1.4.0",
				"1.5.0",
				"1.5.6-alpine",
				"1.5.8-data-alpine",
				"1.6.3",
				"1.6.3-alpine",
				"2.0.3",
			},
			ExpectedTag: "1.6.3",
		},
	}

	for _, tc := range testCases {
		nextTag, err := NextTag(tc.CurTag, tc.Constraint, tc.Tags)
		if err != nil {
			t.Errorf("Unexpected error when handling test case %+v: %+v", tc, err)
		}

		if nextTag != tc.ExpectedTag {
			t.Errorf("Unexpected next tag for test case %+v, got %s instead of %s",
				tc,
				nextTag,
				tc.ExpectedTag)
		}
	}
}
