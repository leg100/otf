<script setup lang="ts">
import { computed } from 'vue'
import { useUserStore } from '../stores/user'

const store = useUserStore()
const username = computed(() => store.user?.username)

</script>

<template>
    <header class="bg-gray-200">
        <nav class="max-w-4xl p-1">
            <div>
                <img src="/logo.svg" alt="logo">
            </div>
            <div>
                <span v-if="username">{{ username }}</span>
            </div>
            <div>
                <RouterLink to="/profile">profile</RouterLink>
            </div>
            <div>
                <RouterLink to="/login">login</RouterLink>
            </div>
        </nav>
    </header>
</template>
