package gitlab

type MergeRequestNote struct {
	Body string `json:"body"`
}

type Tag struct {
	Name string `json:"name"`
}
