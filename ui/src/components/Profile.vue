<script setup lang="ts">
import { computed } from 'vue'
import { useUserStore } from '../stores/user'
import { useMyFetch } from '../services/api'
import { useRouter } from 'vue-router'

const store = useUserStore()

const username = computed(() => store.user?.username)
const router = useRouter()

// logout sends a POST request to /otfapi/v2/logout. If it errors, then do nothing. If it succeeds, then unset current user and send user to /login.
async function logout(e: Event) {
    e.preventDefault()
    await useMyFetch('/logout', { method: "POST" })
    store.user = ref()
    router.push('/login')
}

</script>

<template>
    <div v-if="username" class="m-auto flex flex-col gap-2">
        <p>You are logged in as <span class="bg-gray-200">{{ username }}</span></p>
        <br>
        <form>
            <button @click="logout" class="btn btn-blue" id="logout">logout</button>
        </form>
    </div>
    <div v-else>
        You are currently logged out.
    </div>
</template>
