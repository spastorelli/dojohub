const gulp = require('gulp');
const path = require('path');

var config = {
  projectDir: __dirname,
  taskDir: path.join(__dirname, 'tasks'),
  srcDir: path.join(__dirname, 'src'),
  distDir: path.join(__dirname, 'dist'),
  tsOutDir: path.join(__dirname, 'src', 'out'),
  tsLintConf: path.join(__dirname, 'tslint.json'),
  libName: 'dojohub',
  libNs: 'coderdojo'
};

const cleanTask = require(path.join(config.taskDir, 'clean'));
const tsLintTask = require(path.join(config.taskDir, 'tslint'));
const compileTask = require(path.join(config.taskDir, 'compile'));
const buildTask = require(path.join(config.taskDir, 'build'));

gulp.task('clean', cleanTask(gulp, config));
gulp.task('compile', ['clean'], compileTask(gulp, path, config));
gulp.task('tslint', tsLintTask(gulp, config));
gulp.task('build', ['compile'], buildTask(gulp, path, config));
gulp.task('build:min', ['compile'], buildTask(gulp, path, config, /** minify */ true));
