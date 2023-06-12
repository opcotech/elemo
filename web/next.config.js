/** @type {import('next').NextConfig} */
const nextConfig = {
  images: {
    dangerouslyAllowSVG: true,
    domains: ['tailwindui.com', 'github.com', 'images.unsplash.com', 'picsum.photos']
  }
};

module.exports = nextConfig;
