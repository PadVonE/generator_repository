package entity

type TaskInfo struct {
	IssueKey    string `json:"issue_key"`
	Summary     string `json:"summary"`
	AssigneeImg string `json:"assignee_img"`
	Status      string `json:"status"`
}
