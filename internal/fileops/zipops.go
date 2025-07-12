package fileops

import (
	"archive/zip"
	"github.com/lxn/walk"
	"io"
	"os"
	"path/filepath"
	"zip-editor/internal/common"
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
func setDeleteFlag(zipPath, filePath string, flag bool) {
	key := getDeleteFlagKey(zipPath, filePath)
	deleteFlags[key] = flag
}

func UpdateDeleteFlagRecursively(currentZipPath string, item *model.ZipTreeItem) {
	// 自分の削除フラグ設定
	setDeleteFlag(currentZipPath, item.GetPath(), item.DeleteFlag)

	// 自分の持ってるファイルの削除フラグを全部設定
	for _, file := range item.GetFiles() {
		file.DeleteFlag = item.DeleteFlag
		setDeleteFlag(currentZipPath, file.GetPath(), file.DeleteFlag)
	}

	// 再帰的にすべての子にフラグを設定
	for _, child := range item.GetChildren() {
		child.DeleteFlag = item.DeleteFlag
		UpdateDeleteFlagRecursively(currentZipPath, child)
	}
}

// DeleteFlaggedFiles は削除フラグが付いたファイルをZIPファイルから削除します
func DeleteFlaggedFiles(zipPath string) error {
	// 一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "zip-editor-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir) // 関数終了時に一時ディレクトリを削除

	// 一時ZIPファイルのパスを生成
	tempZipPath := filepath.Join(tempDir, "temp.zip")

	// 元のZIPファイルを開く
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}

	// 新しいZIPファイルを作成
	newZipFile, err := os.Create(tempZipPath)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	// 新しいZIPライターを作成
	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	// 元のZIPファイルの各ファイルを処理
	for _, file := range reader.File {
		// ファイルパスをUTF-8に変換
		path := common.AutoDetectEncoding(file.Name)

		// 削除フラグをチェック
		if GetDeleteFlag(zipPath, path) {
			continue // 削除フラグが付いているファイルはスキップ
		}

		// 元のファイルを開く
		fileReader, err := file.Open()
		if err != nil {
			return err
		}

		// 新しいZIPファイルにファイルを追加
		header := &zip.FileHeader{
			Name:   file.Name,
			Method: file.Method,
		}
		header.SetModTime(file.Modified)

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			fileReader.Close()
			return err
		}

		// ファイルの内容をコピー
		_, err = io.Copy(writer, fileReader)
		fileReader.Close()
		if err != nil {
			return err
		}
	}

	// ZIPライターを閉じる
	err = zipWriter.Close()
	if err != nil {
		return err
	}

	// 新しいZIPファイルを閉じる
	err = newZipFile.Close()
	if err != nil {
		return err
	}

	// 元のZIPファイルを閉じる
	reader.Close()

	// 元のZIPファイルを削除
	err = os.Remove(zipPath)
	if err != nil {
		return err
	}

	// 一時ZIPファイルを元の場所にコピー
	err = copyFile(tempZipPath, zipPath)
	if err != nil {
		return err
	}

	// コピー元の一時ファイルを削除
	err = os.Remove(tempZipPath)
	if err != nil {
		// コピーは成功しているので、一時ファイルの削除に失敗しても処理は続行
		// ただし、ログに記録するなどの対応が望ましい
	}

	return nil
}

// copyFile はファイルをソースからデスティネーションにコピーします
func copyFile(src, dst string) error {
	// ソースファイルを開く
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// デスティネーションファイルを作成
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// ソースからデスティネーションにコピー
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// ファイルをフラッシュして確実にディスクに書き込む
	err = destFile.Sync()
	if err != nil {
		return err
	}

	return nil
}

// UpdateFileList は指定されたディレクトリ内のファイル一覧を更新します
func UpdateFileList(tv *walk.TableView, treeItem *model.ZipTreeItem) error {

	// TableViewのモデルを設定
	fileModel := new(model.FileItemModel)

	// ポインタのスライスをそのまま使用
	fileModel.Items = treeItem.GetFiles()
	tv.SetModel(fileModel)

	return nil
}
