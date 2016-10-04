module.exports = (gulp, path, config) => () => {
  const browserify = require("browserify");
  const globby = require('globby');
  const source = require('vinyl-source-stream');

  var tsOutEntries = path.join(config.tsOutDir, '*.js');
  var jsEntries = path.join(config.srcDir, '*.js');

  var entries = globby.sync([tsOutEntries, jsEntries]);
  return browserify({
      standalone: config.libNs,
      entries: entries
    }).bundle()
      .pipe(source(config.libName + '.js'))
      .pipe(gulp.dest(config.distDir));
};
