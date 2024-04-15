/** @type {import('next').NextConfig} */
const nextConfig = {
  images: {
    dangerouslyAllowSVG: true,
    remotePatterns: [
      { protocol: 'https', hostname: 'tailwindui.com', port: '' },
      { protocol: 'https', hostname: 'github.com', port: '' },
      { protocol: 'https', hostname: 'images.unsplash.com', port: '' },
      { protocol: 'https', hostname: 'picsum.photos', port: '' }
    ]
  }
};

module.exports = nextConfig;
