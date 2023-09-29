const path = require('path');

module.exports = {
  entry: {
    client: {
      import: './javascript/client.js',
    },
  },
  // https://www.robinwieruch.de/webpack-babel-setup-tutorial/
  // https://babeljs.io/docs/en/babel-plugin-proposal-class-properties
  module: {
    rules: [{
      test: /\.(js)?$/,
      exclude: /node_modules/,
      use: ['babel-loader']
    }]
  },
  resolve: {
    extensions: ['*', '.js']
  },
  output: {
    filename: '[name].bundle.js',
    path: path.resolve(__dirname, '../public/javascript'),
  }
};
