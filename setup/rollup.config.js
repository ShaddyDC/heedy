import resolve from "rollup-plugin-node-resolve";
import commonjs from "rollup-plugin-commonjs";
import VuePlugin from "rollup-plugin-vue";
import replace from "rollup-plugin-replace";
import { terser } from "rollup-plugin-terser";

const production = !process.env.ROLLUP_WATCH;
const plugins = [
  resolve(),
  commonjs(),
  VuePlugin(),
  replace({
    "process.env.NODE_ENV": JSON.stringify(production ? "production" : "debug")
  })
];
if (production) {
  plugins.push(terser());
}
function out(name, format = "es") {
  let filename = name.split(".");
  return {
    input: "src/" + name,
    output: {
      name: filename[0],
      file:
        "../assets/setup/js/" + filename[0] + (format == "es" ? ".jsm" : ".js"),
      format: format,
      globals: {
        vue: "Vue",
        vuetify: "Vuetify"
      }
    },
    plugins: plugins,
    external: ["vue", "vuetify"]
  };
}

export default [
  // The base file
  out("setup.js", "iife")
];
