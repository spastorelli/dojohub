module.exports = (gulp, config) => () => {
  const path = require('path');
  const tslint = require('gulp-tslint');
  const srcDir = config.srcDir;

  return gulp.src([
    path.join(srcDir, '**', '*.ts'),
    path.join('!', srcDir, 'references.ts')
  ]).pipe(tslint({
    configuration: require(config.tsLintConf)
  })).pipe(tslint.report('verbose'));
};
