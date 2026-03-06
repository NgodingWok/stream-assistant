package tts

import (
	htgotts "github.com/hegedustibor/htgo-tts"
	"github.com/hegedustibor/htgo-tts/handlers"
)

/** Speaker wraps the htgo-tts library for text-to-speech output. */
type Speaker struct {
	engine htgotts.Speech
}

/** NewSpeaker creates a Speaker configured with the given language and temp folder. */
func NewSpeaker(language, folder string) *Speaker {
	return &Speaker{
		engine: htgotts.Speech{
			Folder:   folder,
			Language: language,
			Handler:  &handlers.Native{},
		},
	}
}

/** Speak converts the given text to speech and plays it. */
func (s *Speaker) Speak(text string) error {
	return s.engine.Speak(text)
}
