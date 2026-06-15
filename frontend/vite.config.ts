import {defineConfig} from 'vite'
import { viteStaticCopy } from 'vite-plugin-static-copy'

import {svelte} from '@sveltejs/vite-plugin-svelte'

// https://vitejs.dev/config/
export default defineConfig({
  build: {
    target: 'esnext'
  },
  esbuild: {
    supported: {
      'destructuring': true
    }
  },
  optimizeDeps: {
    esbuildOptions: {
      supported: {
        'destructuring': true
      }
    }
  },
  plugins: [
    svelte(),
    viteStaticCopy({
      targets: [
        {
          src: 'src/help',
          dest: './'
        }
      ]
    })  
  ]
})
