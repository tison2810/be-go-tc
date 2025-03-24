package models

type FileUploadResponse struct {
	Success bool   `json:"success"`
	FileID  string `json:"file_id,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

type UploadFileRequest struct {
	FileContents string `json:"file_contents"`
}

type RunSpec struct {
	LanguageID     string          `json:"language_id"`              // Bắt buộc: ID của ngôn ngữ (cpp, python, java, v.v.)
	SourceCode     string          `json:"sourcecode"`               // Bắt buộc: Mã nguồn của chương trình
	SourceFilename string          `json:"sourcefilename,omitempty"` // Tùy chọn: Tên file của mã nguồn
	Input          string          `json:"input,omitempty"`          // Tùy chọn: Dữ liệu đầu vào (stdin)
	FileList       []FileListEntry `json:"file_list,omitempty"`      // Tùy chọn: Danh sách các file (file_id, file_name)
	Parameters     interface{}     `json:"parameters,omitempty"`     // Tùy chọn: Các tham số (JSON object)
	Debug          bool            `json:"debug,omitempty"`          // Tùy chọn: Bật debug mode
}

// FileListEntry biểu diễn một cặp (file_id, file_name, is_source) trong file_list
type FileListEntry struct {
	FileID   string `json:"file_id"`
	FileName string `json:"file_name"`
	IsSource bool   `json:"is_source,omitempty"` // Tùy chọn: true nếu file là mã nguồn cần biên dịch
}

// SubmitRunRequest biểu diễn dữ liệu gửi tới Jobe server
type SubmitRunRequest struct {
	RunSpec RunSpec `json:"run_spec"`
}

// SubmitRunResponse biểu diễn response từ Jobe server
type SubmitRunResponse struct {
	Status int    `json:"status"`
	Result string `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
	Score  int    `json:"score,omitempty"`
	Log    string `json:"log,omitempty"`
}

type JobeRunResult struct {
	Stdout  string `json:"stdout"`
	Stderr  string `json:"stderr"`
	Cmpinfo string `json:"cmpinfo"`
	Outcome string `json:"outcome"`
}
