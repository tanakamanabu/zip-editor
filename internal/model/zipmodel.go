package model

import (
	"archive/zip"
	"path/filepath"
	"strings"

	"github.com/lxn/walk"

	"zip-editor/internal/fileops"
)

// ZipTreeItem はZIPファイルツリー内のアイテムを表します
type ZipTreeItem struct {
	name       string
	path       string
	children   []*ZipTreeItem
	parent     *ZipTreeItem
	isDir      bool
	deleteFlag bool
}

// GetName は名前を返します
func (item *ZipTreeItem) GetName() string {
	return item.name
}

// GetPath はパスを返します
func (item *ZipTreeItem) GetPath() string {
	return item.path
}

// IsDir はディレクトリかどうかを返します
func (item *ZipTreeItem) IsDir() bool {
	return item.isDir
}

// GetDeleteFlag は削除フラグを返します
func (item *ZipTreeItem) GetDeleteFlag() bool {
	return item.deleteFlag
}

// SetDeleteFlagRecursively はこのアイテムとすべての子に削除フラグを設定します
func (item *ZipTreeItem) SetDeleteFlagRecursively(flag bool, model *ZipTreeModel) {
	item.deleteFlag = flag
	model.PublishItemChanged(item)

	// 再帰的にすべての子にフラグを設定
	for _, child := range item.children {
		child.SetDeleteFlagRecursively(flag, model)
	}
}

// Text は表示テキストを返します
func (item *ZipTreeItem) Text() string {
	if item.deleteFlag {
		return item.name + " [削除予定]"
	}
	return item.name
}

// Parent は親アイテムを返します
func (item *ZipTreeItem) Parent() walk.TreeItem {
	if item.parent == nil {
		return nil
	}
	return item.parent
}

// ChildCount は子の数を返します
func (item *ZipTreeItem) ChildCount() int {
	return len(item.children)
}

// ChildAt は指定されたインデックスの子を返します
func (item *ZipTreeItem) ChildAt(index int) walk.TreeItem {
	return item.children[index]
}

// Image はアイテムの画像インデックスを返します
func (item *ZipTreeItem) Image() interface{} {
	if item.isDir {
		return 0 // フォルダアイコン
	}
	return 1 // ファイルアイコン
}

// ZipTreeModel はZIPファイルのツリーモデルを表します
type ZipTreeModel struct {
	walk.TreeModelBase
	rootItem *ZipTreeItem
}

// PublishItemChanged はアイテムが変更されたことを通知します
// 注意: この実装は簡略化されており、実際のイベント通知は行われません
func (m *ZipTreeModel) PublishItemChanged(item walk.TreeItem) {
	// 実際のアプリケーションでは、ここでUIに変更を通知する必要があります
	// 現在の実装では、この機能は使用されていません
}

// LazyPopulation はツリーを即時に展開するためfalseを返します
func (m *ZipTreeModel) LazyPopulation() bool {
	return false
}

// RootCount はルートアイテムの数を返します
func (m *ZipTreeModel) RootCount() int {
	return 1
}

// RootAt は指定されたインデックスのルートアイテムを返します
func (m *ZipTreeModel) RootAt(index int) walk.TreeItem {
	return m.rootItem
}

// LoadZipFile はZIPファイルを読み込み、ツリーモデルを作成します
func LoadZipFile(filePath string) (*ZipTreeModel, error) {
	reader, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	// ZIPファイル名でルートアイテムを作成
	rootItem := &ZipTreeItem{
		name:  filepath.Base(filePath),
		path:  "",
		isDir: true,
	}

	// ディレクトリアイテムを素早く検索するためのマップ
	dirMap := make(map[string]*ZipTreeItem)
	dirMap[""] = rootItem

	// ZIPの各ファイルを処理
	for _, file := range reader.File {
		// ディレクトリはスキップ（必要に応じて作成）
		if strings.HasSuffix(file.Name, "/") {
			continue
		}

		// パスをコンポーネントに分割し、エンコーディングを自動検出
		// これは様々なエンコーディング（Shift-JIS、EUC-JP、UTF-8など）を試し、UTF-8に変換します
		path := fileops.AutoDetectEncoding(file.Name)
		dir := filepath.Dir(path)
		dir = strings.TrimSuffix(dir, "/")

		// すべての親ディレクトリが存在することを確認
		parentPath := ""
		parentItem := rootItem
		for _, part := range strings.Split(dir, "/") {
			if part == "" {
				continue
			}

			currentPath := parentPath + part + "/"
			if item, exists := dirMap[currentPath]; exists {
				parentItem = item
			} else {
				// 新しいディレクトリアイテムを作成
				newDir := &ZipTreeItem{
					name:   part,
					path:   currentPath,
					parent: parentItem,
					isDir:  true,
				}
				parentItem.children = append(parentItem.children, newDir)
				dirMap[currentPath] = newDir
				parentItem = newDir
			}
			parentPath = currentPath
		}

		// ツリーにはフォルダのみを表示するため、ファイルアイテムは追加しない
		// ファイルアイテムはparentItem.childrenに追加されない
	}

	return &ZipTreeModel{rootItem: rootItem}, nil
}
