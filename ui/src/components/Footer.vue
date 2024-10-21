<script setup lang="ts">
import { useUserStore } from '../stores/user'

const store = useUserStore()
</script>
<template>
    <footer class="py-0.5 border-t-2 border-gray-200 text-xs font-mono">
        <div class="flex justify-between gap-4 max-w-2xl py-0.5 my-0 mx-auto">
            <div class="footer-version">
                <a href="https://github.com/leg100/otf">v1.2.3</a>
            </div>
            <div class="footer-site-login">
                <RouterLink v-if=!store.user to="/admin/login">site admin</RouterLink>
            </div>
        </div>
    </footer>
</template>
