/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./internal/http/html/static/templates/**/*.tmpl",
    "./internal/http/html/static/templates/layout.tmpl",
    "./internal/http/html/static/images/*.svg",
    "./internal/http/html/static/js/main.js",
  ],
  theme: {
    extend: {},
  },
  plugins: [],
}
