# xcute CLI ツール仕様書

## 概要

`xcute`は標準入力から読み込んだデータを使って、指定されたコマンドを実行するCLIツールです。標準入力の各行に対して、指定されたコマンドテンプレートの`{}`プレースホルダーを置換してコマンドを実行します。

## 基本仕様

### 言語・実装
- Go言語で実装
- コマンド名: `xcute`

### 基本動作
- 標準入力から一行ずつデータを読み込み
- 各行に対して指定されたコマンドを実行
- `{}`プレースホルダーに読み込んだ行の内容を埋め込み

### 基本的な使用例

```bash
# files.txtの各行をechoで出力
cat files.txt | xcute echo {}

# 複数のプレースホルダーを使用
cat words.txt | xcute echo hello {}, thanks {}
```

## 実行モード

### 1. 直接実行モード（デフォルト）
コマンドを直接実行します。空白文字が含まれる入力でも、引数として単一の値として扱われます。

**引数の扱い**:
- コマンドライン引数はそのまま使用され、文字列の結合や再解析は行われません
- 例: `xcute echo hello {}` では `["echo", "hello", "{}"]` という引数リストが使用されます

```bash
cat input.txt | xcute echo {}
cat input.txt | xcute echo hello {}
cat input.txt | xcute cp {} backup/
```

### 2. シェル実行モード（`-c`オプション）
シェルを通じてコマンドを実行します。複雑なシェルコマンドを実行できますが、空白文字の扱いに注意が必要です。

**引数の扱い**:
- `-c`オプション使用時は、その後の引数を1つの文字列として扱います
- `-c`オプション以降に複数の引数がある場合はエラーになります
- 正しい使用例: `xcute -c 'echo hello {}'`
- 誤った使用例: `xcute -c 'echo hello {}' extra_arg` （エラー）

```bash
cat words.txt | xcute -c 'echo hello {} && echo thanks {}'
cat files.txt | xcute -c 'wc -l {} | head -1'
```

**注意**: シェル実行モードでは、入力に空白文字が含まれる場合、ユーザーが適切にクォーテーションで囲む必要があります。

## コマンドライン引数の仕様

### 一般的な形式
```
xcute [OPTIONS] COMMAND_TEMPLATE
```

### 直接実行モード
```
xcute [OPTIONS] COMMAND ARG1 ARG2 ...
```
- `COMMAND`、`ARG1`、`ARG2`等の各引数はそのまま実行時に使用されます
- プレースホルダー`{}`を含む引数は、標準入力の各行で置換されます

### シェル実行モード
```
xcute [OPTIONS] -c 'SHELL_COMMAND'
```
- `-c`オプションの後には**1つの引数のみ**を指定できます
- その引数はシェルコマンドとして解釈されます
- `-c`オプション使用時に複数の引数が指定された場合はエラーとなります

### エラーとなるケース
```bash
# シェル実行モードで複数引数を指定
xcute -c 'echo {}' extra_argument  # エラー

# 引数なし
xcute                              # エラー
xcute -c                          # エラー
```

## オプション

### `-n` (dry run)
実際には実行せず、実行予定のコマンドラインを表示します。

```bash
cat input.txt | xcute -n rm {}
```

### `-i` (interactive)
各コマンド実行前に実行確認のプロンプトを表示します（`rm -i`のような動作）。

```bash
cat files.txt | xcute -i rm {}
```

### `-w` (show target)
各行の処理前に、対象となるファイル名（入力行の内容）を標準エラー出力に表示します。ANSI colorを使用して見やすく表示します。

```bash
cat files.txt | xcute -w cat {}
```

### `-l` (show command line)
各行の処理前に実行予定のコマンドラインを標準エラー出力に表示し、実行後にはステータスコードも表示します。ANSI colorを使用して見やすく表示します。

```bash
cat input.txt | xcute -l echo {}
```

### `-f` (force continue)
エラーが発生しても処理を継続します。デフォルトではエラーが発生すると処理が停止します。

- `-f`なし: 最初のエラーで停止、そのエラーのステータスコードで終了
- `-f`あり: エラーが発生しても継続、最後のエラーのステータスコードで終了（エラーがあった場合は0以外）

```bash
cat files.txt | xcute -f rm {}
```

### `-t <秒数>` (interval)
各行の処理間にインターバル（待機時間）を設定します。

```bash
cat urls.txt | xcute -t 1 curl {}
```

## 特殊な処理

### 空行の処理
- 入力が空行の場合、そのコマンドは実行されません
- `-w`または`-l`オプションが指定されている場合、空行であった旨を標準エラー出力に表示します

### エラーハンドリング
- デフォルト: エラー発生時に処理停止、該当エラーのステータスコードで終了
- `-f`オプション使用時: エラーが発生しても処理継続、最後のエラーのステータスコードで終了

### 出力の色付け
`-w`および`-l`オプションの標準エラー出力には色付けを行い、視認性を向上させます。

**使用している色**:
- シアン: 入力行の表示（`-w`オプション）
- ブルー: コマンドライン表示（`-l`オプション）
- グリーン: 正常終了のステータス表示
- レッド: エラー終了のステータス表示  
- イエロー: 空行の表示

**実装**: 色付け機能は[fatih/color](https://github.com/fatih/color)パッケージを使用しています。

## 使用例

```bash
# 基本的な使用
echo -e "file1.txt\nfile2.txt" | xcute cat {}

# 複数のプレースホルダー
echo -e "apple\nbanana" | xcute echo "I like {} and {}"

# シェルコマンドの実行
find . -name "*.txt" | xcute -c 'wc -l {} && echo "processed {}"'

# ドライラン
cat filelist.txt | xcute -n rm {}

# インタラクティブな実行
cat filelist.txt | xcute -i rm {}

# 詳細ログ付きでの実行
cat filelist.txt | xcute -l -w cp {} backup/

# エラーを無視して継続
cat filelist.txt | xcute -f rm {}

# インターバル付き実行
cat urls.txt | xcute -t 0.5 curl -s {}
```

## オプションの組み合わせ

複数のオプションを組み合わせて使用することができます。

```bash
cat files.txt | xcute -l -w -f -t 1 process {}
```
