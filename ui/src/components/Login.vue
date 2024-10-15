<script setup lang="ts">
import { useMyFetch } from '../services/api'
import { useUserStore } from '../stores/user'
import { useRouter } from 'vue-router'

const router = useRouter()

const { data } = useMyFetch('/login/clients').json()

const store = useUserStore()

// if user is logged whilst this page is mounted then navigate to /profile
watch(store, (updatedStore) => {
    console.log("watch updated")
    if (updatedStore.user) {
        router.push('/profile')
    }
}, { immediate: true })
</script>

<template>
    <div class="m-auto flex flex-col gap-2">
        <template v-if="data">
            <div v-for="{ name, icon, request_path } in data" class="">
                <a :href=request_path class="p-4 border border-black flex justify-center items-center gap-1">
                    <img :src="`data:image/png;base64,${icon}`" width="30" height="30" />
                    <span>Login with <span class="capitalize">{{ name }}</span></span>
                </a>
            </div>
        </template>
        <div v-else>No identity providers configured.</div>
    </div>
</template>
