module.exports = (gulp, config) => () => {
  const del = require('del');
  const path = require('path');

  return del([config.distDir, config.tsOutDir]);
};
