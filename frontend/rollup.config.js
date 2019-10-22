import resolve from "rollup-plugin-node-resolve";
import commonjs from "rollup-plugin-commonjs";
import postcss from "rollup-plugin-postcss";
import json from "rollup-plugin-json";
import VuePlugin from "rollup-plugin-vue";
import replace from "rollup-plugin-replace";
import {
  terser
} from "rollup-plugin-terser";

import postcss_url from "postcss-url";
import path from "path";
import fs from "fs";
import glob from "glob";

let fontFolder = "../assets/public/static/fonts";
fs.mkdirSync(fontFolder, {
  recursive: true
});

const production = !process.env.NODE_ENV === 'debug';
const plugins = [
  VuePlugin({
    // https://github.com/vuejs/rollup-plugin-vue/issues/238
    needMap: false
  }),
  commonjs(),
  resolve({
    //browser: true,
    preferBuiltins: false
  }),
  postcss({
    minimize: production,
    plugins: [postcss_url({
      // copy ALMOST does what we want - it renames the asset files... however,
      // what we ACTUALLY want is to move the files AND rename them relative to the root
      // so let's do that here
      url: function (asset, dir, options, decl, warn, result, addDependency) {
        if (asset.url.startsWith("data:")) {
          return asset.url;
        }
        // We really only need woff2 files, since we are targeting modern browsers
        if (path.extname(asset.absolutePath) == ".woff2") {
          let toURL = fontFolder + "/" + path.basename(asset.absolutePath);
          fs.copyFile(asset.absolutePath, toURL, (err) => {
            if (err) throw err;
            console.log(asset.relativePath, " -> ", toURL);
          });
        }


        return "/static/fonts/" + path.basename(asset.url);

      },
    })]
  }),
  json({
    compact: production
  }),
  replace({
    "process.env.NODE_ENV": JSON.stringify(production ? "production" : "debug")
  })
];
if (production) {
  plugins.push(terser({
    compress: {
      drop_console: true,
      ecma: 10 // Heedy doesn't do backwards compatibility
    },
    mangle: true,
    module: true
  }));
}

function checkExternal(modid, parent, isResolved) {
  return (!isResolved && modid.endsWith(".mjs")) || modid.startsWith("http");
}

function out(name, loc = "", format = "es") {
  let filename = name.slice(0, name.lastIndexOf("."))
  return {
    input: "src/" + name,
    output: {
      name: filename,
      file: "../assets/public/static/" +
        loc +
        filename +
        (format == "es" ? ".mjs" : ".js"),
      format: format
    },
    plugins: plugins,
    external: checkExternal
  };
}
export default [
  // The base files
  out("app.js"),
  out("worker.js"),
  out("auth.js"),
  out("setup.js"),
  out("api.js"),
  // The main app's files
  out("heedy/main.js")
].concat(glob.sync("dist/*.js", {
  cwd: "./src"
}).map(a => out(a)));
/*
.concat(
  glob.sync("heedy/components/*.vue", { cwd: "./src" }).map(a => out(a))
);*/