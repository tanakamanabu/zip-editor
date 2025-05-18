package fileops

import (
	"archive/zip"
	"path/filepath"
	"strings"

	"github.com/lxn/walk"

	"zip-editor/internal/filemodel"
)

// deleteFlags は削除フラグの状態を保持するマップ
// キーはZIPファイルパスとファイルパスの組み合わせ
var deleteFlags = make(map[string]bool)

// getDeleteFlagKey はマップのキーを生成します
func getDeleteFlagKey(zipPath, filePath string) string {
	return zipPath + "::" + filePath
}

// GetDeleteFlag は指定されたファイルの削除フラグを取得します
func GetDeleteFlag(zipPath, filePath string) bool {
	key := getDeleteFlagKey(zipPath, filePath)
	return deleteFlags[key]
}

// SetDeleteFlag は指定されたファイルの削除フラグを設定します
func SetDeleteFlag(zipPath, filePath string, flag bool) {
	key := getDeleteFlagKey(zipPath, filePath)
	deleteFlags[key] = flag
}

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
			// 完全なファイルパスを作成
			fullPath := dir + name

			// 削除フラグの状態を取得
			deleteFlag := GetDeleteFlag(zipPath, fullPath)

			// ファイルアイテムをリストに追加
			items = append(items, filemodel.FileItem{
				Name:      name,
				Size:      int64(file.UncompressedSize64),
				Date:      file.Modified,
				DeleteFlag: deleteFlag,
			})
		}
	}

	// TableViewのモデルを設定
	model := new(filemodel.FileItemModel)
	model.Items = items
	tv.SetModel(model)

	return nil
}
