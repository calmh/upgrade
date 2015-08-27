package upgrade

import (
	"strconv"
	"strings"
)

type relation int

const (
	majorOlder relation = -2 // Older by a major version (x in x.y.z or 0.x.y).
	older               = -1 // Older by a minor version (y or z in x.y.z, or y in 0.x.y)
	equal               = 0  // Versions are semantically equal
	newer               = 1  // Newer by a minor version (y or z in x.y.z, or y in 0.x.y)
	majorNewer          = 2  // Newer by a major version (x in x.y.z or 0.x.y).
)

// CompareVersions returns a relation describing how a compares to b.
func compareVersions(a, b string) relation {
	arel, apre := versionParts(a)
	brel, bpre := versionParts(b)

	minlen := len(arel)
	if l := len(brel); l < minlen {
		minlen = l
	}

	// First compare major-minor-patch versions
	for i := 0; i < minlen; i++ {
		if arel[i] < brel[i] {
			if i == 0 {
				return majorOlder
			}
			if i == 1 && arel[0] == 0 {
				return majorOlder
			}
			return older
		}
		if arel[i] > brel[i] {
			if i == 0 {
				return majorNewer
			}
			if i == 1 && arel[0] == 0 {
				return majorNewer
			}
			return newer
		}
	}

	// Longer version is newer, when the preceding parts are equal
	if len(arel) < len(brel) {
		return older
	}
	if len(arel) > len(brel) {
		return newer
	}

	// Prerelease versions are older, if the versions are the same
	if len(apre) == 0 && len(bpre) > 0 {
		return newer
	}
	if len(apre) > 0 && len(bpre) == 0 {
		return older
	}

	minlen = len(apre)
	if l := len(bpre); l < minlen {
		minlen = l
	}

	// Compare prerelease strings
	for i := 0; i < minlen; i++ {
		switch av := apre[i].(type) {
		case int:
			switch bv := bpre[i].(type) {
			case int:
				if av < bv {
					return older
				}
				if av > bv {
					return newer
				}
			case string:
				return older
			}
		case string:
			switch bv := bpre[i].(type) {
			case int:
				return newer
			case string:
				if av < bv {
					return older
				}
				if av > bv {
					return newer
				}
			}
		}
	}

	// If all else is equal, longer prerelease string is newer
	if len(apre) < len(bpre) {
		return older
	}
	if len(apre) > len(bpre) {
		return newer
	}

	// Looks like they're actually the same
	return equal
}

// Split a version into parts.
// "1.2.3-beta.2" -> []int{1, 2, 3}, []interface{}{"beta", 2}
func versionParts(v string) ([]int, []interface{}) {
	if strings.HasPrefix(v, "v") || strings.HasPrefix(v, "V") {
		// Strip initial 'v' or 'V' prefix if present.
		v = v[1:]
	}
	parts := strings.SplitN(v, "+", 2)
	parts = strings.SplitN(parts[0], "-", 2)
	fields := strings.Split(parts[0], ".")

	release := make([]int, len(fields))
	for i, s := range fields {
		v, _ := strconv.Atoi(s)
		release[i] = v
	}

	var prerelease []interface{}
	if len(parts) > 1 {
		fields = strings.Split(parts[1], ".")
		prerelease = make([]interface{}, len(fields))
		for i, s := range fields {
			v, err := strconv.Atoi(s)
			if err == nil {
				prerelease[i] = v
			} else {
				prerelease[i] = s
			}
		}
	}

	return release, prerelease
}
