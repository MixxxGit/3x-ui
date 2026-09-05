import { createMDX } from 'fumadocs-mdx/next';

const withMDX = createMDX();

// Set DEPLOY_TARGET=static to produce a fully static export (e.g. for GitHub
// Pages). Search already uses a static index, and OG images are prerendered, so
// the export is self-contained. Default (unset) builds for Vercel/Node hosting.
const isStaticExport = process.env.DEPLOY_TARGET === 'static';

// GitHub Pages project sites are served from a subpath
// (https://<user>.github.io/<repo>/), so every asset and route URL needs that
// prefix — otherwise the browser requests /_next/... from the domain root and
// gets a 404 (no CSS, oversized SVG icons, dead internal links). Custom-domain
// Pages sites are served at the domain root and must stay unprefixed, so this
// is an explicit opt-in env var rather than being derived from isStaticExport.
// NEXT_PUBLIC_ because client components (Logo) need the same prefix.
const basePath = process.env.NEXT_PUBLIC_BASE_PATH || '';

/** @type {import('next').NextConfig} */
const config = {
  reactStrictMode: true,
  // On the static host (GitHub Pages) emit directory-style routes
  // (`en/index.html` rather than `en.html`) so URLs with a trailing slash —
  // including the root `/` → `/en/` redirect — resolve instead of 404ing.
  ...(basePath ? { basePath } : {}),
  ...(isStaticExport
    ? { output: 'export', trailingSlash: true, images: { unoptimized: true } }
    : {}),
};

export default withMDX(config);
