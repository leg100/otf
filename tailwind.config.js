/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./internal/http/html/static/templates/**/*.tmpl",
    "./internal/http/html/static/templates/layout.tmpl",
    "./internal/http/html/static/images/*.svg",
  ],
  theme: {
    extend: {},
  },
  plugins: [],
}
