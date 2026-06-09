package bugua

import (
	"strings"
	"testing"
)

func TestLookupHexagramByLines(t *testing.T) {
	tests := []struct {
		name string
		yao  yaoValues
		want hexagram
	}{
		{
			name: "all yang is qian",
			yao:  yaoValues{youngYang, youngYang, youngYang, youngYang, youngYang, youngYang},
			want: hexagram{Number: 1, Name: "乾"},
		},
		{
			name: "all yin is kun",
			yao:  yaoValues{youngYin, youngYin, youngYin, youngYin, youngYin, youngYin},
			want: hexagram{Number: 2, Name: "坤"},
		},
		{
			name: "water over fire is jiji",
			yao:  yaoValues{youngYang, youngYin, youngYang, youngYin, youngYang, youngYin},
			want: hexagram{Number: 63, Name: "既济"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lookupHexagram(tt.yao)
			if got.Number != tt.want.Number || got.Name != tt.want.Name {
				t.Fatalf("lookupHexagram() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestChangedYaoValues(t *testing.T) {
	got := yaoValues{oldYang, youngYang, oldYin, youngYin, oldYang, oldYin}.changed()
	want := yaoValues{youngYin, youngYang, youngYang, youngYin, youngYin, youngYang}
	if got != want {
		t.Fatalf("changed() = %#v, want %#v", got, want)
	}
}

func TestMovingLineNames(t *testing.T) {
	got := yaoValues{oldYang, youngYin, youngYang, oldYin, youngYin, oldYang}.movingLines()
	want := []string{"初九", "六四", "上九"}
	if len(got) != len(want) {
		t.Fatalf("movingLines() = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("movingLines()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestBuildDivinationPrompt(t *testing.T) {
	result := divinationResult{
		Question: "事业该不该换方向",
		Original: hexagram{Number: 1, Name: "乾"},
		Changed:  hexagram{Number: 2, Name: "坤"},
		Yao:      yaoValues{oldYang, oldYang, oldYang, oldYang, oldYang, oldYang},
	}

	prompt := result.prompt()
	for _, want := range []string{
		"事业该不该换方向",
		"本卦: 第1卦 乾",
		"变卦: 第2卦 坤",
		"动爻: 初九、九二、九三、九四、九五、上九",
		"不要把占卜结果表述为确定事实",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt() missing %q in:\n%s", want, prompt)
		}
	}
}

func TestSplitTextChunks(t *testing.T) {
	got := splitTextChunks("甲乙丙丁", 3)
	if len(got) != 2 {
		t.Fatalf("splitTextChunks() chunks = %d, want 2", len(got))
	}
	if got[0] != "甲乙丙" || got[1] != "丁" {
		t.Fatalf("splitTextChunks() = %#v, want []string{\"甲乙丙\", \"丁\"}", got)
	}
}
