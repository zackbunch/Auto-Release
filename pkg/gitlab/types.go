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
