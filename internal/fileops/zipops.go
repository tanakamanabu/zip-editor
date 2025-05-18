package fileops

import (
	"archive/zip"
	"path/filepath"
	"strings"

	"github.com/lxn/walk"

	"zip-editor/internal/filemodel"
)

// UpdateFileList は指定されたディレクトリ内のファイル一覧を更新します
func UpdateFileList(tv *walk.TableView, zipPath string, dirPath string) error {
	if zipPath == "" {
		return nil
	}

	// ZIPファイルを開く
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	// ZIPの各ファイルを処理
	var items []filemodel.FileItem
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
			// ファイルアイテムをリストに追加
			items = append(items, filemodel.FileItem{
				Name: name,
				Size: int64(file.UncompressedSize64),
				Date: file.Modified,
			})
		}
	}

	// TableViewのモデルを設定
	model := new(filemodel.FileItemModel)
	model.Items = items
	tv.SetModel(model)

	return nil
}
