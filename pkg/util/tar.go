package util

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func TarDir(srcDir string) ([]byte, error) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	srcDir = filepath.Clean(srcDir)
	baseName := filepath.Base(srcDir)

	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return fmt.Errorf("get relative path failed: %w", err)
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return fmt.Errorf("create tar header failed: %w", err)
		}

		if relPath == "." {
			header.Name = baseName + "/"
		} else {
			header.Name = baseName + "/" + filepath.ToSlash(relPath)
		}

		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("write tar header failed: %w", err)
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("open file failed: %w", err)
		}
		_, err = io.Copy(tw, file)
		_ = file.Close()
		if err != nil {
			return fmt.Errorf("write file content failed: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("walk directory failed: %w", err)
	}

	if err := tw.Close(); err != nil {
		return nil, fmt.Errorf("close tar writer failed: %w", err)
	}
	if err := gw.Close(); err != nil {
		return nil, fmt.Errorf("close gzip writer failed: %w", err)
	}

	return buf.Bytes(), nil
}

func TarFile(srcFile string) ([]byte, error) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	info, err := os.Stat(srcFile)
	if err != nil {
		return nil, fmt.Errorf("stat file failed: %w", err)
	}

	file, err := os.Open(srcFile)
	if err != nil {
		return nil, fmt.Errorf("open file failed: %w", err)
	}
	defer func() { _ = file.Close() }()

	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return nil, fmt.Errorf("create tar header failed: %w", err)
	}
	header.Name = filepath.Base(srcFile)

	if err := tw.WriteHeader(header); err != nil {
		return nil, fmt.Errorf("write tar header failed: %w", err)
	}

	if _, err := io.Copy(tw, file); err != nil {
		return nil, fmt.Errorf("write file content failed: %w", err)
	}

	if err := tw.Close(); err != nil {
		return nil, fmt.Errorf("close tar writer failed: %w", err)
	}
	if err := gw.Close(); err != nil {
		return nil, fmt.Errorf("close gzip writer failed: %w", err)
	}

	return buf.Bytes(), nil
}

func Untar(tarData []byte, destDir string) error {
	gr, err := gzip.NewReader(bytes.NewReader(tarData))
	if err != nil {
		return fmt.Errorf("create gzip reader failed: %w", err)
	}
	defer func() { _ = gr.Close() }()

	tr := tar.NewReader(gr)
	destDir = filepath.Clean(destDir)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar header failed: %w", err)
		}

		parts := strings.Split(header.Name, "/")
		if len(parts) > 1 {
			header.Name = strings.Join(parts[1:], "/")
		}
		if header.Name == "" {
			continue
		}

		target := filepath.Join(destDir, header.Name)
		if !strings.HasPrefix(filepath.Clean(target), destDir+string(os.PathSeparator)) {
			return fmt.Errorf("invalid path: %s (traversal attempt)", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("create directory failed: %w", err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("create parent directory failed: %w", err)
			}
			file, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("create file failed: %w", err)
			}
			if _, err := io.Copy(file, tr); err != nil {
				_ = file.Close()
				_ = os.Remove(target)
				return fmt.Errorf("write file content failed: %w", err)
			}
			_ = file.Close()
		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("create parent directory failed: %w", err)
			}
			_ = os.Remove(target)
			if err := os.Symlink(header.Linkname, target); err != nil {
				return fmt.Errorf("create symlink failed: %w", err)
			}
		}
	}

	return nil
}
