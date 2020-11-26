package utils

import (
	"crypto/md5"
	"io"
	"os"
)

func GetMd5FromFile(path string) (md5string string, err error)  {
	file, err := os.Open(path)

	if err != nil {
		return "", err
	}

	defer file.Close()

	hash := md5.New()
	_, err = io.Copy(hash, file)

	if err != nil {
		return "", err
	}

	return string(hash.Sum(nil)), nil
}

func GetFileSizeInByte(path string) (fileSize int64, err error) {
	fi, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	// get the size
	size := fi.Size()
	return size, nil
}
