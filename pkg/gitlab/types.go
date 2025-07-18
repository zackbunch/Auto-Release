package gitlab

type MergeRequestNote struct {
	Body string `json:"body"`
}

type Tag struct {
	Name string `json:"name"`
}

type MergeRequest struct {
	IID   int    `json:"iid"`
	Title string `json:"title"`
}

type ProtectedBranch struct {
	Name string `json:"name"`
}

type Release struct {
	TagName     string `json:"tag_name"`              // e.g. v1.2.3
	Name        string `json:"name"`                  // e.g. "Release v1.2.3"
	Description string `json:"description,omitempty"` // release notes or changelog
	CreatedAt   string `json:"created_at,omitempty"`  // ISO 8601 UTC timestamp

	Ref          string `json:"ref,omitempty"`        // commit SHA or branch/tag
	Committer    string `json:"committer,omitempty"`  // name/email if available
	IsDraft      bool   `json:"draft,omitempty"`      // pre-release state
	IsPreRelease bool   `json:"prerelease,omitempty"` // semantic pre-release flag (v1.2.3-beta.1)
}
