import { fileURLToPath, URL } from 'node:url'
import { readFileSync } from 'node:fs'
import AutoImport from 'unplugin-auto-import/vite'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vitejs.dev/config/
export default defineConfig({
    plugins: [
        vue(),
        AutoImport({
            imports: ['vue'],
            eslintrc: {
                enabled: true,
            },
        }),
    ],
    resolve: {
        alias: {
            '@': fileURLToPath(new URL('./src', import.meta.url))
        }
    },
    server: {
        https: {
            key: readFileSync('../internal/integration/fixtures/key.pem'),
            cert: readFileSync('../internal/integration/fixtures/cert.pem'),
        },
        proxy: {
            '/otfapi/v2': {
                target: 'https://localhost:8080',
                changeOrigin: true,
                secure: false,
                ws: true,
            },
            '/oauth': {
                target: 'https://localhost:8080',
                changeOrigin: true,
                secure: false,
                ws: true,
            },
        }
    }
})
