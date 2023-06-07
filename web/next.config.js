/** @type {import('next').NextConfig} */
const nextConfig = {
  experimental: {
    serverActions: true
  },
  images: {
    dangerouslyAllowSVG: true,
    domains: ['tailwindui.com', 'github.com', 'images.unsplash.com', 'picsum.photos']
  }
};

module.exports = nextConfig;
