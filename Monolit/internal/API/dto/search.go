package dto

type SearchResponse struct {
	Calls        []SearchCallResponse        `json:"calls"`
	Companies    []SearchCompanyResponse     `json:"companies"`
	Reports      []SearchReportResponse      `json:"reports"`
	Instructions []SearchInstructionResponse `json:"instructions"`
}

type SearchCallResponse struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

type SearchCompanyResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type SearchReportResponse struct {
	ID       string `json:"id"`
	CallUUID string `json:"call_uuid"`
	FileName string `json:"file_name"`
	Status   string `json:"status"`
}

type SearchInstructionResponse struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Scope string `json:"scope"`
}
