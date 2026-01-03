/** @type {import('tailwindcss').Config} */
export default {
    content: [
        "./index.html",
        "./src/**/*.{js,ts,tsx}",
    ],
    theme: {
        extend: {
            colors: {
                background: '#0a0a0c',
                card: '#141417',
                primary: '#3b82f6',
                success: '#10b981',
                warning: '#f59e0b',
                danger: '#ef4444',
            },
        },
    },
    plugins: [],
}
