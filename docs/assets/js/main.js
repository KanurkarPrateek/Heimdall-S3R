// ===================================
// Heimdall S3R Documentation
// Interactive Features
// ===================================

// Initialize on DOM load
document.addEventListener('DOMContentLoaded', () => {
    initNavigation();
    initSmoothScroll();
    initCopyButtons();
    initScrollAnimations();
});

// ===================================
// Sidebar Navigation Tracking
// ===================================
function initNavigation() {
    window.addEventListener('scroll', () => {
        updateActiveNavLink();
    });

    // Set initial active link
    updateActiveNavLink();
}

function updateActiveNavLink() {
    const sections = document.querySelectorAll('.section');
    const navLinks = document.querySelectorAll('.nav-link');

    let currentSection = '';

    sections.forEach(section => {
        const sectionTop = section.offsetTop;
        const sectionHeight = section.clientHeight;

        // Tolerance for active section
        if (window.scrollY >= sectionTop - 100) {
            currentSection = section.getAttribute('id');
        }
    });

    navLinks.forEach(link => {
        link.classList.remove('active');
        if (link.getAttribute('href') === `#${currentSection}`) {
            link.classList.add('active');
        }
    });
}

// ===================================
// Smooth Scrolling
// ===================================
function initSmoothScroll() {
    const links = document.querySelectorAll('a[href^="#"]');

    links.forEach(link => {
        link.addEventListener('click', (e) => {
            const targetId = link.getAttribute('href');

            // Skip if it's just "#"
            if (targetId === '#') return;

            e.preventDefault();

            const targetElement = document.querySelector(targetId);
            if (targetElement) {
                const targetPosition = targetElement.offsetTop - 20;

                window.scrollTo({
                    top: targetPosition,
                    behavior: 'smooth'
                });
            }
        });
    });
}

// ===================================
// Copy Code Buttons
// ===================================
function initCopyButtons() {
    // Look for pre tags directly or within code-block
    const codeBlocks = document.querySelectorAll('pre');

    codeBlocks.forEach(block => {
        // Create copy button
        const copyBtn = document.createElement('button');
        copyBtn.innerText = 'Copy';
        copyBtn.className = 'copy-btn';
        copyBtn.style.position = 'absolute';
        copyBtn.style.right = '10px';
        copyBtn.style.top = '10px';
        copyBtn.style.padding = '4px 8px';
        copyBtn.style.fontSize = '12px';
        copyBtn.style.background = '#27272a';
        copyBtn.style.color = '#fff';
        copyBtn.style.border = '1px solid #3f3f46';
        copyBtn.style.borderRadius = '4px';
        copyBtn.style.cursor = 'pointer';

        block.style.position = 'relative';
        block.appendChild(copyBtn);

        copyBtn.addEventListener('click', async () => {
            const code = block.querySelector('code').textContent;
            try {
                await navigator.clipboard.writeText(code);
                copyBtn.innerText = 'Copied!';
                setTimeout(() => copyBtn.innerText = 'Copy', 2000);
            } catch (err) {
                console.error('Failed to copy', err);
            }
        });
    });
}

// ===================================
// Scroll Animations
// ===================================
function initScrollAnimations() {
    const observerOptions = {
        threshold: 0.1
    };

    const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                entry.target.style.opacity = '1';
                entry.target.style.transform = 'translateY(0)';
            }
        });
    }, observerOptions);

    const animatedElements = document.querySelectorAll('.card, .feature');

    animatedElements.forEach((el, index) => {
        el.style.opacity = '0';
        el.style.transform = 'translateY(20px)';
        el.style.transition = `opacity 0.4s ease ${index * 0.05}s, transform 0.4s ease ${index * 0.05}s`;
        observer.observe(el);
    });
}
