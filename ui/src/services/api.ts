import { ref } from 'vue'
import { useFetch, createFetch } from '@vueuse/core'
import { useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'

let fetcher: typeof useFetch

export function useMyFetch(url: string, opts?: any) {
    if (!fetcher) {
        const store = useUserStore()
        const router = useRouter()

        fetcher = createFetch({
            baseUrl: '/otfapi/v2',
            options: {
                beforeFetch({ options }) {
                    options.headers = {
                        ...options.headers,
                        'Content-Type': 'application/json',
                    }
                    return { options }
                },
                onFetchError(ctx) {
                    if (ctx.response?.status == 401) {
                        router.push('/login')
                        store.user = ref()
                    }
                    return ctx
                },
            },
        })
    }
    if (!opts) {
        return fetcher(url)
    }
    return fetcher(url, opts)
}
