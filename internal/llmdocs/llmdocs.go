// Package llmdocs は `mtb llm` が出力するリファレンスを埋め込みます。
//
// 00- と 10- は手書き、90-commands.md は cobra のコマンドツリーから
// `go generate ./...` で生成してコミットします（go:embed がビルド時に実体を
// 要求するため）。CI が再生成して差分が出たら失敗します。
// go:generate の指定は cmd/mtb/main.go にあります。
//
// 拡張（internal/extension）の章は実行時に足します。名前は 80- で始まり、
// 手書きの章と生成されたコマンドカタログの間に並びます。
package llmdocs

import (
	"embed"
	"io/fs"
	"sort"
	"time"

	kit "github.com/ideamans/go-llm-cli-kit/llmdocs"
)

//go:embed *.md
var files embed.FS

// Docs は埋め込みリファレンスです。
func Docs() *kit.Docs { return kit.New(files, ".") }

// DocsWith は埋め込みリファレンスに章を足したものです。extra のキーは
// ファイル名（並び順を決める）、値は Markdown です。
func DocsWith(extra map[string]string) *kit.Docs {
	if len(extra) == 0 {
		return Docs()
	}
	m := memFS{}
	entries, _ := fs.ReadDir(files, ".")
	for _, e := range entries {
		data, err := fs.ReadFile(files, e.Name())
		if err == nil {
			m[e.Name()] = data
		}
	}
	for name, body := range extra {
		m[name] = []byte(body)
	}
	return kit.New(m, ".")
}

// memFS は1階層だけのメモリ上のファイルシステムです。kit が使うのは
// ReadDir と ReadFile だけなので、その2つを直接持たせています。
type memFS map[string][]byte

func (m memFS) Open(name string) (fs.File, error) {
	return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
}

func (m memFS) ReadFile(name string) ([]byte, error) {
	data, ok := m[name]
	if !ok {
		return nil, &fs.PathError{Op: "read", Path: name, Err: fs.ErrNotExist}
	}
	return data, nil
}

func (m memFS) ReadDir(name string) ([]fs.DirEntry, error) {
	if name != "." {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrNotExist}
	}
	out := make([]fs.DirEntry, 0, len(m))
	for n, data := range m {
		out = append(out, fs.FileInfoToDirEntry(memInfo{name: n, size: int64(len(data))}))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name() < out[j].Name() })
	return out, nil
}

type memInfo struct {
	name string
	size int64
}

func (i memInfo) Name() string       { return i.name }
func (i memInfo) Size() int64        { return i.size }
func (i memInfo) Mode() fs.FileMode  { return 0o444 }
func (i memInfo) ModTime() time.Time { return time.Time{} }
func (i memInfo) IsDir() bool        { return false }
func (i memInfo) Sys() any           { return nil }
