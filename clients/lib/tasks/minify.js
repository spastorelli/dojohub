module.exports = (gulp, path, config) => () => {
  const fs = require('fs');
  const uglify = require('uglify-js');

  const bundleJsFileName = config.libName + '.js';
  const bundleJsFile = path.join(config.distDir, bundleJsFileName);
  const minfiedJsFile = path.join(config.distDir, config.libName + '.min.js');

  var minfiedJs = uglify.minify(bundleJsFile, {
    mangle: true,
    compress: {
      sequences: true,
      properties: true,
  		dead_code: true,
  		conditionals: true,
  		booleans: true,
  		unused: true,
  		if_return: true,
  		join_vars: true,
  		drop_console: true
  	}
  });
  return fs.writeFileSync(minfiedJsFile, minfiedJs.code);
};
