package entity

type ViewData struct {
	Title string
}

type ContentData struct {
	ThemeId int

	DomainName  string
	NewsId      int
	Slug        string
	QueryIds    string
	QueryNewsId string

	// for rotation script
	GeoCode string
	Service int
	Device  int
	Stream  int
	Safe    int

	// For News
	Limit  int
	Offset int

	// Connection
}
