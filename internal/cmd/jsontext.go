package cmd

import (
	"encoding/json"
	"unicode/utf16"
	"unicode/utf8"
)

// unescapeUnicode は JSON 中の \uXXXX を実際の文字に戻します。
//
// Movable Type は日本語を ア のようにエスケープした JSON を返します。
// そのまま出すと人が読めず、AIエージェントにとっても同じ内容が3倍の
// トークンになります。JSON としての意味は変わらないので、既定で戻します。
// 元のバイト列がほしいときは --format raw を使ってください。
//
// 文字列の外にバックスラッシュは現れないので、エスケープの並びだけを見れば
// 安全に置換できます。JSON で裸にできない文字（" \ 制御文字）はエスケープの
// まま残します。
func unescapeUnicode(src []byte) []byte {
	// \u が無ければ何もしない（大半の応答はこの分岐で抜ける）。
	if !hasUnicodeEscape(src) {
		return src
	}

	out := make([]byte, 0, len(src))
	for i := 0; i < len(src); {
		c := src[i]
		if c != '\\' {
			out = append(out, c)
			i++
			continue
		}
		if i+1 >= len(src) {
			out = append(out, c)
			i++
			continue
		}
		// \uXXXX 以外のエスケープ（\\ \" \n など）は2バイトまとめて写す。
		// ここで \\ を消費するので、\\u3042 を誤って復号することはない。
		if src[i+1] != 'u' {
			out = append(out, src[i], src[i+1])
			i += 2
			continue
		}
		r, size, ok := decodeEscape(src[i:])
		if !ok || mustStayEscaped(r) {
			out = append(out, src[i], src[i+1])
			i += 2
			continue
		}
		out = utf8.AppendRune(out, r)
		i += size
	}
	return out
}

func hasUnicodeEscape(src []byte) bool {
	for i := 0; i+1 < len(src); i++ {
		if src[i] == '\\' {
			if src[i+1] == 'u' {
				return true
			}
			i++ // \\ を飛ばす
		}
	}
	return false
}

// decodeEscape は src の先頭にある \uXXXX（必要ならサロゲートペア）を読みます。
// 戻り値は文字と消費したバイト数です。
func decodeEscape(src []byte) (rune, int, bool) {
	v, ok := hex4(src)
	if !ok {
		return 0, 0, false
	}
	r := rune(v)
	if !utf16.IsSurrogate(r) {
		return r, 6, true
	}
	// 上位サロゲートなら、続く \uXXXX と合わせて1文字にする。
	if len(src) >= 12 && src[6] == '\\' && src[7] == 'u' {
		if v2, ok := hex4(src[6:]); ok {
			if combined := utf16.DecodeRune(r, rune(v2)); combined != utf8.RuneError {
				return combined, 12, true
			}
		}
	}
	// 片割れだけのサロゲートは復号できない。エスケープのまま残す。
	return 0, 0, false
}

func hex4(src []byte) (int, bool) {
	if len(src) < 6 || src[0] != '\\' || src[1] != 'u' {
		return 0, false
	}
	v := 0
	for _, c := range src[2:6] {
		switch {
		case c >= '0' && c <= '9':
			v = v<<4 | int(c-'0')
		case c >= 'a' && c <= 'f':
			v = v<<4 | int(c-'a'+10)
		case c >= 'A' && c <= 'F':
			v = v<<4 | int(c-'A'+10)
		default:
			return 0, false
		}
	}
	return v, true
}

// mustStayEscaped は JSON の文字列中に裸で置けない文字かを返します。
func mustStayEscaped(r rune) bool {
	return r < 0x20 || r == '"' || r == '\\'
}

// jsonValid はテストから使う検証用の薄いラッパーです。
func jsonValid(b []byte) bool { return json.Valid(b) }
