package cleaner

import (
	"regexp"
	"strings"
	"unicode"
)

var ArtifactPhrases = []string{
	"продолжение следует",
	"субтитры сделал",
	"субтитры сделаны",
	"субтитры создавал",
	"субтитры подготовлены",
	"редактор субтитров",
	"корректор субтитров",
	"спасибо за просмотр",
	"смотрите продолжение",
	"подписывайтесь на канал",
	"ставьте лайки",
	"thanks for watching",
	"subtitles by",
	"captioned by",
	"transcript by",
}

var (
	spacePattern         = regexp.MustCompile(`[ \t\f\v]+`)
	blankLinePattern     = regexp.MustCompile(`\n{3,}`)
	repeatedBangPattern  = regexp.MustCompile(`!{2,}`)
	repeatedQuestPattern = regexp.MustCompile(`\?{2,}`)
	repeatedDotPattern   = regexp.MustCompile(`\.{4,}`)
	bracketNoisePattern  = regexp.MustCompile(`(?i)(\[[^\]]*(музыка|аплодисменты|смех|шум|тишина|неразборчиво|music|applause|laughter|noise|silence|inaudible)[^\]]*\]|\([^)]*(музыка|аплодисменты|смех|шум|тишина|неразборчиво|music|applause|laughter|noise|silence|inaudible)[^)]*\))`)
)

func Clean(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = strings.ReplaceAll(text, "♪", " ")
	text = strings.ReplaceAll(text, "♫", " ")
	text = bracketNoisePattern.ReplaceAllString(text, " ")

	lines := strings.Split(text, "\n")
	cleanedLines := make([]string, 0, len(lines))

	for _, line := range lines {
		line = cleanLine(line)
		if line == "" || isArtifactLine(line) {
			continue
		}
		cleanedLines = append(cleanedLines, line)
	}

	text = strings.Join(cleanedLines, "\n")
	text = blankLinePattern.ReplaceAllString(text, "\n\n")

	return strings.TrimSpace(text)
}

func cleanLine(line string) string {
	line = spacePattern.ReplaceAllString(line, " ")
	line = repeatedBangPattern.ReplaceAllString(line, "!")
	line = repeatedQuestPattern.ReplaceAllString(line, "?")
	line = repeatedDotPattern.ReplaceAllString(line, "...")
	line = strings.TrimSpace(line)

	return line
}

func isArtifactLine(line string) bool {
	normalizedLine := normalizeForCompare(line)
	if normalizedLine == "" {
		return true
	}

	for _, phrase := range ArtifactPhrases {
		normalizedPhrase := normalizeForCompare(phrase)
		if normalizedLine == normalizedPhrase || strings.HasPrefix(normalizedLine, normalizedPhrase+" ") {
			return true
		}
	}

	return false
}

func normalizeForCompare(value string) string {
	value = strings.ToLower(value)

	var builder strings.Builder
	builder.Grow(len(value))

	previousSpace := false
	for _, r := range value {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			builder.WriteRune(r)
			previousSpace = false
		case unicode.IsSpace(r):
			if !previousSpace {
				builder.WriteRune(' ')
				previousSpace = true
			}
		default:
			if !previousSpace {
				builder.WriteRune(' ')
				previousSpace = true
			}
		}
	}

	return strings.TrimSpace(builder.String())
}
