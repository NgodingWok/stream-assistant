const consola = require('consola');

function execute(args) {
  consola.info("stream-assistant - TikTok live assistant");
  consola.log("Usage:");
  consola.log("  stream-assistant <username> [options]    Connect to a TikTok streamer and log chat/gifts");
  consola.log("  stream-assistant help                  Display this help message");
  consola.log("  stream-assistant version               Show installed version\n");
  consola.log("Options:");
  consola.log("  -n, --no-log          Disable logging of chat messages");
  consola.log("  -g, --no-log-gift     Disable logging of gift events");
  consola.log("  -t, --no-tts          Disable text-to-speech for messages");
  consola.log("  -l, --lang <code>     Set language code for TTS (default 'id')\n");
  consola.log("Examples:");
  consola.log("  stream-assistant some_user");
  consola.log("  stream-assistant some_user -n -t");
  consola.log("  stream-assistant --version");
  consola.log("  stream-assistant --help");
}

module.exports = execute;