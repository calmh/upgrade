// Copyright (C) 2015 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at http://mozilla.org/MPL/2.0/.

package upgrade

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
)

// The maximum number of matching releases returned by a release lister.
const maxReleases = 5

// A Release represents a given software release.
type Release struct {
	// The Version should be in semver format, i.e. "0.1.2" or "v2.3.4".
	Version string

	// Assets are the actual files that make up the release, being one archive
	// file per OS and architecture.
	Assets []Asset
}

// An Asset is an archive file for a given OS and architecture.
type Asset struct {
	// The asset name or description.
	Name string

	// The download URL.
	URL string
}

// This is an HTTP/HTTPS client that does *not* perform certificate
// validation. We do this because some systems where Syncthing runs have
// issues with old or missing CA roots. It doesn't actually matter that we
// load the upgrade insecurely as we verify an ECDSA signature of the actual
// binary contents before accepting the upgrade.
var insecureHTTP = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

// GithubRelease returns a list of releases for the project that are newer
// than currentVersion. If allowMajorUpgrade is false, releases that are
// majorly newer than current are filtered out. If allowPrerelease is false,
// prerelease releases are filtered out. If an error occurs, the list of
// releases is empty and an error is set. An empty list of releases and a nil
// error means that no error ocurred, but there is no newer release available
// that meets the criteria.
//
//     rels, err := GithubReleases("calmh/someproject", "v1.2.3", true, false)
func GithubReleases(project string, currentVersion string, allowMajorUpgrade bool, allowPrerelease bool) ([]Release, error) {
	resp, err := insecureHTTP.Get("https://api.github.com/repos/" + project + "/releases?per_page=30")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("API call returned HTTP error: %s", resp.Status)
	}

	// We use a temporary type because field names differ from what we want to
	// present, and we don't want to encode Github-specific stuff in the
	// exported Release type.
	type ghRel struct {
		Tag        string `json:"tag_name"`
		Prerelease bool
		Assets     []Asset
	}

	var ghRels []ghRel
	if err := json.NewDecoder(resp.Body).Decode(&ghRels); err != nil {
		return nil, err
	}

	var rels []Release
	for _, r := range ghRels {
		if r.Prerelease && !allowPrerelease {
			continue
		}
		comp := compareVersions(r.Tag, currentVersion)
		if comp <= 0 {
			continue
		}
		if comp == majorNewer && !allowMajorUpgrade {
			continue
		}
		rels = append(rels, Release{
			Version: r.Tag,
			Assets:  r.Assets,
		})
	}

	sort.Sort(sortByRelease(rels))

	if len(rels) > maxReleases {
		rels = rels[:maxReleases]
	}
	return rels, nil
}

// MatchingAssets returns the list of assets that have names matching the
// given expression. I.e., to get a list of assets suitable for Darwin/AMD64
// where the convention is for assets to contain
// {{runtime.GOOS}}-{{runtime.GOARCH}} in the name:
//
//     assets := MatchingAssets(regex.MustCompile(`darwin-amd64`), myRelease)
func MatchingAssets(exp *regexp.Regexp, rel Release) []Asset {
	var matches []Asset
	for _, asset := range rel.Assets {
		if exp.MatchString(asset.Name) {
			matches = append(matches, asset)
		}
	}
	return matches
}

// Release sorting

type sortByRelease []Release

func (s sortByRelease) Len() int {
	return len(s)
}
func (s sortByRelease) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s sortByRelease) Less(i, j int) bool {
	return compareVersions(s[i].Version, s[j].Version) > 0
}
