package fileops

import (
	"github.com/lxn/walk"
	"zip-editor/internal/model"
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
func UpdateFileList(tv *walk.TableView, treeItem *model.ZipTreeItem) error {

	// TableViewのモデルを設定
	fileModel := new(model.FileItemModel)

	// ポインタのスライスから値のスライスに変換
	ptrFiles := treeItem.GetFiles()
	valueFiles := make([]model.ZipTreeItem, len(ptrFiles))
	for i, file := range ptrFiles {
		if file != nil {
			valueFiles[i] = *file
		}
	}

	fileModel.Items = valueFiles
	tv.SetModel(fileModel)

	return nil
}
