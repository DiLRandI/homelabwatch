/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{js,jsx}"],
  theme: {
    extend: {
      colors: {
        base: "#09121a",
        panel: "#10202b",
        ink: "#ecf3ed",
        accent: "#f2c45a",
        danger: "#ff7a6d",
        warn: "#ffb347",
        ok: "#52d49b",
        muted: "#9bb2b5",
      },
      fontFamily: {
        display: ['"Avenir Next"', '"Segoe UI"', "sans-serif"],
        mono: ['"IBM Plex Mono"', '"SFMono-Regular"', "monospace"],
      },
      boxShadow: {
        halo: "0 18px 60px rgba(7, 13, 19, 0.35)",
      },
      animation: {
        floatIn: "floatIn 420ms ease-out",
      },
      keyframes: {
        floatIn: {
          "0%": { opacity: "0", transform: "translateY(10px)" },
          "100%": { opacity: "1", transform: "translateY(0)" },
        },
      },
    },
  },
  plugins: [],
};
