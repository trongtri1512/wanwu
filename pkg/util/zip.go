package util

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func DirToBytes(dir string) ([]byte, error) {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	defer func() { _ = zipWriter.Close() }()

	// 检查路径是否存在
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, fmt.Errorf("路径不存在: %s", dir)
	}

	// 获取顶层目录名（用于 ZIP 内的根目录）
	dirName := filepath.Base(dir)

	err := filepath.Walk(dir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 计算相对于源目录的路径
		relPath, err := filepath.Rel(dir, filePath)
		if err != nil {
			return err
		}

		// 构建 ZIP 内的路径：dirName/relPath
		zipPath := filepath.ToSlash(filepath.Join(dirName, relPath))
		if info.IsDir() {
			zipPath += "/" // 目录必须以 / 结尾
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = zipPath
		// 指明使用store只存储不使用压缩算法
		header.Method = zip.Store

		// 创建 ZIP 条目
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		// 如果是文件，写入内容
		if !info.IsDir() {
			file, err := os.Open(filePath)
			if err != nil {
				return err
			}
			_, err = io.Copy(writer, file)
			_ = file.Close()
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 必须 Close 才会 flush 数据到 buf
	if err := zipWriter.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
