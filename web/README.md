# Web Dashboard for Bybit Trading Bot

This directory contains the static web files for the trading bot dashboard, designed for deployment on Vercel.

## Structure

- `static/` - Contains all static assets (HTML, CSS, JavaScript)
- `static/index.html` - Main dashboard page
- `static/styles.css` - Styling for the dashboard
- `static/script.js` - Client-side JavaScript for dashboard functionality

## Deployment to Vercel

1. Push this repository to GitHub
2. Connect your GitHub repository to Vercel
3. Vercel will automatically detect the `vercel.json` configuration file
4. The dashboard will be deployed and accessible online

## Local Development

To test locally:

1. Start the trading bot backend (runs on port 8080 by default)
2. Open `static/index.html` in a browser
3. The frontend will automatically connect to the backend API

Note: For local development, you may need to configure CORS settings in the backend if serving the frontend from a different port.