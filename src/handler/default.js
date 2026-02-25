const consola = require("consola");
const { TikTokLiveConnection, WebcastEvent } = require('tiktok-live-connector');
const textToSpeech = require('../utils/tts');

async function execute(args) {
  // Extract args
  const username = args[0];
  const noLogMessage = args.includes('--no-log') || args.includes('-n'); // Loging for chat messages
  const noLogGift = args.includes('--no-log-gift') || args.includes('-g'); // Loging for gift messages
  const noTTS = args.includes('--no-tts') || args.includes('-t'); // Text to speech for chat messages
  
  // Getting lang
  let lang = 'id';
  const langIndex = args.findIndex(arg => arg === '--lang' || arg === '-l');
  if (langIndex !== -1 && args[langIndex + 1]) {
    lang = args[langIndex + 1];
  }

  if (!username) {
    throw new Error("Username is required. Usage: stream-assistant <username> | see --help for more information.");
  }

  console.log("Default command executed.");
  console.log("Username:", username);

  consola.info("Creating new connection...");
  const connection = new TikTokLiveConnection(username);

  consola.info(`Connecting to ${username}'s live stream...`);
  connection.connect().then(state => {
    consola.success(`Connected to roomId ${state.roomId}`);
  }).catch(err => {
    consola.error('Failed to connect', err);
  });

  connection.on(WebcastEvent.CHAT, data => {
    if (!connection.isConnected) return;

    if (!noLogMessage) {
      consola.info(`[CHAT] ${data.user.uniqueId}: ${data.comment}`);
    }
    if (!noTTS) {
      textToSpeech(data.comment, lang);
    }
  })

  connection.on(WebcastEvent.GIFT, data => {
    if (!connection.isConnected) return;
    
    if (!noLogGift) {
      consola.info(`[GIFT] Received ${data.giftDetails.giftName} x${data.comboCount} from ${data.user.uniqueId}`);
    }
    if (!noTTS) {
      textToSpeech(`Received ${data.giftDetails.giftName} x${data.giftDetails.combo} from ${data.user.uniqueId}`, 'en');
    }
  });
}

module.exports = execute;