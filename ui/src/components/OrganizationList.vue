<script setup lang="ts">
import { useGet } from '../services/openapi'
import BlockLink from '../components/BlockLink.vue'
import type { ParamsOption, RequestBodyOption } from "openapi-fetch";
import type { paths } from "#generated/api";

type GetOrganizationsResponse =
    paths["/organizations"]["get"]["responses"]["200"]["content"]["application/json"];

const { data } = useGet('/organizations')

</script>

<template>
    <div class="m-auto flex flex-col gap-2">
        <div v-for="{ name } in data" class="">
            <BlockLink :link="{ name: 'organization', params: { organization: name } }">
                <div>
                    <span>{{ name }}</span>
                </div>
            </BlockLink>
        </div>
    </div>
</template>
