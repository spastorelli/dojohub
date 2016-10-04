module.exports = (gulp, path, config) => () => {
  const ts = require('gulp-typescript');
  var tsConfig = path.join(config.projectDir, 'tsconfig.json');
  var tsProject = ts.createProject(tsConfig);

  return tsProject.src()
          .pipe(tsProject())
          .js.pipe(gulp.dest(config.tsOutDir));
};
