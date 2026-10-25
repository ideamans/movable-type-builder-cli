package config

import "strings"

// DotEnvFile はカレントディレクトリから読む環境変数ファイルの名前です。
const DotEnvFile = ".env"

// ParseDotEnv は .env の中身を解釈します。
//
// 対応するのは `KEY=value`、行頭の `export `、`#` で始まる行のコメント、
// 値を囲む引用符の3つだけです。**行末コメントは扱いません。**
// `MTB_PASSWORD=pa#ss` の `#` から先を落とすと、パスワードが黙って
// 変わった状態で 401 を踏むことになります。取りこぼすより安全です。
//
// 複数行の値やエスケープ列（\n など）も解きません。ここで受け取るのは
// 接続先の名前・URL・パスワードで、いずれも1行に収まります。
func ParseDotEnv(data []byte) map[string]string {
	vars := map[string]string{}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(strings.TrimSuffix(line, "\r"))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, value, ok := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		if !ok || key == "" {
			continue
		}
		vars[key] = unquote(strings.TrimSpace(value))
	}
	return vars
}

// unquote は値を囲む引用符だけを外します。中身には手を触れません。
func unquote(v string) string {
	if len(v) < 2 {
		return v
	}
	q := v[0]
	if (q == '"' || q == '\'') && v[len(v)-1] == q {
		return v[1 : len(v)-1]
	}
	return v
}

// DotEnv は base の後ろに .env の値を重ねた Environ を返します。
//
// **プロセスの環境変数が勝ちます。** `MTB_PROFILE=stg mtb ...` と打った
// ものを、置いてあるファイルが黙って覆すことがあってはいけません。
// .env は「打っていないときの既定」であって、上書きではありません。
func DotEnv(base Environ, vars map[string]string) Environ {
	if len(vars) == 0 {
		return base
	}
	return func(key string) string {
		if base != nil {
			if v := base(key); v != "" {
				return v
			}
		}
		return vars[key]
	}
}
