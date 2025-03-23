package model

// UploadFileReq 是用于文件上传请求的结构体。
// 它封装了文件上传所需的所有参数，便于在上传过程中跟踪和管理文件的各个片段及其元数据。
type UploadFileReq struct {
	// TaskCode 是文件上传任务的唯一标识码。
	// 用于在上传过程中标识特定的上传任务。
	TaskCode string `form:"taskCode"`
	// ProjectCode 是项目的唯一标识码。
	// 用于指示文件属于哪个项目，以便在上传后正确归档。
	ProjectCode string `form:"projectCode"`
	// ProjectName 是项目的名称。
	// 提供额外的项目信息，便于在上传过程中进行参考。
	ProjectName string `form:"projectName"`
	// TotalChunks 表示文件被分割成的总片段数。
	// 文件可能被分割成多个片段分别上传，此字段指示总共有多少个片段。
	TotalChunks int `form:"totalChunks"`
	// RelativePath 是文件相对于项目目录的路径。
	// 用于指示文件上传后应存储的相对位置。
	RelativePath string `form:"relativePath"`
	// Filename 是上传文件的名称。
	// 用于在上传过程中标识文件，并在上传完成后进行命名。
	Filename string `form:"filename"`
	// ChunkNumber 表示当前上传片段的编号。
	// 与TotalChunks结合使用，以跟踪上传进度。
	ChunkNumber int `form:"chunkNumber"`
	// ChunkSize 是除了最后一个片段外，每个上传片段的大小。
	// 用于在上传过程中保持片段大小的一致性。
	ChunkSize int `form:"chunkSize"`
	// CurrentChunkSize 表示当前上传片段的实际大小。
	// 可能会与ChunkSize不同，特别是对于最后一个片段。
	CurrentChunkSize int `form:"currentChunkSize"`
	// TotalSize 是上传文件的总大小。
	// 用于验证上传过程中的数据完整性和一致性。
	TotalSize int `form:"totalSize"`
	// Identifier 是文件上传的唯一标识符。
	// 用于在上传过程中唯一地标识和跟踪文件上传会话。
	Identifier string `form:"identifier"`
}
