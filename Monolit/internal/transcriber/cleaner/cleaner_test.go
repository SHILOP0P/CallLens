package cleaner

import "testing"

func TestCleanRemovesNoiseMarkersAndArtifactLines(t *testing.T) {
	input := " [музыка]\nЗдравствуйте, это Дмитрий!!!  \nСубтитры сделал Иван\nКлиент: да, мне интересно....\n[аплодисменты]\n"

	got := Clean(input)
	want := "Здравствуйте, это Дмитрий!\nКлиент: да, мне интересно..."

	if got != want {
		t.Fatalf("Clean() = %q, want %q", got, want)
	}
}

func TestCleanKeepsRealSpeech(t *testing.T) {
	input := "Менеджер: эээ, правильно понимаю, что вам нужен тариф?\nКлиент: да, но цена кажется высокой."

	got := Clean(input)
	if got != input {
		t.Fatalf("Clean() changed real speech: %q", got)
	}
}

func TestCleanRemovesEnglishArtifactLine(t *testing.T) {
	input := "Hello, this is a call.\nThanks for watching.\nClient: I need pricing."

	got := Clean(input)
	want := "Hello, this is a call.\nClient: I need pricing."

	if got != want {
		t.Fatalf("Clean() = %q, want %q", got, want)
	}
}
