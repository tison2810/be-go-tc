// package models

// // RunSpec là cấu trúc dữ liệu cho một yêu cầu chạy code trên Jobe
// type FileEntry struct {
// 	Filename string `json:"filename"`
// 	Content  string `json:"content"`
// }
// type RunSpec struct {
// 	Language_id    string                 `json:"language_id"`
// 	Sourcecode     string                 `json:"sourcecode,omitempty"`
// 	Sourcefilename string                 `json:"sourcefilename,omitempty"`
// 	FileList       []FileEntry            `json:"file_list,omitempty"`
// 	Parameters     map[string]interface{} `json:"parameters,omitempty"`
// 	Maxruntime     int                    `json:"maxruntime,omitempty"`
// 	Memorylimit    int                    `json:"memorylimit,omitempty"`
// 	Cmpinfo        string                 `json:"cmpinfo,omitempty"`
// }

// // JobeRunRequest là payload chính gửi đến Jobe
// type JobeRunRequest struct {
// 	RunSpec RunSpec `json:"run_spec"`
// }

// // JobeRunResponse là phản hồi từ Jobe sau khi chạy code
//
//	type JobeRunResponse struct {
//		Outcome int    `json:"outcome"`
//		Stdout  string `json:"stdout,omitempty"`
//		Stderr  string `json:"stderr,omitempty"`
//		Cmpinfo string `json:"cmpinfo,omitempty"`
//	}
package models

// FileEntry đại diện cho một file được gửi lên Jobe
// type FileEntry struct {
// 	Filename string `json:"filename"`
// 	Content  string `json:"content"`
// }

// RunParameters chứa các tham số khi chạy chương trình
// type RunParameters struct {
// 	RunArgs     []string `json:"runargs,omitempty"`
// 	CompileArgs []string `json:"compileargs,omitempty"`
// 	LinkArgs    []string `json:"linkargs,omitempty"`
// 	CPUTime     int      `json:"cputime,omitempty"`
// 	MemoryLimit int      `json:"memorylimit,omitempty"`
// }

// // RunSpec chứa thông tin về đoạn code cần chạy trên Jobe
// type RunSpec struct {
// 	LanguageID     string        `json:"language_id"`
// 	SourceCode     string        `json:"sourcecode,omitempty"`
// 	SourceFilename string        `json:"sourcefilename,omitempty"`
// 	FileList       []string      `json:"file_list,omitempty"`
// 	Parameters     RunParameters `json:"parameters,omitempty"`
// 	MaxRuntime     int           `json:"maxruntime,omitempty"`
// 	MemoryLimit    int           `json:"memorylimit,omitempty"`
// 	CmpInfo        string        `json:"cmpinfo,omitempty"`
// }

// // JobeRunRequest là payload gửi lên Jobe để thực thi code
// type JobeRunRequest struct {
// 	RunSpec RunSpec `json:"run_spec"`
// }

// // JobeRunResponse là phản hồi từ Jobe sau khi chạy code
// type JobeRunResponse struct {
// 	Outcome    int     `json:"outcome"`
// 	Stdout     string  `json:"stdout,omitempty"`
// 	Stderr     string  `json:"stderr,omitempty"`
// 	CmpInfo    string  `json:"cmpinfo,omitempty"`
// 	ExitCode   int     `json:"exit_code,omitempty"`
// 	TimeUsed   float64 `json:"time_used,omitempty"`
// 	MemoryUsed int     `json:"memory_used,omitempty"`
// }

// type PostFileRequest struct {
// 	FileContents string `json:"file_contents"`
// }

// type PostFileResponse struct {
// 	FileID string `json:"file_id"`
// }

type RunSpec struct {
	LanguageID string      `json:"language_id"`
	SourceCode string      `json:"sourcecode"`
	SourceFile string      `json:"sourcefilename"`
	FileList   [][2]string `json:"file_list"`
	Parameters interface{} `json:"parameters,omitempty"`
}

// Cấu trúc phản hồi từ Jobe
type JobeResponse struct {
	Outcome int    `json:"outcome"`
	Output  string `json:"stdout"`
	Error   string `json:"stderr"`
}
type UploadResponse struct {
	FileID string `json:"file_id"`
}
