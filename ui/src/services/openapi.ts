import { ref } from 'vue'
import { useFetch, createFetch } from '@vueuse/core'
import { useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'
import type { ParamsOption, RequestBodyOption, PathsWithMethod } from "openapi-fetch";
import createClient from "openapi-fetch";
import type { paths } from "#generated/api";

const client = createClient<paths>({});

type QueryOptions<Type> = ParamsOption<Type> & RequestBodyOption<Type>;

export function useGet<T>(path: PathsWithMethod, opts?: QueryOptions<T>) {
    const store = useUserStore()
    const router = useRouter()
    const state = ref<CatFactResponse>();
    const isReady = ref(false);
    const isFetching = ref(false);
    const error = ref<AppError | undefined>(undefined);

    async function execute() {
        error.value = undefined;
        isReady.value = false;
        isFetching.value = true;

        const { data, error: fetchError } = await client.GET(path, opts);

        if (fetchError) {
            error.value = fetchError;
        } else {
            state.value = data;
            isReady.value = true;
        }

        isFetching.value = false;
    }

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
return fetcher(url, opts)
}
