import react from '@vitejs/plugin-react';
import { configDefaults, defineConfig } from 'vitest/config';

const apiProxyTarget = process.env.VITE_API_PROXY_TARGET ?? 'http://127.0.0.1:8080';

export default defineConfig({
  plugins: [react()],
  server: {
    host: '0.0.0.0',
    port: 5188,
    strictPort: true,
    proxy: {
      '/api': apiProxyTarget,
      '/health': apiProxyTarget,
    },
  },
  test: {
    environment: 'jsdom',
    exclude: [...configDefaults.exclude, 'e2e/**'],
    setupFiles: './src/test/setupTests.ts',
    css: true,
  },
});
