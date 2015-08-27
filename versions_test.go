// Copyright (C) 2014 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at http://mozilla.org/MPL/2.0/.

package upgrade

import "testing"

var versions = []struct {
	a, b string
	r    relation
}{
	{"0.1.2", "0.1.2", equal},
	{"0.1.3", "0.1.2", newer},
	{"0.1.1", "0.1.2", older},
	{"0.3.0", "0.1.2", majorNewer},
	{"0.0.9", "0.1.2", majorOlder},
	{"1.3.0", "1.1.2", newer},
	{"1.0.9", "1.1.2", older},
	{"2.3.0", "1.1.2", majorNewer},
	{"1.0.9", "2.1.2", majorOlder},
	{"1.1.2", "0.1.2", majorNewer},
	{"0.1.2", "1.1.2", majorOlder},
	{"0.1.10", "0.1.9", newer},
	{"0.10.0", "0.2.0", majorNewer},
	{"30.10.0", "4.9.0", majorNewer},
	{"0.9.0-beta7", "0.9.0-beta6", newer},
	{"0.9.0-beta7", "1.0.0-alpha", majorOlder},
	{"1.0.0-alpha", "1.0.0-alpha.1", older},
	{"1.0.0-alpha.1", "1.0.0-alpha.beta", older},
	{"1.0.0-alpha.beta", "1.0.0-beta", older},
	{"1.0.0-beta", "1.0.0-beta.2", older},
	{"1.0.0-beta.2", "1.0.0-beta.11", older},
	{"1.0.0-beta.11", "1.0.0-rc.1", older},
	{"1.0.0-rc.1", "1.0.0", older},
	{"1.0.0+45", "1.0.0+23-dev-foo", equal},
	{"1.0.0-beta.23+45", "1.0.0-beta.23+23-dev-foo", equal},
	{"1.0.0-beta.3+99", "1.0.0-beta.24+0", older},

	{"v1.1.2", "1.1.2", equal},
	{"v1.1.2", "V1.1.2", equal},
	{"1.1.2", "V1.1.2", equal},
}

func TestCompareVersions(t *testing.T) {
	for _, v := range versions {
		if r := compareVersions(v.a, v.b); r != v.r {
			t.Errorf("compareVersions(%q, %q): %d != %d", v.a, v.b, r, v.r)
		}
	}
}
