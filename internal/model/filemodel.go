package model

import (
	"fmt"
	"github.com/lxn/walk"
)

// FileItemModel はTableView用のモデルを表します
type FileItemModel struct {
	walk.TableModelBase
	Items []ZipTreeItem
}

// SetValue は指定された行と列の値を設定します
func (m *FileItemModel) SetValue(row, col int, value interface{}) error {
	if row < 0 || row >= len(m.Items) {
		return nil
	}

	switch col {
	case 0:
		if val, ok := value.(bool); ok {
			m.Items[row].DeleteFlag = val
			return nil
		}
	}

	return nil
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
		return item.DeleteFlag
	case 1:
		return item.GetName()
	case 2:
		// サイズをKBに変換し、カンマ区切りで表示
		sizeKB := float64(item.GetSize()) / 1024.0
		if sizeKB < 0.1 {
			return "0.1 KB" // 最小表示サイズ
		}
		return formatWithCommas(sizeKB) + " KB"
	case 3:
		return item.GetDate().Format("2006/01/02 15:04:05")
	}

	return nil
}

// ColumnCount はカラム数を返します
func (m *FileItemModel) ColumnCount() int {
	return 4
}

// ColumnName は指定された列の名前を返します
func (m *FileItemModel) ColumnName(col int) string {
	switch col {
	case 0:
		return "" // チェックボックス用の空の列名
	case 1:
		return "ファイル名"
	case 2:
		return "サイズ"
	case 3:
		return "日付"
	}
	return ""
}

// formatWithCommas は数値をカンマ区切りでフォーマットします
func formatWithCommas(num float64) string {
	// 整数部と小数部に分ける
	intPart := int(num)
	fracPart := num - float64(intPart)

	// 整数部をカンマ区切りにする
	result := ""

	// 0の場合は特別処理
	if intPart == 0 {
		result = "0"
	} else {
		for intPart > 0 {
			if len(result) > 0 {
				result = "," + result
			}
			remainder := intPart % 1000
			if intPart < 1000 {
				result = fmt.Sprintf("%d", remainder) + result
			} else {
				result = fmt.Sprintf("%03d", remainder) + result
			}
			intPart /= 1000
		}
	}

	// 小数部がある場合は追加
	if fracPart > 0.01 {
		return fmt.Sprintf("%s.%d", result, int(fracPart*10))
	}

	return result
}
