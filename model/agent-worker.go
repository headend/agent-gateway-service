package model

type WorkerUpdateSignal struct {
	FileName	string	`json:"fileName"`
	FilePath	string	`json:"filePath"`
	FileSizeInByte	int64	`json:"fileSizeInByte"`
	Md5	string	`json:"md5"`
}
