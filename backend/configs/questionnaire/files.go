package questionnaireconfig

import "embed"

// Files 打包问卷 Markdown 模板文件，供后端运行时解析。
//
//go:embed *.md
var Files embed.FS
