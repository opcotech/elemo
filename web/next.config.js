/** @type {import('next').NextConfig} */
const nextConfig = {
  experimental: {
    appDir: true,
  },
  images: {
    dangerouslyAllowSVG: true,
    domains: [
      'tailwindui.com',
      'github.com',
      'images.unsplash.com',
    ],
  }
}

module.exports = nextConfig
