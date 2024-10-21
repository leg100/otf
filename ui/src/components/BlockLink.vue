<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useTemplateRef, onMounted } from 'vue'

const props = defineProps(['link'])
const router = useRouter()
const myRef = useTemplateRef('my-ref')

function linkIfNothingSelected() {
    const isTextSelected = window.getSelection().toString();
    if (!isTextSelected) {
        router.push(props.link);
    }
}

onMounted(() => {
    const links = myRef.value.querySelectorAll("a");
    links.forEach(link => {
        link.addEventListener("click", (e) => e.stopPropagation());
    });
})
</script>

<template>
    <div ref="my-ref" class="widget cursor-pointer hover:bg-gray-100" @click="linkIfNothingSelected">
        <slot />
    </div>
</template>
