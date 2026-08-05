import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/admin': 'http://localhost:8080',
      '/metrics': 'http://localhost:9090',
      '/health': 'http://localhost:8080',
      '/history': 'http://localhost:9090',
    },
  },
});
