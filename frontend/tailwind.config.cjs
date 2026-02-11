/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./src/**/*.{html,js,svelte,ts}"],
  theme: {
    extend: {
      colors: {
        brand: {
          50: "#ecfdf3",
          100: "#d1fae5",
          500: "#10b981",
          700: "#047857",
          900: "#064e3b"
        }
      }
    }
  },
  plugins: []
};
