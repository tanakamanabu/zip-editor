package fileops

import (
	"archive/zip"
	"path/filepath"
	"strings"

	"github.com/lxn/walk"
)

// UpdateFileList は指定されたディレクトリ内のファイル一覧を更新します
func UpdateFileList(te *walk.TextEdit, zipPath string, dirPath string) error {
	if zipPath == "" {
		return nil
	}

	// ZIPファイルを開く
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	// テキストエディットをクリア
	te.SetText("")

	// ZIPの各ファイルを処理
	var fileList strings.Builder
	for _, file := range reader.File {
		// ディレクトリはスキップ
		if strings.HasSuffix(file.Name, "/") {
			continue
		}

		// ファイルパスをUTF-8に変換
		path := AutoDetectEncoding(file.Name)
		dir := filepath.Dir(path)
		name := filepath.Base(path)

		// 比較のためにディレクトリパスを正規化
		dir = strings.TrimSuffix(dir, "/")
		if dir == "." {
			dir = ""
		} else {
			dir += "/"
		}

		// このファイルが現在のディレクトリにあるかチェック
		if dir == dirPath {
			// ファイルをリストに追加
			fileList.WriteString(name)
			fileList.WriteString("\r\n")
		}
	}

	// テキストエディットを更新
	te.SetText(fileList.String())

	return nil
}
