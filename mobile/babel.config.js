module.exports = function (api) {
  // cache keyed on the build mode so the production branch is re-evaluated per env
  api.cache.using(() => process.env.BABEL_ENV || process.env.NODE_ENV);
  const production = process.env.BABEL_ENV === 'production' || process.env.NODE_ENV === 'production';
  return {
    presets: ['babel-preset-expo'],
    // release builds drop every console call so nothing is logged to a connected debugger
    plugins: production ? [['transform-remove-console', { exclude: [] }]] : [],
  };
};
