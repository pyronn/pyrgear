# pyrgear

A collection of commonly used Go utilities and tools.

## Installation

```bash
# As a dependency
go get github.com/pyronn/pyrgear

# As a command line tool
go install github.com/pyronn/pyrgear/cmd/pyrgear@latest
```

## Features

- Coming soon...

## Batch Rename Command

The `rename` command allows you to batch rename files in a specified directory based on a regular expression pattern.

### Usage

```bash
pyrgear rename --dir <directory> --pattern <regex_pattern> --replacement <replacement_pattern> [--recursive] [--dry-run]
```

或者使用预定义规则：

```bash
pyrgear rename --dir <directory> --rule <rule_name> [--recursive] [--dry-run]
```

### Options

- `--dir`: Directory to process (required)
- `--pattern`: Regular expression pattern to match filenames
- `--replacement`: Replacement pattern for new filenames
- `--recursive`: Process subdirectories recursively
- `--dry-run`: Show what would be renamed without actually renaming
- `--rule`: Predefined rule for renaming (e.g., 'timestamp', 'sequence', 'lowercase')

### Examples

1. Rename all files starting with "file_" followed by a number to "document_" followed by the same number:

```bash
pyrgear rename --dir ./my_files --pattern "file_(\d+)" --replacement "document_$1"
```

2. Preview renaming without actually changing files:

```bash
pyrgear rename --dir ./my_files --pattern "file_(\d+)" --replacement "document_$1" --dry-run
```

3. Process all subdirectories recursively:

```bash
pyrgear rename --dir ./my_files --pattern "file_(\d+)" --replacement "document_$1" --recursive
```

4. 使用预定义规则重命名文件：

```bash
# 添加时间戳前缀
pyrgear rename --dir ./my_files --rule timestamp

# 按顺序重命名文件（file_001.jpg, file_002.jpg, ...）
pyrgear rename --dir ./my_files --rule sequence

# 将所有文件名转换为小写
pyrgear rename --dir ./my_files --rule lowercase

# 导出微信小程序资源图片
pyrgear rename --rule wx-exporter --source-path "/path/to/project" --output-dir "./wx-images"
```

### 微信小程序资源导出 (wx-exporter)

`wx-exporter` 规则用于从微信小程序项目中提取资源图片，并按照特定格式重命名。

#### 工作原理

1. 扫描指定的 path1 目录下的所有子目录（作为 path2）
2. 在每个 path2 目录中查找 "assets" 文件夹
3. 提取 assets 文件夹中的图片文件（jpg, jpeg, png, gif, webp）
4. 将图片重命名为 `path2_序号` 格式（例如：page1_001.png, page1_002.png）
5. 将所有图片复制到指定的输出目录

注意：
- 用户只需指定顶层目录（path1），程序会自动处理其下所有子目录
- "assets" 是固定的文件夹名称，必须位于 path2 目录下
- 序号是按照每个 path2 分别计数的，不同 path2 的序号会分别从 1 开始

#### 目录结构示例

```
path1/                  <- 用户指定的源目录
  ├── page1/            <- path2 (自动检测)
  │   └── assets/       <- 固定文件夹名
  │       ├── img1.png
  │       └── img2.jpg
  └── page2/            <- path2 (自动检测)
      └── assets/       <- 固定文件夹名
          ├── icon.png
          └── bg.jpg
```

#### 选项

- `--source-path`: 源目录路径（path1，可选，默认为当前目录）
- `--output-dir`: 输出目录（可选，默认为 "wx-export"）
- `--dry-run`: 预览模式，不实际复制文件

#### 示例

```bash
# 从当前目录扫描并导出到默认的 wx-export 目录
pyrgear rename --rule wx-exporter

# 从指定目录扫描并导出到自定义目录
pyrgear rename --rule wx-exporter --source-path "/path/to/project" --output-dir "./images"

# 预览模式，不实际复制文件
pyrgear rename --rule wx-exporter --dry-run
```

## License

MIT License
