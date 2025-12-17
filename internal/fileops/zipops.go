package fileops

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"zip-editor/internal/common"
	"zip-editor/internal/model"

	"github.com/lxn/walk"
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

// ExtractFileToTemp は指定したZIP内の単一ファイルを一時ディレクトリに展開し、そのパスを返します
// エンコーディングは自動検出し、UTF-8のパス（model側と同一ロジック）でマッチングします
func ExtractFileToTemp(zipPath, entryUTF8Path string) (string, error) {
	// 一時ディレクトリを作成（規約に従いプレフィックスを使用）
	tempDir, err := os.MkdirTemp("", "zip-editor-")
	if err != nil {
		return "", err
	}

	// ZIPを開く
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	// エントリを探索
	var target *zip.File
	for _, f := range reader.File {
		// ZIP内パスのエンコーディングをUTF-8へ
		utf8Path := common.AutoDetectEncoding(f.Name)
		if utf8Path == entryUTF8Path {
			target = f
			break
		}
	}
	if target == nil {
		return "", os.ErrNotExist
	}

	// 出力先フルパス（Zip内のサブディレクトリ構造を維持）
	rel := filepath.FromSlash(entryUTF8Path)
	// 先頭にスラッシュがあれば削除
	rel = strings.TrimLeft(rel, "\\/")
	outPath := filepath.Join(tempDir, rel)

	// 親ディレクトリを作成
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return "", err
	}

	// ファイルを展開
	rc, err := target.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	outFile, err := os.Create(outPath)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, rc); err != nil {
		return "", err
	}

	// 正常終了
	return outPath, nil
}
