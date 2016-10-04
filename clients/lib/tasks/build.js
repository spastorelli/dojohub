module.exports = (gulp, path, config, minify) => () => {
  const browserify = require("browserify");
  const uglify = require('gulp-uglify');
  const globby = require('globby');
  const source = require('vinyl-source-stream');
  const buffer = require('vinyl-buffer');

  var tsOutEntries = path.join(config.tsOutDir, '*.js');
  var jsEntries = path.join(config.srcDir, '*.js');

  return globby([tsOutEntries, jsEntries]).then(entries => {
    var b = browserify({
      debug: true,
      standalone: config.libNs,
      entries: entries
    });
    var bundle = b.bundle();
    var pStream = bundle.pipe(source(config.libName + '.js'));

    if (minify) {
      pStream = pStream.pipe(buffer());
      pStream = pStream.pipe(uglify(config.libName + '.min.js'));
    }
    return pStream.pipe(gulp.dest(config.distDir));
  });
};
