// Get input flag
const args = process.argv.slice(2);

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
