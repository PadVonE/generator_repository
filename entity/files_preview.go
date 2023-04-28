package entity

type FilesPreview struct {
	FilePath string `json:"file_path"`
	NewCode  string `json:"new_code"`
	OldCode  string `json:"old_code"`
	HasFile  bool   `json:"has_file"`
	HasDiff  bool   `json:"has_diff"`
}
