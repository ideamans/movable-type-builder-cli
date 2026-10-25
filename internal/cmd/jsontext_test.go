package cmd

import "testing"

// Movable Type は日本語を \uXXXX でエスケープした JSON を返します。
// そのまま出すと人が読めず、エージェントにとっては同じ内容が数倍の
// トークンになるので、既定で戻します。
func TestUnescapeUnicode(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"日本語", `{"name":"\u30a2\u30a4\u30c7\u30a2\u30de\u30f3\u30ba"}`, `{"name":"アイデアマンズ"}`},
		{"エスケープ無しは素通し", `{"id":1,"ok":true}`, `{"id":1,"ok":true}`},
		{"記号", `{"tz":"\u002b09:00"}`, `{"tz":"+09:00"}`},
		{"サロゲートペア(絵文字)", `{"e":"\ud83d\ude80"}`, `{"e":"🚀"}`},
		{"制御文字は残す", `{"s":"a\u0009b"}`, `{"s":"a\u0009b"}`},
		{"引用符は残す", `{"s":"\u0022"}`, `{"s":"\u0022"}`},
		{"バックスラッシュは残す", `{"s":"\u005c"}`, `{"s":"\u005c"}`},
		{"リテラルのバックスラッシュ", `{"s":"\\u3042"}`, `{"s":"\\u3042"}`},
		{"他のエスケープは触らない", `{"s":"a\nb\"c"}`, `{"s":"a\nb\"c"}`},
		{"片割れサロゲートは残す", `{"s":"\ud83d"}`, `{"s":"\ud83d"}`},
		{"混在", `{"a":"\u3042\u3044\u3046","b":2}`, `{"a":"あいう","b":2}`},
	}
	for _, tt := range tests {
		if got := string(unescapeUnicode([]byte(tt.in))); got != tt.want {
			t.Errorf("%s: unescapeUnicode(%s) = %s, want %s", tt.name, tt.in, got, tt.want)
		}
	}
}

// 復号した結果が JSON として壊れていないこと。
func TestUnescapeUnicodeStaysValidJSON(t *testing.T) {
	in := `{"items":[{"name":"\u30d6\u30ed\u30b0","note":"a\"b\\c\u0009d","emoji":"\ud83d\ude80"}]}`
	got := unescapeUnicode([]byte(in))
	if !jsonValid(got) {
		t.Fatalf("復号後が不正な JSON です: %s", got)
	}
	if string(got) == in {
		t.Error("何も復号されていません")
	}
}
