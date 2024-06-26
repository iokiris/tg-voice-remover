package main

import (
	"strings"
)

func QMString(s string) string {
	// Экранировать только символы, которые могут повлиять на MarkdownV2
	escaped := strings.NewReplacer(
		"_", "\\_",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "\\`",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"!", "\\!",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"?", "\\?",
	).Replace(s)
	return escaped
}
