package filemodel

import (
	"time"

	"github.com/lxn/walk"
)

// FileItem はZIPファイル内のファイルアイテムを表します
type FileItem struct {
	Name string
	Size int64
	Date time.Time
}

// FileItemModel はTableView用のモデルを表します
type FileItemModel struct {
	walk.TableModelBase
	Items []FileItem
}

// RowCount は行数を返します
func (m *FileItemModel) RowCount() int {
	return len(m.Items)
}

// Value は指定された行と列の値を返します
func (m *FileItemModel) Value(row, col int) interface{} {
	if row < 0 || row >= len(m.Items) {
		return nil
	}

	item := m.Items[row]

	switch col {
	case 0:
		return item.Name
	case 1:
		return item.Size
	case 2:
		return item.Date.Format("2006/01/02 15:04:05")
	}

	return nil
}

// ColumnCount はカラム数を返します
func (m *FileItemModel) ColumnCount() int {
	return 3
}

// ColumnName は指定された列の名前を返します
func (m *FileItemModel) ColumnName(col int) string {
	switch col {
	case 0:
		return "ファイル名"
	case 1:
		return "サイズ"
	case 2:
		return "日付"
	}
	return ""
}
