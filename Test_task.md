# Description of implemented solution

## Main steps:

1. Wrote high-level technical specification for the application
2. Started brainstorming session with Claude to write detailed guidelines for the application (see `/docs` directory)
3. After guidelines were written and reviewed, started implementing the application step by step with help of gemini 3 pro
4. Small adjustments were made to the guidelines that were added extra steps to the implementation process
5. Final review, check of the tests and manual run of the application

## Assumptions/decisions made in case of unclear requirements or missing information:

1. What HTML version has the document? - DetectHTMLVersion checking html tag
2. What is the page title? - ExtractTitle - checking title tag content
3. How many headings of what level are in the document? - CountHeadings - checking h1, h2, h3, h4, h5, h6 tags
4. How many internal and external links are in the document? Are there any inaccessible links and how many? - CountLinks - checking a tags only. CSS, JS, images are not checked.
5. Does the page contain a login form? - HasLoginForm - Check if this form has a password input

## Suggestions on possible improvements of the application:

1. JavaScript-heavy site rendering (SPA support)
2. Authentication/user accounts
3. Analysis history/persistence
4. API endpoints (only web UI)
5. Image/CSS/JS file analysis (separate from links)
6. Performance Metrics
7. SEO analysis
8. Accessibility scoring

