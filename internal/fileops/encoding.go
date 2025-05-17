package fileops

import (
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// AutoDetectEncoding はエンコーディングを自動検出してUTF-8に変換するヘルパー関数です
func AutoDetectEncoding(input string) string {
	// 入力が既に有効なUTF-8かどうかをチェック
	if utf8.ValidString(input) {
		return input
	}

	// 試すエンコーディングのリスト
	encodings := []encoding.Encoding{
		japanese.ShiftJIS,       // 日本語 Shift-JIS
		japanese.EUCJP,          // 日本語 EUC-JP
		japanese.ISO2022JP,      // 日本語 ISO-2022-JP
		korean.EUCKR,            // 韓国語 EUC-KR
		simplifiedchinese.GBK,   // 簡体字中国語 GBK
		traditionalchinese.Big5, // 繁体字中国語 Big5
		charmap.Windows1252,     // Windows-1252（西ヨーロッパ）
		unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM),    // UTF-16BE
		unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM), // UTF-16LE
	}

	// 各エンコーディングを試す
	for _, enc := range encodings {
		decoder := enc.NewDecoder()
		output, _, err := transform.String(decoder, input)
		if err == nil && utf8.ValidString(output) {
			// 出力に有効な文字が含まれているかチェック
			// これは誤検出をフィルタリングするのに役立ちます
			if !containsControlCharacters(output) {
				return output
			}
		}
	}

	// すべてのエンコーディングが失敗した場合、フォールバックとしてShift-JISを試す（後方互換性のため）
	transformer := japanese.ShiftJIS.NewDecoder()
	output, _, err := transform.String(transformer, input)
	if err == nil {
		return output
	}

	// すべてが失敗した場合、元の文字列を返す
	return input
}

// containsControlCharacters は文字列に制御文字が含まれているかをチェックするヘルパー関数です
// これは不正なエンコーディング検出を示す可能性があります
func containsControlCharacters(s string) bool {
	for _, r := range s {
		// 一般的な空白を除く制御文字をチェック
		if r < 32 && r != '\t' && r != '\n' && r != '\r' {
			return true
		}
	}
	return false
}