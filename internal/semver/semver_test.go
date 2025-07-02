package semver

import "testing"

var parseVersionNoteTests = []struct {
	name     string
	body     string
	expected VersionType
}{
	{
		name: "Patch checked",
		body: `[SYAC] Please check **one** to indicate the semantic versioning level of this release
- [x] **Patch** (*.*.x) - Bug fixes only, no breaking changes
- [ ] **Minor** (*.x.*) - New functionality, backwards compatible
- [ ] **Major** (x.*.*) - Breaking changes or incompatible API updates
`,
		expected: Patch,
	},
	{
		name: "Minor checked",
		body: `[SYAC] Please check **one** to indicate the semantic versioning level of this release
- [ ] **Patch** (*.*.x) - Bug fixes only, no breaking changes
- [x] **Minor** (*.x.*) - New functionality, backwards compatible
- [ ] **Major** (x.*.*) - Breaking changes or incompatible API updates
`,
		expected: Minor,
	},
	{
		name: "Major checked",
		body: `[SYAC] Please check **one** to indicate the semantic versioning level of this release
- [ ] **Patch** (*.*.x) - Bug fixes only, no breaking changes
- [ ] **Minor** (*.x.*) - New functionality, backwards compatible
- [x] **Major** (x.*.*) - Breaking changes or incompatible API updates
`,
		expected: Major,
	},
	{
		name: "No checkboxes checked, should default to Patch",
		body: `[SYAC] Please check **one** to indicate the semantic versioning level of this release
- [ ] **Patch**
- [ ] **Minor**
- [ ] **Major**
`,
		expected: Patch,
	},
	{
		name:     "No SYAC prefix, should default to Patch",
		body:     `This is not a valid SYAC comment`,
		expected: Patch,
	},
	{
		name: "Case insensitive match for patch",
		body: `[SYAC] Please check **one** to indicate the semantic versioning level of this release
- [x] **patch** (*.*.x) - Bug fixes only, no breaking changes
- [ ] **Minor** (*.x.*)
- [ ] **Major** (x.*.*)
`,
		expected: Patch,
	},
}

func TestParseVersionNote(t *testing.T) {
	for _, tt := range parseVersionNoteTests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseVersionNote(tt.body)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}
