<script setup lang="ts">
import { useMyFetch } from '../services/api'
import BlockLink from '../components/BlockLink.vue'
import { useRoute } from 'vue-router'

const route = useRoute()

const { data } = useMyFetch(`/organizations/${route.params.organization}/workspaces`).json()

</script>
<template>
    <div class="m-auto flex flex-col gap-2">
        <div v-for="{ id, name } in data" class="">
            <BlockLink :link="{ name: 'workspaces', params: { workspaceId: id } }">
                <div>
                    <span>{{ name }}</span>
                </div>
            </BlockLink>
        </div>
    </div>
</template>
