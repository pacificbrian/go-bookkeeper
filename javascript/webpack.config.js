const path = require('path');

module.exports = {
  entry: './javascript/client.js',
  output: {
    path: path.resolve(__dirname, '../public/javascript'),
    filename: 'bundle.js',
  },
};
