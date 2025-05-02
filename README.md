# シンプルなWalk GUIアプリケーション

[Walk](https://github.com/lxn/walk)を使用して作成したシンプルなWindows GUIアプリケーションです。Walkは、Go言語用のWindows GUIツールキットです。

## 機能

- ボタン付きのウィンドウを表示
- ボタンをクリックするとメッセージボックスを表示

## 必要条件

- Go 1.24以降
- Windowsオペレーティングシステム（Walkはウィンドウズ専用です）
- GCCコンパイラ（CGo用）

## インストール方法

1. このリポジトリをクローンします：
   ```
   git clone https://github.com/yourusername/zip-editor.git
   cd zip-editor
   ```

2. 依存関係をインストールします：
   ```
   go mod tidy
   ```

## ビルドと実行

### 通常のビルド

```
go build
```

その後、実行ファイルを実行します：

```
zip-editor.exe
```

### より良いDPIサポートのためのマニフェストを使用したビルド

より良いDPIスケーリングサポートのために、rsrcツールを使用してビルドできます：

1. rsrcをインストールします：
   ```
   go install github.com/akavel/rsrc@latest
   ```

2. マニフェストファイルを作成します（このリポジトリにはすでに含まれています）

3. rsrcでビルドします：
   ```
   rsrc -manifest zip-editor.manifest -o rsrc.syso
   go build
   ```

## ライセンス

[MITライセンス](LICENSE)
