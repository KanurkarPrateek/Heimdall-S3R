# RPC Load Balancer Documentation

This folder contains the complete documentation website for the RPC Load Balancer project.

## 📁 Structure

```
docs/
├── index.html              # Main documentation page
├── assets/
│   ├── css/
│   │   └── style.css      # Styling and design system
│   ├── js/
│   │   └── main.js        # Interactive features
│   └── images/
│       ├── architecture.png
│       ├── failover.png
│       ├── cost.png
│       └── deployment.png
└── README.md              # This file
```

## 🚀 Viewing Locally

1. **Using Python's built-in server:**
   ```bash
   cd docs
   python3 -m http.server 8000
   ```
   Then open `http://localhost:8000` in your browser.

2. **Using npx serve:**
   ```bash
   npx serve docs
   ```

3. **Opening directly:**
   Simply open `index.html` in your browser.

## 🌐 GitHub Pages Deployment

This documentation site is configured for GitHub Pages. To deploy:

1. Push the `docs/` folder to your repository
2. Go to repository Settings → Pages
3. Set Source to "Deploy from a branch"
4. Select branch: `main`, folder: `/docs`
5. Click Save

Your site will be available at: `https://kanurkarprateek.github.io/rpc-load-balancer/`

## 📝 Documentation Sections

The documentation includes:

1. **Introduction** - What the project is, why it exists, key features
2. **Getting Started** - Installation, requirements, first-run example
3. **Quick Start Tutorial** - Step-by-step walkthrough with examples
4. **Core Concepts** - Architecture overview, how components work
5. **Configuration** - All config options with examples
6. **Usage Guide** - Common workflows and best practices
7. **API Reference** - Endpoints, parameters, responses
8. **Deployment** - Local, Docker, and Cloud deployment
9. **Troubleshooting** - Common issues and debugging
10. **Security Notes** - Auth, secrets, permissions
11. **Contributing Guide** - How to contribute, coding style
12. **FAQ** - Frequently asked questions
13. **Roadmap** - Future plans and phases

## 🎨 Design Features

- **Modern Dark Theme** - Premium look with vibrant accent colors
- **Responsive Design** - Works on desktop, tablet, and mobile
- **Smooth Animations** - Scroll effects and interactive elements
- **Code Highlighting** - Syntax-aware with copy buttons
- **Interactive Navigation** - Auto-highlighting active sections
- **Accessibility** - Semantic HTML, keyboard navigation

## 🛠️ Customization

### Update Colors
Edit `assets/css/style.css` and modify CSS variables:
```css
:root {
  --color-primary: #4CAF50;
  --color-secondary: #2196F3;
  /* ... */
}
```

### Add New Sections
1. Add section to `index.html`
2. Add navigation link in navbar
3. Update the navigation JavaScript in `assets/js/main.js`

### Replace Images
Add your images to `assets/images/` and update references in `index.html`.

## 📦 Dependencies

The documentation uses:
- **Google Fonts** (Inter, JetBrains Mono) - loaded via CDN
- **Pure HTML/CSS/JS** - no build step required
- **No frameworks** - vanilla JavaScript for maximum performance

## 📄 License

MIT License - same as the main project
