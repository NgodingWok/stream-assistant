const path = require('path');
const fs = require('fs');

const pkgPath = path.join(__dirname, '..', '..', 'package.json');
const pkg = JSON.parse(fs.readFileSync(pkgPath, 'utf8'));

const version = pkg.version;

function execute(args) {
  console.log(`v${version}`);
}

module.exports = execute;