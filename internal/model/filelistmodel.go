package model

import (
    "path/filepath"
    "github.com/lxn/walk"
)

// FileListModel は左ペインのZIPファイル一覧用のモデルです（ファイル名のみ表示、内部的にはフルパスを保持）。
type FileListModel struct {
    walk.TableModelBase
    paths    []string          // フルパスを保持
    deleting map[string]bool   // キー: フルパス, 値: 削除中かどうか
}

// NewFileListModel は空のモデルを返します。
func NewFileListModel() *FileListModel {
    return &FileListModel{paths: []string{}, deleting: make(map[string]bool)}
}

// RowCount は行数を返します。
func (m *FileListModel) RowCount() int {
    return len(m.paths)
}

// ColumnCount はカラム数を返します。
func (m *FileListModel) ColumnCount() int {
    return 1
}

// ColumnName は列名を返します。
func (m *FileListModel) ColumnName(col int) string {
    if col == 0 {
        return "ファイル"
    }
    return ""
}

// Value は指定された行と列の値を返します。
func (m *FileListModel) Value(row, col int) interface{} {
    if row < 0 || row >= len(m.paths) {
        return nil
    }
    switch col {
    case 0:
        base := filepath.Base(m.paths[row])
        if m.deleting[m.paths[row]] {
            return base + "（削除中）"
        }
        return base
    }
    return nil
}

// PathAt は指定行のフルパスを返します。
func (m *FileListModel) PathAt(row int) string {
    if row < 0 || row >= len(m.paths) {
        return ""
    }
    return m.paths[row]
}

// AddPath はパスを一覧に追加します（重複は無視）。
func (m *FileListModel) AddPath(p string) {
    for _, ex := range m.paths {
        if ex == p {
            return
        }
    }
    m.paths = append(m.paths, p)
    m.PublishRowsReset()
}

// IndexOfPath は指定パスの行インデックスを返します（見つからない場合は-1）。
func (m *FileListModel) IndexOfPath(p string) int {
    for i, v := range m.paths {
        if v == p {
            return i
        }
    }
    return -1
}

// IsDeleting は指定パスが削除中かどうかを返します。
func (m *FileListModel) IsDeleting(p string) bool {
    return m.deleting[p]
}

// SetDeleting は指定パスの削除中状態を設定し、行更新を通知します。
func (m *FileListModel) SetDeleting(p string, d bool) {
    if m.deleting == nil {
        m.deleting = make(map[string]bool)
    }
    m.deleting[p] = d
    if row := m.IndexOfPath(p); row >= 0 {
        m.PublishRowChanged(row)
    } else {
        // 念のため全体再描画
        m.PublishRowsReset()
    }
}
