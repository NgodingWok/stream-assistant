#!/usr/bin/env node
const consola = require("consola");

// Get input flag
const args = process.argv.slice(2);

// Deprecated warning
consola.warn("This project is deprecated and no longer maintained. Visit https://github.com/NgodingWok/stream-assistant for more information.");

(async () => {
  // Handle flag
  switch (args[0]) {
    case "--help":
    case "-h":
		case "help": {
      require("./handler/help")(args);
      break;
    }
    case "--version":
    case "-v": {
      require("./handler/version")(args);
      break;
    }
    default: {
      await require("./handler/default")(args);
      break;
    }
  }
})();
