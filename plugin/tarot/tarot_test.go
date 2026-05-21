package tarot

import (
	"strings"
	"testing"
)

func TestSplitSingleCardQuestion(t *testing.T) {
	tests := []struct {
		name       string
		drawCount  string
		question   string
		want       string
		wantUsable bool
	}{
		{
			name:       "single card question",
			question:   " 最近工作会怎样 ",
			want:       "最近工作会怎样",
			wantUsable: true,
		},
		{
			name:     "single card without question",
			question: " ",
		},
		{
			name:      "multi card question ignored",
			drawCount: "3张",
			question:  "最近工作会怎样",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := splitSingleCardQuestion(tt.drawCount, tt.question)
			if ok != tt.wantUsable {
				t.Fatalf("splitSingleCardQuestion() usable = %v, want %v", ok, tt.wantUsable)
			}
			if got != tt.want {
				t.Fatalf("splitSingleCardQuestion() question = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSplitFormationQuestion(t *testing.T) {
	formations := map[string]formation{
		"时间之流": {},
		"圣三角":  {},
	}

	tests := []struct {
		name          string
		raw           string
		wantFormation string
		wantQuestion  string
		wantOK        bool
	}{
		{
			name:          "formation only",
			raw:           "时间之流",
			wantFormation: "时间之流",
			wantOK:        true,
		},
		{
			name:          "formation with spaced question",
			raw:           "时间之流 最近关系会怎样",
			wantFormation: "时间之流",
			wantQuestion:  "最近关系会怎样",
			wantOK:        true,
		},
		{
			name:          "formation with compact question",
			raw:           "圣三角我该换工作吗",
			wantFormation: "圣三角",
			wantQuestion:  "我该换工作吗",
			wantOK:        true,
		},
		{
			name: "unknown formation",
			raw:  "不存在 事情",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFormation, gotQuestion, ok := splitFormationQuestion(tt.raw, formations)
			if ok != tt.wantOK {
				t.Fatalf("splitFormationQuestion() ok = %v, want %v", ok, tt.wantOK)
			}
			if gotFormation != tt.wantFormation {
				t.Fatalf("splitFormationQuestion() formation = %q, want %q", gotFormation, tt.wantFormation)
			}
			if gotQuestion != tt.wantQuestion {
				t.Fatalf("splitFormationQuestion() question = %q, want %q", gotQuestion, tt.wantQuestion)
			}
		})
	}
}

func TestBuildTarotPrompt(t *testing.T) {
	t.Run("single card", func(t *testing.T) {
		prompt := buildTarotPrompt("我该换工作吗", "", []drawResult{
			{
				Name:        "愚者",
				Position:    "正位",
				Description: "新的开始",
			},
		})

		for _, want := range []string{"我该换工作吗", "愚者", "正位", "新的开始"} {
			if !strings.Contains(prompt, want) {
				t.Fatalf("buildTarotPrompt() missing %q in %q", want, prompt)
			}
		}
	})

	t.Run("formation", func(t *testing.T) {
		prompt := buildTarotPrompt("最近关系会怎样", "时间之流", []drawResult{
			{
				Name:        "恋人",
				Position:    "逆位",
				Description: "关系失衡",
				Represent:   "过去",
			},
		})

		for _, want := range []string{"最近关系会怎样", "时间之流", "过去", "恋人", "逆位", "关系失衡"} {
			if !strings.Contains(prompt, want) {
				t.Fatalf("buildTarotPrompt() missing %q in %q", want, prompt)
			}
		}
	})
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

func TestBuildTarotAnalysisChunks(t *testing.T) {
	got := buildTarotAnalysisChunks("结果", 1000)
	if len(got) != 1 {
		t.Fatalf("buildTarotAnalysisChunks() chunks = %d, want 1", len(got))
	}
	if got[0] != "塔罗解析:\n结果" {
		t.Fatalf("buildTarotAnalysisChunks() first chunk = %q, want %q", got[0], "塔罗解析:\n结果")
	}
}
