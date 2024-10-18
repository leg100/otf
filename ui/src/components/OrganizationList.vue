<script setup lang="ts">
import { useMyFetch } from '../services/api'

const { data } = useMyFetch('/organizations').json()
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
