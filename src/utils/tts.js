const path = require('path');
const fs = require('fs');
const gTTS = require('gtts');
const sound = require('sound-play');
const consola = require('consola');

var isPlaying = false;
const filePath = path.join(__dirname, 'temp.mp3');

async function textToSpeech(text, lang = 'id') {
  if (isPlaying) return;
  if (!text) return;
  isPlaying = true;

  const gtts = new gTTS(text, lang);

  gtts.save(filePath, async function (err) {
    if (err) return consola.error("Failed to save:", err);
  });

  try {
    await sound.play(filePath);
    fs.unlinkSync(filePath);
    isPlaying = false;
  } catch (error) {
    consola.error("Failed to play:", error);
  }
}

module.exports = textToSpeech;