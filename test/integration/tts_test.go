//go:build integration

package integration

import (
	"os"
	"testing"

	"NgodingWok/stream-assistant/internal/tts"
)

func TestTTS_Speak(t *testing.T) {
	folder := t.TempDir()
	speaker := tts.NewSpeaker("id", folder)

	if err := speaker.Speak("halo dunia, ini tes suara"); err != nil {
		t.Fatalf("Speak failed: %v", err)
	}

	entries, err := os.ReadDir(folder)
	if err != nil {
		t.Fatalf("reading temp dir: %v", err)
	}
	if len(entries) == 0 {
		t.Error("expected TTS to generate audio file in temp folder")
	}
}

func TestTTS_SpeakEnglish(t *testing.T) {
	folder := t.TempDir()
	speaker := tts.NewSpeaker("en", folder)

	if err := speaker.Speak("hello world, this is a test"); err != nil {
		t.Fatalf("Speak failed: %v", err)
	}
}

func TestTTS_EmptyText(t *testing.T) {
	folder := t.TempDir()
	speaker := tts.NewSpeaker("id", folder)

	err := speaker.Speak("")
	if err == nil {
		t.Log("empty text did not error (library may accept it)")
	}
}
