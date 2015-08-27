// Copyright (C) 2015 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"

	"github.com/calmh/upgrade"
)

func main() {
	allowPre := flag.Bool("pre", false, "Allow prereleases")
	allowMajor := flag.Bool("major", false, "Allow major uprades")
	match := flag.String("match", "", "Match asset name")
	flag.Parse()
	project := flag.Arg(0)
	ver := flag.Arg(1)

	rels, err := upgrade.GithubReleases(project, ver, *allowMajor, *allowPre)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, rel := range rels {
		if *match != "" {
			rel.Assets = upgrade.MatchingAssets(regexp.MustCompile(*match), rel)
		}
		if len(rel.Assets) > 0 {
			fmt.Println("Release", rel.Version)
			for _, asset := range rel.Assets {
				fmt.Println("    Asset", asset.Name, "at", asset.URL)
			}
		}
	}
}
